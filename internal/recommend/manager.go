package recommend

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"rssreader/internal/models"
)

type Manager struct {
	filePath string
	mu       sync.RWMutex
}

func NewManager(filePath string) *Manager {
	return &Manager{filePath: filePath}
}

func (m *Manager) List() ([]models.RecommendedFeed, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.readAll()
}

func (m *Manager) Create(feed models.RecommendedFeed) (*models.RecommendedFeed, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if strings.TrimSpace(feed.Name) == "" || strings.TrimSpace(feed.URL) == "" {
		return nil, errors.New("name and url are required")
	}

	feeds, err := m.readAll()
	if err != nil {
		return nil, err
	}

	for _, item := range feeds {
		if strings.EqualFold(item.URL, feed.URL) {
			return nil, errors.New("feed url already exists")
		}
	}

	feed.ID = generateFeedID()
	feeds = append(feeds, feed)
	if err := m.writeAll(feeds); err != nil {
		return nil, err
	}
	return &feed, nil
}

func (m *Manager) Update(id string, feed models.RecommendedFeed) (*models.RecommendedFeed, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if strings.TrimSpace(id) == "" {
		return nil, errors.New("id is required")
	}
	if strings.TrimSpace(feed.Name) == "" || strings.TrimSpace(feed.URL) == "" {
		return nil, errors.New("name and url are required")
	}

	feeds, err := m.readAll()
	if err != nil {
		return nil, err
	}

	found := false
	for i := range feeds {
		if feeds[i].ID != id {
			continue
		}
		found = true
		feeds[i].Name = feed.Name
		feeds[i].URL = feed.URL
		feeds[i].Description = feed.Description
		feeds[i].Category = feed.Category
		feeds[i].Icon = feed.Icon
		updated := feeds[i]
		if err := m.writeAll(feeds); err != nil {
			return nil, err
		}
		return &updated, nil
	}

	if !found {
		return nil, errors.New("recommended feed not found")
	}
	return nil, errors.New("unexpected state")
}

func (m *Manager) Delete(id string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	feeds, err := m.readAll()
	if err != nil {
		return err
	}

	filtered := make([]models.RecommendedFeed, 0, len(feeds))
	found := false
	for _, item := range feeds {
		if item.ID == id {
			found = true
			continue
		}
		filtered = append(filtered, item)
	}

	if !found {
		return errors.New("recommended feed not found")
	}
	return m.writeAll(filtered)
}

func (m *Manager) readAll() ([]models.RecommendedFeed, error) {
	data, err := os.ReadFile(m.filePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return []models.RecommendedFeed{}, nil
		}
		return nil, err
	}

	if len(data) == 0 {
		return []models.RecommendedFeed{}, nil
	}

	var feeds []models.RecommendedFeed
	if err := json.Unmarshal(data, &feeds); err != nil {
		return nil, err
	}
	return feeds, nil
}

func (m *Manager) writeAll(feeds []models.RecommendedFeed) error {
	dir := filepath.Dir(m.filePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(feeds, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(m.filePath, data, 0644)
}

func generateFeedID() string {
	return "feed_" + strconv.FormatInt(time.Now().UnixNano(), 10)
}
