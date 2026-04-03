package store

import (
	"database/sql"
	"strings"
	"time"

	"rssreader/internal/models"
)

func (db *DB) GetFeeds(userID int64) ([]*models.Feed, error) {
	rows, err := db.Query(`
		SELECT f.id, f.user_id, f.folder_id, f.url, f.title, f.site_url, f.description,
			   f.last_fetched, f.error_count, f.last_error, f.disabled, f.created_at,
			   (SELECT COUNT(*) FROM articles a WHERE a.feed_id = f.id AND a.is_read = 0) as unread_count
		FROM feeds f
		WHERE f.user_id = ?
		ORDER BY f.title
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feeds []*models.Feed
	for rows.Next() {
		var f models.Feed
		var createdAt int64
		var folderID, lastFetched sql.NullInt64
		var title, siteURL, description, lastError sql.NullString
		err := rows.Scan(&f.ID, &f.UserID, &folderID, &f.URL, &title, &siteURL,
			&description, &lastFetched, &f.ErrorCount, &lastError, &f.Disabled,
			&createdAt, &f.UnreadCount)
		if err != nil {
			return nil, err
		}
		if folderID.Valid {
			f.FolderID = &folderID.Int64
		}
		if lastFetched.Valid {
			f.LastFetched = &lastFetched.Int64
		}
		f.Title = title.String
		f.SiteURL = siteURL.String
		f.Description = description.String
		f.LastError = lastError.String
		f.CreatedAt = time.Unix(createdAt, 0)
		feeds = append(feeds, &f)
	}
	return feeds, nil
}

func (db *DB) GetFeedByID(id int64) (*models.Feed, error) {
	var f models.Feed
	var createdAt int64
	var folderID, lastFetched sql.NullInt64
	var title, siteURL, description, etag, lastModified, lastError sql.NullString
	err := db.QueryRow(`
		SELECT id, user_id, folder_id, url, title, site_url, description,
			   last_fetched, etag, last_modified, error_count, last_error, disabled, created_at
		FROM feeds WHERE id = ?
	`, id).Scan(&f.ID, &f.UserID, &folderID, &f.URL, &title, &siteURL,
		&description, &lastFetched, &etag, &lastModified, &f.ErrorCount,
		&lastError, &f.Disabled, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if folderID.Valid {
		f.FolderID = &folderID.Int64
	}
	if lastFetched.Valid {
		f.LastFetched = &lastFetched.Int64
	}
	f.Title = title.String
	f.SiteURL = siteURL.String
	f.Description = description.String
	f.ETag = etag.String
	f.LastModified = lastModified.String
	f.LastError = lastError.String
	f.CreatedAt = time.Unix(createdAt, 0)
	return &f, nil
}

func (db *DB) GetFeedByURL(userID int64, url string) (*models.Feed, error) {
	var f models.Feed
	var createdAt int64
	var folderID, lastFetched sql.NullInt64
	var title, siteURL, description, etag, lastModified, lastError sql.NullString
	err := db.QueryRow(`
		SELECT id, user_id, folder_id, url, title, site_url, description,
			   last_fetched, etag, last_modified, error_count, last_error, disabled, created_at
		FROM feeds WHERE user_id = ? AND url = ?
	`, userID, url).Scan(&f.ID, &f.UserID, &folderID, &f.URL, &title, &siteURL,
		&description, &lastFetched, &etag, &lastModified, &f.ErrorCount,
		&lastError, &f.Disabled, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if folderID.Valid {
		f.FolderID = &folderID.Int64
	}
	if lastFetched.Valid {
		f.LastFetched = &lastFetched.Int64
	}
	f.Title = title.String
	f.SiteURL = siteURL.String
	f.Description = description.String
	f.ETag = etag.String
	f.LastModified = lastModified.String
	f.LastError = lastError.String
	f.CreatedAt = time.Unix(createdAt, 0)
	return &f, nil
}

func (db *DB) CreateFeed(userID int64, url, title, siteURL, description string, folderID *int64) (*models.Feed, error) {
	result, err := db.Exec(`
		INSERT INTO feeds (user_id, folder_id, url, title, site_url, description)
		VALUES (?, ?, ?, ?, ?, ?)
	`, userID, folderID, url, title, siteURL, description)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return db.GetFeedByID(id)
}

func (db *DB) UpdateFeed(id int64, req models.UpdateFeedRequest) error {
	// Build dynamic query based on what fields are provided
	setParts := []string{}
	args := []interface{}{}
	
	if req.Title != nil {
		setParts = append(setParts, "title = ?")
		args = append(args, *req.Title)
	}
	if req.FolderID != nil {
		setParts = append(setParts, "folder_id = ?")
		args = append(args, *req.FolderID)
	}
	
	if len(setParts) == 0 {
		return nil // Nothing to update
	}
	
	query := "UPDATE feeds SET " + strings.Join(setParts, ", ") + " WHERE id = ?"
	args = append(args, id)
	
	_, err := db.Exec(query, args...)
	return err
}

func (db *DB) UpdateFeedMetadata(id int64, title, siteURL, description string) error {
	_, err := db.Exec(`
		UPDATE feeds SET title = ?, site_url = ?, description = ? WHERE id = ?
	`, title, siteURL, description, id)
	return err
}

func (db *DB) UpdateFeedFolder(id int64, folderID *int64) error {
	_, err := db.Exec(`UPDATE feeds SET folder_id = ? WHERE id = ?`, folderID, id)
	return err
}

func (db *DB) UpdateFeedFetchStatus(id int64, etag, lastModified string, errorCount int, lastError string) error {
	_, err := db.Exec(`
		UPDATE feeds SET 
			last_fetched = strftime('%s','now'),
			etag = ?,
			last_modified = ?,
			error_count = ?,
			last_error = ?
		WHERE id = ?
	`, etag, lastModified, errorCount, lastError, id)
	return err
}

func (db *DB) DeleteFeed(id int64) error {
	_, err := db.Exec(`DELETE FROM feeds WHERE id = ?`, id)
	return err
}

func (db *DB) GetAllFeedsForFetch() ([]*models.Feed, error) {
	rows, err := db.Query(`
		SELECT id, user_id, folder_id, url, title, site_url, description,
			   last_fetched, etag, last_modified, error_count, last_error, disabled, created_at
		FROM feeds
		WHERE disabled = 0
		ORDER BY last_fetched ASC NULLS FIRST
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var feeds []*models.Feed
	for rows.Next() {
		var f models.Feed
		var createdAt int64
		var folderID, lastFetched sql.NullInt64
		var title, siteURL, description, etag, lastModified, lastError sql.NullString
		err := rows.Scan(&f.ID, &f.UserID, &folderID, &f.URL, &title, &siteURL,
			&description, &lastFetched, &etag, &lastModified, &f.ErrorCount,
			&lastError, &f.Disabled, &createdAt)
		if err != nil {
			return nil, err
		}
		if folderID.Valid {
			f.FolderID = &folderID.Int64
		}
		if lastFetched.Valid {
			f.LastFetched = &lastFetched.Int64
		}
		f.Title = title.String
		f.SiteURL = siteURL.String
		f.Description = description.String
		f.ETag = etag.String
		f.LastModified = lastModified.String
		f.LastError = lastError.String
		f.CreatedAt = time.Unix(createdAt, 0)
		feeds = append(feeds, &f)
	}
	return feeds, nil
}
