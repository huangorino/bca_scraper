package main

import (
	"fmt"
	"strconv"

	"bca_crawler/internal/db"
	"bca_crawler/internal/services"
	"bca_crawler/internal/utils"

	"github.com/jmoiron/sqlx"
)

// parser for change in boardroom announcements

var database *sqlx.DB

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
	database, err = db.Setup(cfg.DBPath, db.DriverType(cfg.DBDriver))
	if err != nil {
		log.Fatalf("❌ Failed to setup DB: %v", err)
	}
	defer database.Close()

	// -------------------------------------------------------------------------
	// 3️⃣ Fetch rows to process
	// -------------------------------------------------------------------------
	data, err := db.FetchAnnouncementsByCategory(database, "Change in Boardroom")
	if err != nil {
		log.Fatalf("❌ Failed to fetch change in boardroom announcements: %v", err)
	}

	if len(data) == 0 {
		log.Info("⚠️ No change in boardroom announcements found. Exiting.")
		return
	}

	updated := 0
	for i := range data {
		ann := data[i]
		annID := strconv.Itoa(ann.AnnID)

		change, err := services.ParseBoardroomChangeHTML(ann)
		if err != nil {
			log.Warnf("⚠️ Parse failed for ann_id %s: %v", annID, err)
			continue
		}

		err = GetOrCreateEntity(*change)
		if err != nil {
			log.Errorf("❌ Entity lookup/creation failed for ann_id %s: %v", annID, err)
			continue
		}

		err = db.UpdateBoardroomChange(database, change)
		if err != nil {
			log.Errorf("❌ Boardroom change update failed for ann_id %s: %v", annID, err)
			continue
		}

		updated++

		log.Infof("🏁 Done. Updated %d records.", updated)
	}
}
