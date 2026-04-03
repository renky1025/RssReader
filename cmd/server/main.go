package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"rssreader/internal/api"
	"rssreader/internal/auth"
	"rssreader/internal/config"
	"rssreader/internal/fetcher"
	"rssreader/internal/store"
)

func main() {
	cfg := config.Load()

	// Initialize database
	db, err := store.New(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Run migrations
	migrationSQL, err := os.ReadFile("migrations/001_init.sql")
	if err != nil {
		log.Fatalf("Failed to read migration file: %v", err)
	}
	if err := db.MigrateFromString(string(migrationSQL)); err != nil {
		log.Printf("Migration warning (may already exist): %v", err)
	}

	// Run user management migration
	migration2SQL, err := os.ReadFile("migrations/002_user_management.sql")
	if err != nil {
		log.Printf("Migration 002 not found: %v", err)
	} else {
		if err := db.MigrateFromString(string(migration2SQL)); err != nil {
			log.Printf("Migration 002 warning (may already exist): %v", err)
		}
	}

	// Ensure default admin user exists
	adminUser, _ := db.GetUserByUsername("admin")
	if adminUser == nil {
		hash, _ := auth.HashPassword("admin")
		_, err := db.CreateUser("admin", "admin@localhost", hash, true)
		if err != nil {
			log.Printf("Failed to create admin user: %v", err)
		} else {
			log.Println("Created default admin user (admin/admin)")
		}
	}

	// Initialize fetcher
	f := fetcher.New(db, cfg.FetchConcurrency)

	intervalMinutes := cfg.FetchInterval
	if raw, ok, err := db.GetAppSetting("fetch_interval_minutes"); err == nil && ok {
		if v, convErr := strconv.Atoi(raw); convErr == nil && v > 0 {
			intervalMinutes = v
		}
	}

	// Start background fetch scheduler
	f.StartScheduler(time.Duration(intervalMinutes) * time.Minute)

	// Initialize and start server
	server := api.NewServer(cfg, db, f)

	log.Printf("Starting server on :%s", cfg.ServerPort)
	if err := http.ListenAndServe(":"+cfg.ServerPort, server); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
