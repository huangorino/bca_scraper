package main

import (
	"bca_crawler/internal/db"
	"bca_crawler/internal/utils"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"os"
	"path/filepath"
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

			today := ann.DatePosted.Format("20060102")
			baseDir := filepath.Join("attachments", today, annID)
			if err := os.MkdirAll(baseDir, 0755); err != nil {
				log.Fatalf("[Error] Failed to create directory %s: %v", baseDir, err)
			}
			log.Infof("📂 Download directory: %s", baseDir)

			if err := utils.DownloadFile(client, cfg, url, baseDir); err != nil {
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

func buildAttachmentPath(base string, annID int, annDate time.Time) string {
	id := strconv.Itoa(annID)

	// zfill(4)
	if len(id) < 4 {
		id = strings.Repeat("0", 4-len(id)) + id
	}

	shard1 := id[:2]
	shard2 := id[2:4]

	month := annDate.Format("2006-01")

	return filepath.Join(
		base,
		month,
		shard1,
		shard2,
		strconv.Itoa(annID),
	)
}
