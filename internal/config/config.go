package config

import (
	"os"
	"strconv"
)

type Config struct {
	DatabasePath     string
	JWTSecret        string
	ServerPort       string
	FetchInterval    int // minutes
	FetchConcurrency int
}

func Load() *Config {
	return &Config{
		DatabasePath:     getEnv("DATABASE_PATH", "./data/rssreader.db"),
		JWTSecret:        getEnv("JWT_SECRET", "change-me-in-production"),
		ServerPort:       getEnv("SERVER_PORT", "8080"),
		FetchInterval:    getEnvInt("FETCH_INTERVAL", 15),
		FetchConcurrency: getEnvInt("FETCH_CONCURRENCY", 5),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if i, err := strconv.Atoi(value); err == nil {
			return i
		}
	}
	return defaultValue
}
