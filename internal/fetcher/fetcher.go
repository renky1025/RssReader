package fetcher

import (
	"context"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/mmcdole/gofeed"

	"rssreader/internal/models"
	"rssreader/internal/store"
)

// cleanContent removes CDATA artifacts and fixes relative URLs
func cleanContent(content string, baseURL string) string {
	// Remove CDATA markers that might have leaked through
	content = strings.ReplaceAll(content, "<![CDATA[", "")
	content = strings.ReplaceAll(content, "]]>", "")

	// Parse base URL for resolving relative URLs
	base, err := url.Parse(baseURL)
	if err != nil {
		return content
	}

	// Fix relative image URLs
	imgRegex := regexp.MustCompile(`<img[^>]+src=["']([^"']+)["']`)
	content = imgRegex.ReplaceAllStringFunc(content, func(match string) string {
		srcMatch := regexp.MustCompile(`src=["']([^"']+)["']`).FindStringSubmatch(match)
		if len(srcMatch) < 2 {
			return match
		}
		src := srcMatch[1]

		// If already absolute, skip
		if strings.HasPrefix(src, "http://") || strings.HasPrefix(src, "https://") || strings.HasPrefix(src, "//") {
			return match
		}

		// Resolve relative URL
		resolved, err := base.Parse(src)
		if err != nil {
			return match
		}

		return strings.Replace(match, srcMatch[1], resolved.String(), 1)
	})

	// Fix relative link URLs
	linkRegex := regexp.MustCompile(`<a[^>]+href=["']([^"']+)["']`)
	content = linkRegex.ReplaceAllStringFunc(content, func(match string) string {
		hrefMatch := regexp.MustCompile(`href=["']([^"']+)["']`).FindStringSubmatch(match)
		if len(hrefMatch) < 2 {
			return match
		}
		href := hrefMatch[1]

		// Skip anchors, mailto, javascript, and absolute URLs
		if strings.HasPrefix(href, "#") || strings.HasPrefix(href, "mailto:") ||
			strings.HasPrefix(href, "javascript:") || strings.HasPrefix(href, "http://") ||
			strings.HasPrefix(href, "https://") || strings.HasPrefix(href, "//") {
			return match
		}

		resolved, err := base.Parse(href)
		if err != nil {
			return match
		}

		return strings.Replace(match, hrefMatch[1], resolved.String(), 1)
	})

	return content
}

type Fetcher struct {
	db          *store.DB
	parser      *gofeed.Parser
	client      *http.Client
	concurrency int
	schedulerMu sync.Mutex
	ticker      *time.Ticker
	interval    time.Duration
	stopCh      chan struct{}
	started     bool
}

func New(db *store.DB, concurrency int) *Fetcher {
	return &Fetcher{
		db:     db,
		parser: gofeed.NewParser(),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		concurrency: concurrency,
	}
}

