package store

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"rssreader/internal/models"
)

func (db *DB) GetAdminFeeds(params models.AdminFeedListParams) ([]*models.Feed, int, error) {
	conditions := []string{}
	args := []interface{}{}

	if params.UserID != nil {
		conditions = append(conditions, "f.user_id = ?")
		args = append(args, *params.UserID)
	}
	if params.Disabled != nil {
		conditions = append(conditions, "f.disabled = ?")
		args = append(args, *params.Disabled)
	}
	if strings.TrimSpace(params.Query) != "" {
		conditions = append(conditions, "(f.title LIKE ? OR f.url LIKE ? OR u.username LIKE ?)")
		q := "%" + strings.TrimSpace(params.Query) + "%"
		args = append(args, q, q, q)
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	var total int
	countSQL := fmt.Sprintf(`
		SELECT COUNT(*)
		FROM feeds f
		JOIN users u ON u.id = f.user_id
		%s
	`, where)
	if err := db.QueryRow(countSQL, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	if params.Limit <= 0 {
		params.Limit = 50
	}
	querySQL := fmt.Sprintf(`
		SELECT f.id, f.user_id, f.folder_id, f.url, f.title, f.site_url, f.description,
			   f.last_fetched, f.error_count, f.last_error, f.disabled, f.created_at, u.username
		FROM feeds f
		JOIN users u ON u.id = f.user_id
		%s
		ORDER BY f.created_at DESC, f.id DESC
		LIMIT ? OFFSET ?
	`, where)
	queryArgs := append(args, params.Limit, params.Offset)

	rows, err := db.Query(querySQL, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	feeds := make([]*models.Feed, 0)
	for rows.Next() {
		var f models.Feed
		var createdAt int64
		var folderID, lastFetched sql.NullInt64
		var title, siteURL, description, lastError sql.NullString
		if err := rows.Scan(
			&f.ID, &f.UserID, &folderID, &f.URL, &title, &siteURL, &description,
			&lastFetched, &f.ErrorCount, &lastError, &f.Disabled, &createdAt, &f.Username,
		); err != nil {
			return nil, 0, err
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

	return feeds, total, nil
}

func (db *DB) AdminUpdateFeed(id int64, req models.AdminUpdateFeedRequest) error {
	updates := []string{}
	args := []interface{}{}

	if req.UserID != nil {
		updates = append(updates, "user_id = ?")
		args = append(args, *req.UserID)
	}
	if req.FolderID != nil {
		updates = append(updates, "folder_id = ?")
		args = append(args, *req.FolderID)
	}
	if req.URL != nil {
		updates = append(updates, "url = ?")
		args = append(args, *req.URL)
	}
	if req.Title != nil {
		updates = append(updates, "title = ?")
		args = append(args, *req.Title)
	}
	if req.SiteURL != nil {
		updates = append(updates, "site_url = ?")
		args = append(args, *req.SiteURL)
	}
	if req.Description != nil {
		updates = append(updates, "description = ?")
		args = append(args, *req.Description)
	}
	if req.Disabled != nil {
		updates = append(updates, "disabled = ?")
		args = append(args, *req.Disabled)
	}
	if req.ErrorCount != nil {
		updates = append(updates, "error_count = ?")
		args = append(args, *req.ErrorCount)
	}
	if req.LastError != nil {
		updates = append(updates, "last_error = ?")
		args = append(args, *req.LastError)
	}

	if len(updates) == 0 {
		return nil
	}

	args = append(args, id)
	sql := fmt.Sprintf("UPDATE feeds SET %s WHERE id = ?", strings.Join(updates, ", "))
	_, err := db.Exec(sql, args...)
	return err
}

func (db *DB) AdminCreateFeed(req models.AdminCreateFeedRequest) (*models.Feed, error) {
	result, err := db.Exec(`
		INSERT INTO feeds (user_id, folder_id, url, title, site_url, description)
		VALUES (?, ?, ?, ?, ?, ?)
	`, req.UserID, req.FolderID, req.URL, req.Title, req.SiteURL, req.Description)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return db.GetFeedByID(id)
}

func (db *DB) AdminDeleteFeed(id int64) error {
	_, err := db.Exec(`DELETE FROM feeds WHERE id = ?`, id)
	return err
}
