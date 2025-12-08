package main

import (
	"strconv"
	"strings"
	"time"

	"bca_crawler/internal/db"
	"bca_crawler/internal/models"
	"bca_crawler/internal/services"
	"bca_crawler/internal/utils"
)

func main() {
	// Load configuration
	cfg, err := utils.LoadCfg()
	if err != nil {
		panic(err)
	}

	// Initialize logger with level from config
	utils.InitLogger()
	log := utils.Logger

	// Setup database
	database, err := db.Connect(cfg.DBPath, db.DriverType(cfg.DBDriver))
	if err != nil {
		log.Fatalf("[Error] Failed to setup DB: %v", err)
	}
	defer database.Close()

	// Load Bursa main page
	chromeCtx, cancel := services.InitCtx(cfg.UserAgent)
	defer cancel()

	data, err := db.FetchMissingAnnID(database)
	if err != nil {
		log.Fatalf("Failed to fetch missing announcement IDs: %v", err)
	}

	for _, id := range data {
		annID := strconv.FormatInt(id, 10)
		url := cfg.DetailDomain + cfg.DetailURL + annID
		log.Infof("Processing announcement ID: %s", annID)

		var html string
		var err error
		maxRetries := 3
		retryTriggers := []string{
			"Web server is returning an unknown error",
			"SSL handshake failed",
			"A timeout occurred",
			"Just a moment",
			"Connection timed out",
			"Internal server error",
			"you have been blocked",
			"SYSTEM MAINTENANCE NOTICE",
			"HTML file is not found",
		}

		for attempt := 1; attempt <= maxRetries; attempt++ {
			html, err = services.RunPage(chromeCtx, &url)

			shouldRetry := false
			if err != nil {
				if strings.Contains(err.Error(), "net::ERR_SOCKET_NOT_CONNECTED") {
					shouldRetry = true
				} else {
					for _, trigger := range retryTriggers {
						if strings.Contains(err.Error(), trigger) {
							shouldRetry = true
							break
						}
					}
				}
			} else {
				for _, trigger := range retryTriggers {
					if strings.Contains(html, trigger) {
						shouldRetry = true
						break
					}
				}
			}

			if !shouldRetry {
				break
			}

			log.Warnf("Retrying ID %d (attempt %d/%d)...", id, attempt, maxRetries)
			time.Sleep(3 * time.Second)
		}
		if err != nil {
			log.Errorf("[Error] Failed to load ID %d: %v", id, err)
			continue
		}
		if strings.Contains(html, "HTML file is not found") {
			log.Warnf("Announcement ID %d not found (404). Skipping.", id)
			continue
		}

		a := &models.Announcement{
			AnnID:   int(id),
			Link:    url,
			Content: html,
		}
		if err := db.SaveAnnouncement(database, a); err != nil {
			log.Errorf("[Error] Failed to save ID %d: %v", id, err)
		} else {
			log.Infof("Saved announcement ID %d", id)
		}
	}

	log.Info("Done scraping all announcements.")
}
