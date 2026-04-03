package store

import (
	"path/filepath"
	"testing"

	"rssreader/internal/models"
)

func setupAdminFeedTestDB(t *testing.T) *DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "test.db")
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("failed to init db: %v", err)
	}

	_, err = db.Exec(`
		CREATE TABLE users (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			username TEXT NOT NULL UNIQUE,
			password_hash TEXT NOT NULL DEFAULT '',
			is_admin INTEGER DEFAULT 0,
			status INTEGER DEFAULT 1,
			created_at INTEGER DEFAULT (strftime('%s','now'))
		);

		CREATE TABLE feeds (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			user_id INTEGER NOT NULL,
			folder_id INTEGER,
			url TEXT NOT NULL,
			title TEXT,
			site_url TEXT,
			description TEXT,
			last_fetched INTEGER,
			etag TEXT,
			last_modified TEXT,
			error_count INTEGER DEFAULT 0,
			last_error TEXT,
			disabled INTEGER DEFAULT 0,
			created_at INTEGER DEFAULT (strftime('%s','now'))
		);
	`)
	if err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	_, err = db.Exec(`
		INSERT INTO users (id, username, password_hash, is_admin, status) VALUES
		(1, 'admin', 'x', 1, 1),
		(2, 'alice', 'x', 0, 1);

		INSERT INTO feeds (id, user_id, url, title, site_url, description, disabled, error_count) VALUES
		(1, 2, 'https://example.com/rss', 'Example', 'https://example.com', 'desc', 0, 0),
		(2, 2, 'https://example.com/alt', 'Alt', 'https://example.com', 'alt desc', 1, 2);
	`)
	if err != nil {
		t.Fatalf("failed to seed data: %v", err)
	}

	return db
}

func TestAdminFeedListAndUpdate(t *testing.T) {
	db := setupAdminFeedTestDB(t)
	defer db.Close()

	feeds, total, err := db.GetAdminFeeds(models.AdminFeedListParams{
		Query: "example",
		Limit: 10,
	})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if total != 2 {
		t.Fatalf("expected total=2, got %d", total)
	}
	if len(feeds) != 2 {
		t.Fatalf("expected 2 items, got %d", len(feeds))
	}

	title := "Example Updated"
	disabled := true
	errCount := 9
	req := models.AdminUpdateFeedRequest{
		Title:      &title,
		Disabled:   &disabled,
		ErrorCount: &errCount,
	}

	if err := db.AdminUpdateFeed(1, req); err != nil {
		t.Fatalf("update failed: %v", err)
	}

	updated, err := db.GetFeedByID(1)
	if err != nil {
		t.Fatalf("get feed failed: %v", err)
	}
	if updated.Title != title {
		t.Fatalf("expected title=%q, got %q", title, updated.Title)
	}
	if !updated.Disabled {
		t.Fatalf("expected feed disabled")
	}
	if updated.ErrorCount != errCount {
		t.Fatalf("expected error_count=%d, got %d", errCount, updated.ErrorCount)
	}
}
