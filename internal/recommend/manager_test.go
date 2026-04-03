package recommend

import (
	"os"
	"path/filepath"
	"testing"

	"rssreader/internal/models"
)

func TestManagerCRUD(t *testing.T) {
	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "recommended_feeds.json")

	mgr := NewManager(filePath)

	created, err := mgr.Create(models.RecommendedFeed{
		Name:        "HN",
		URL:         "https://news.ycombinator.com/rss",
		Description: "Hacker News",
		Category:    "Tech",
		Icon:        "https://news.ycombinator.com/favicon.ico",
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}
	if created.ID == "" {
		t.Fatalf("expected generated id")
	}

	feeds, err := mgr.List()
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(feeds) != 1 {
		t.Fatalf("expected 1 feed, got %d", len(feeds))
	}

	updated, err := mgr.Update(created.ID, models.RecommendedFeed{
		Name:        "HN Updated",
		URL:         created.URL,
		Description: created.Description,
		Category:    created.Category,
		Icon:        created.Icon,
	})
	if err != nil {
		t.Fatalf("update failed: %v", err)
	}
	if updated.Name != "HN Updated" {
		t.Fatalf("unexpected updated name: %s", updated.Name)
	}

	if err := mgr.Delete(created.ID); err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	feeds, err = mgr.List()
	if err != nil {
		t.Fatalf("list after delete failed: %v", err)
	}
	if len(feeds) != 0 {
		t.Fatalf("expected empty feed list after delete, got %d", len(feeds))
	}

	if _, err := os.Stat(filePath); err != nil {
		t.Fatalf("expected data file to exist: %v", err)
	}
}
