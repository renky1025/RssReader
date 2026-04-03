package store

import (
	"path/filepath"
	"testing"
)

func TestAppSettingCRUD(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "settings.db")
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("init db failed: %v", err)
	}
	defer db.Close()

	val, ok, err := db.GetAppSetting("fetch_interval_minutes")
	if err != nil {
		t.Fatalf("get empty setting failed: %v", err)
	}
	if ok || val != "" {
		t.Fatalf("expected missing setting, got ok=%v val=%q", ok, val)
	}

	if err := db.SetAppSetting("fetch_interval_minutes", "25"); err != nil {
		t.Fatalf("set setting failed: %v", err)
	}

	val, ok, err = db.GetAppSetting("fetch_interval_minutes")
	if err != nil {
		t.Fatalf("get setting failed: %v", err)
	}
	if !ok || val != "25" {
		t.Fatalf("expected value=25, got ok=%v val=%q", ok, val)
	}
}
