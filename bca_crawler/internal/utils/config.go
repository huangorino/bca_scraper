package utils

import (
	"flag"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

// Config holds all runtime configuration
type Config struct {
	StartURL  string
	DetailURL string
	DBPath    string
	UserAgent string
	LogLevel  string
}

// Load reads .env and CLI flags, sets defaults, and returns a Config struct
func LoadCfg() (*Config, error) {
	// Load from .env (if exists)
	_ = godotenv.Load()

	cfg := &Config{}

	flag.StringVar(&cfg.StartURL, "start-url", os.Getenv("START_URL"), "Base Bursa announcements URL")
	flag.StringVar(&cfg.DetailURL, "detail-url", os.Getenv("DETAIL_URL"), "Announcement detail URL prefix")
	flag.StringVar(&cfg.DBPath, "db-path", os.Getenv("DB_PATH"), "SQLite DB file path")
	flag.StringVar(&cfg.UserAgent, "ua", os.Getenv("UA"), "Browser User-Agent")
	flag.StringVar(&cfg.LogLevel, "log-level", os.Getenv("LOG_LEVEL"), "Log level (debug, info, warn, error)")

	flag.Parse()

	// Validate
	if cfg.StartURL == "" || cfg.DetailURL == "" {
		return nil, fmt.Errorf("missing required URLs")
	}

	return cfg, nil
}
