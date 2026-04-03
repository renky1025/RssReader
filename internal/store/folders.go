package store

import (
	"database/sql"
	"time"

	"rssreader/internal/models"
)

func (db *DB) GetFolders(userID int64) ([]*models.Folder, error) {
	rows, err := db.Query(`
		SELECT f.id, f.user_id, f.name, f.parent_id, f.created_at,
			   (SELECT COUNT(*) FROM feeds WHERE folder_id = f.id) as feed_count
		FROM folders f
		WHERE f.user_id = ?
		ORDER BY f.name
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var folders []*models.Folder
	for rows.Next() {
		var f models.Folder
		var createdAt int64
		err := rows.Scan(&f.ID, &f.UserID, &f.Name, &f.ParentID, &createdAt, &f.FeedCount)
		if err != nil {
			return nil, err
		}
		f.CreatedAt = time.Unix(createdAt, 0)
		folders = append(folders, &f)
	}
	return folders, nil
}

func (db *DB) GetFolderByID(id int64) (*models.Folder, error) {
	var f models.Folder
	var createdAt int64
	err := db.QueryRow(`
		SELECT id, user_id, name, parent_id, created_at
		FROM folders WHERE id = ?
	`, id).Scan(&f.ID, &f.UserID, &f.Name, &f.ParentID, &createdAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	f.CreatedAt = time.Unix(createdAt, 0)
	return &f, nil
}

func (db *DB) CreateFolder(userID int64, name string, parentID *int64) (*models.Folder, error) {
	result, err := db.Exec(`
		INSERT INTO folders (user_id, name, parent_id) VALUES (?, ?, ?)
	`, userID, name, parentID)
	if err != nil {
		return nil, err
	}
	id, _ := result.LastInsertId()
	return db.GetFolderByID(id)
}

func (db *DB) UpdateFolder(id int64, name string) error {
	_, err := db.Exec(`UPDATE folders SET name = ? WHERE id = ?`, name, id)
	return err
}

func (db *DB) DeleteFolder(id int64) error {
	// Move feeds to no folder
	_, err := db.Exec(`UPDATE feeds SET folder_id = NULL WHERE folder_id = ?`, id)
	if err != nil {
		return err
	}
	_, err = db.Exec(`DELETE FROM folders WHERE id = ?`, id)
	return err
}
