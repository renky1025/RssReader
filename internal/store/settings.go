package store

import "database/sql"

func (db *DB) ensureAppSettingsTable() error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS app_settings (
			key TEXT PRIMARY KEY,
			value TEXT NOT NULL,
			updated_at INTEGER DEFAULT (strftime('%s','now'))
		)
	`)
	return err
}

func (db *DB) GetAppSetting(key string) (string, bool, error) {
	if err := db.ensureAppSettingsTable(); err != nil {
		return "", false, err
	}

	var value string
	err := db.QueryRow(`SELECT value FROM app_settings WHERE key = ?`, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return value, true, nil
}

func (db *DB) SetAppSetting(key, value string) error {
	if err := db.ensureAppSettingsTable(); err != nil {
		return err
	}

	_, err := db.Exec(`
		INSERT INTO app_settings (key, value, updated_at)
		VALUES (?, ?, strftime('%s','now'))
		ON CONFLICT(key) DO UPDATE SET
			value = excluded.value,
			updated_at = excluded.updated_at
	`, key, value)
	return err
}
