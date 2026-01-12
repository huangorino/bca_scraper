package main

import (
	"bca_crawler/internal/db"
	"bca_crawler/internal/utils"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"strings"
	"time"
)

// parser for announcement attachments download

func main() {
	// -------------------------------------------------------------------------
	// 1️⃣ Load Configuration
	// -------------------------------------------------------------------------
	cfg, err := utils.LoadCfg()
	if err != nil {
		panic(fmt.Sprintf("[Error] Config load failed: %v", err))
	}

	// Initialize logger
	utils.InitLogger()
	log := utils.Logger
	log.Infof("Configuration loaded: %+v", *cfg)

	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("[Error] Failed to create cookie jar: %v", err)
	}

	client := &http.Client{
		Timeout: 60 * time.Second,
		Jar:     jar,
	}

	// -------------------------------------------------------------------------
	// 2️⃣ Connect to Database
	// -------------------------------------------------------------------------
	database, err := db.Connect(cfg.DBPath, db.DriverType(cfg.DBDriver))
	if err != nil {
		log.Fatalf("[Error] Failed to setup DB: %v", err)
	}
	defer database.Close()

	// -------------------------------------------------------------------------
	// 3️⃣ Fetch rows to process
	// -------------------------------------------------------------------------
	data, err := db.FetchAnnouncementsByCategory(database, "attachments")
	if err != nil {
		log.Fatalf("[Error] Failed to fetch attachments: %v", err)
	}

	updated := 0
	for i := range data {
		ann := data[i]
		annID := strconv.Itoa(ann.AnnID)

		// Download attachments
		for _, attURL := range ann.Attachments {
			if attURL == "" {
				continue
			}

			var url string
			if !strings.Contains(attURL, "http") {
				url = cfg.DetailDomain + attURL
			} else {
				url = attURL
			}

			destPath := fmt.Sprintf("attachments/%s/", annID)

			if err := utils.DownloadFile(client, cfg, url, destPath); err != nil {
				log.Errorf("[Error] Failed to download %s: %v", url, err)
				continue
			}

			//wait for 1 second
			time.Sleep(1 * time.Second)
		}

		log.Infof("Downloaded %d attachments for announcement %s", len(ann.Attachments), annID)

		updated++

	}
	log.Infof("Completed. Processed %d records.", updated)
}