func (f *Fetcher) FetchFeed(feed *models.Feed) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", feed.URL, nil)
	if err != nil {
		return err
	}

	req.Header.Set("User-Agent", "RSSReader/1.0")
	if feed.ETag != "" {
		req.Header.Set("If-None-Match", feed.ETag)
	}
	if feed.LastModified != "" {
		req.Header.Set("If-Modified-Since", feed.LastModified)
	}

	resp, err := f.client.Do(req)
	if err != nil {
		f.db.UpdateFeedFetchStatus(feed.ID, feed.ETag, feed.LastModified, feed.ErrorCount+1, err.Error())
		return err
	}
	defer resp.Body.Close()

	// Not modified
	if resp.StatusCode == http.StatusNotModified {
		f.db.UpdateFeedFetchStatus(feed.ID, feed.ETag, feed.LastModified, 0, "")
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		errMsg := "HTTP " + resp.Status
		f.db.UpdateFeedFetchStatus(feed.ID, feed.ETag, feed.LastModified, feed.ErrorCount+1, errMsg)
		return nil
	}

	parsedFeed, err := f.parser.Parse(resp.Body)
	if err != nil {
		f.db.UpdateFeedFetchStatus(feed.ID, feed.ETag, feed.LastModified, feed.ErrorCount+1, err.Error())
		return err
	}

	// Update feed metadata if changed
	if parsedFeed.Title != "" && parsedFeed.Title != feed.Title {
		f.db.UpdateFeedMetadata(feed.ID, parsedFeed.Title, parsedFeed.Link, parsedFeed.Description)
	}

	// Save articles
	for _, item := range parsedFeed.Items {
		guid := item.GUID
		if guid == "" {
			guid = item.Link
		}

		var publishedAt int64
		if item.PublishedParsed != nil {
			publishedAt = item.PublishedParsed.Unix()
		} else if item.UpdatedParsed != nil {
			publishedAt = item.UpdatedParsed.Unix()
		} else {
			publishedAt = time.Now().Unix()
		}

		var imageURL string
		if item.Image != nil {
			imageURL = item.Image.URL
		} else if len(item.Enclosures) > 0 {
			for _, enc := range item.Enclosures {
				if strings.HasPrefix(enc.Type, "image/") {
					imageURL = enc.URL
					break
				}
			}
		}

		// Try to extract first image from content if no image found
		if imageURL == "" {
			contentToSearch := item.Content
			if contentToSearch == "" {
				contentToSearch = item.Description
			}
			imgMatch := regexp.MustCompile(`<img[^>]+src=["']([^"']+)["']`).FindStringSubmatch(contentToSearch)
			if len(imgMatch) >= 2 {
				imageURL = imgMatch[1]
				// Resolve relative URL
				if !strings.HasPrefix(imageURL, "http://") && !strings.HasPrefix(imageURL, "https://") && !strings.HasPrefix(imageURL, "//") {
					if base, err := url.Parse(item.Link); err == nil {
						if resolved, err := base.Parse(imageURL); err == nil {
							imageURL = resolved.String()
						}
					}
				}
			}
		}

		content := item.Content
		if content == "" {
			content = item.Description
		}

		// Clean content: remove CDATA artifacts and fix relative URLs
		baseURL := item.Link
		if baseURL == "" {
			baseURL = feed.SiteURL
		}
		if baseURL == "" {
			baseURL = feed.URL
		}
		content = cleanContent(content, baseURL)

		summary := item.Description
		// Clean summary too
		summary = cleanContent(summary, baseURL)
		// Remove HTML tags from summary for plain text display
		summary = regexp.MustCompile(`<[^>]*>`).ReplaceAllString(summary, "")
		summary = strings.TrimSpace(summary)
		if len(summary) > 500 {
			summary = summary[:500] + "..."
		}

		var authorName string
		if item.Author != nil {
			authorName = item.Author.Name
		}

		f.db.CreateArticle(
			feed.ID,
			guid,
			item.Link,
			item.Title,
			authorName,
			content,
			summary,
			imageURL,
			publishedAt,
		)
	}

	// Update fetch status
	etag := resp.Header.Get("ETag")
	lastModified := resp.Header.Get("Last-Modified")
	f.db.UpdateFeedFetchStatus(feed.ID, etag, lastModified, 0, "")

	return nil
}

func (f *Fetcher) FetchAll() {
	feeds, err := f.db.GetAllFeedsForFetch()
	if err != nil {
		log.Printf("Error getting feeds: %v", err)
		return
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, f.concurrency)

	for _, feed := range feeds {
		wg.Add(1)
		sem <- struct{}{}

		go func(feed *models.Feed) {
			defer wg.Done()
			defer func() { <-sem }()

			log.Printf("Fetching feed: %s", feed.URL)
			if err := f.FetchFeed(feed); err != nil {
				log.Printf("Error fetching %s: %v", feed.URL, err)
			}
		}(feed)
	}

	wg.Wait()
}

func (f *Fetcher) StartScheduler(interval time.Duration) {
	f.schedulerMu.Lock()
	initialFetch := !f.started
	f.started = true
	f.schedulerMu.Unlock()
	f.startScheduler(interval, initialFetch)
}

func (f *Fetcher) startScheduler(interval time.Duration, initialFetch bool) {
	f.schedulerMu.Lock()
	f.interval = interval
	if f.ticker != nil {
		f.ticker.Stop()
	}
	if f.stopCh != nil {
		close(f.stopCh)
	}
	f.ticker = time.NewTicker(interval)
	f.stopCh = make(chan struct{})
	ticker := f.ticker
	stopCh := f.stopCh
	f.schedulerMu.Unlock()

	go func() {
		if initialFetch {
			f.FetchAll()
		}

		for {
			select {
			case <-ticker.C:
				log.Println("Starting scheduled feed fetch...")
				f.FetchAll()
				log.Println("Scheduled feed fetch completed")
			case <-stopCh:
				return
			}
		}
	}()
}

func (f *Fetcher) SetSchedulerInterval(interval time.Duration) {
	f.startScheduler(interval, false)
}

func (f *Fetcher) SchedulerInterval() time.Duration {
	f.schedulerMu.Lock()
	defer f.schedulerMu.Unlock()
	return f.interval
}
