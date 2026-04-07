package main

import (
	"fmt"
	"strconv"

	"bca_crawler/internal/db"
	"bca_crawler/internal/services"
	"bca_crawler/internal/utils"
)

// main parser for bursa annoucements

func main() {
	// -------------------------------------------------------------------------
	// 1️⃣ Load Configuration
	// -------------------------------------------------------------------------
	cfg, err := utils.LoadCfg()
	if err != nil {
		panic(fmt.Sprintf("❌ Config load failed: %v", err))
	}

	// Initialize logger
	utils.InitLogger()
	log := utils.Logger
	log.Infof("🔧 Configuration loaded: %+v", *cfg)

	// -------------------------------------------------------------------------
	// 2️⃣ Connect to Database
	// -------------------------------------------------------------------------
	database, err := db.Connect(cfg.DBPath, db.DriverType(cfg.DBDriver))
	if err != nil {
		log.Fatalf("❌ Failed to setup DB: %v", err)
	}
	defer database.Close()

	// -------------------------------------------------------------------------
	// 3️⃣ Fetch rows to process
	// -------------------------------------------------------------------------
	data, err := db.FetchUnparsedAnnouncements(database)
	if err != nil {
		log.Fatalf("❌ Failed to fetch unparsed announcements: %v", err)
	}

	if len(data) == 0 {
		log.Info("⚠️ No unparsed announcements found. Exiting.")
		return
	}

	updated := 0
	for i := range data {
		ann := data[i]
		annID := strconv.Itoa(ann.AnnID)
		log.Infof("Processing ann_id %s", annID)

		// -------------------------------------------------------------------------
		// 4️⃣ Parse Announcement HTML
		// -------------------------------------------------------------------------
		if err := services.ParseAnnouncementHTML(ann); err != nil {
			log.Warnf("⚠️ Parse failed for ann_id %s: %v", annID, err)
			continue
		}

		// -------------------------------------------------------------------------
		// 5️⃣ Update Announcement in DB
		// -------------------------------------------------------------------------
		if err := db.UpdateAnnouncement(database, ann); err != nil {
			log.Errorf("❌ Update failed for ann_id %s: %v", annID, err)
			continue
		}

		updated++
	}

	log.Infof("🏁 Done. Updated %d records.", updated)
}
