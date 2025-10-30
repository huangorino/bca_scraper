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
	database, err := db.Setup(cfg.DBPath)
	if err != nil {
		log.Fatalf("[Error] Failed to setup DB: %v", err)
	}
	defer database.Close()

	// Load Bursa main page
	body, err := services.InitCtx(&cfg.StartURL, cfg.UserAgent)
	if err != nil {
		log.Fatalf("[Error] Failed to load start page: %v", err)
		return
	}

	log.Info("Page loaded successfully, parsing announcements...")
	maxID := services.GetMaxAnnID(body)
	if maxID == 0 {
		log.Warn("No announcements found. Exiting.")
		return
	}

	log.Infof("Parsed announcements. Max ann_id: %d", maxID)

	startID := 1

	// Fetch existing announcements to determine starting ID
	data, err := db.GetMaxAnnID(database)
	if err != nil {
		log.Infof("[Error] Failed to fetch max ann_id from DB: %v", err)
	} else {
		if data >= maxID {
			log.Info("Database is already up-to-date. No new announcements to scrape.")
			return
		}

		startID = data + 1
	}

	log.Infof("Starting from ann_id: %d", startID)

	for i := startID; i <= maxID; i++ {
		annID := strconv.Itoa(i)
		url := cfg.DetailDomain + cfg.DetailURL + annID
		log.Infof("Processing announcement ID: %s", annID)

		var html string
		var err error
		maxRetries := 3
		for attempt := 1; attempt <= maxRetries; attempt++ {
			html, err = services.InitCtx(&url, cfg.UserAgent)
			if err == nil {
				break
			}
			if strings.Contains(err.Error(), "net::ERR_SOCKET_NOT_CONNECTED") {
				log.Warnf("Retrying ID %d (attempt %d/%d)...", i, attempt, maxRetries)
				time.Sleep(3 * time.Second)
			} else {
				break
			}
		}
		if err != nil {
			log.Errorf("[Error] Failed to load ID %d: %v", i, err)
			continue
		}
		if strings.Contains(html, "HTML file is not found") {
			log.Warnf("Announcement ID %d not found (404). Skipping.", i)
			continue
		}

		a := &models.Announcement{
			AnnID:   i,
			Link:    url,
			Content: html,
		}
		if err := db.SaveAnnouncement(database, a); err != nil {
			log.Errorf("[Error] Failed to save ID %d: %v", i, err)
		} else {
			log.Infof("Saved announcement ID %d", i)
		}
	}

	log.Info("Done scraping all announcements.")
}
