package main

import (
	"fmt"
	"strconv"

	"bca_crawler/internal/db"
	"bca_crawler/internal/services"
	"bca_crawler/internal/utils"
)

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
	database, err := db.Setup(cfg.DBPath, db.DriverType(cfg.DBDriver))
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

		change, _, person, background, err := services.ParseBoardroomChangeHTML(ann)
		if err != nil {
			log.Warnf("⚠️ Parse failed for ann_id %s: %v", annID, err)
			continue
		}

		// companyID, err := db.UpdateEntity(database, company)
		// if err != nil {
		// 	log.Errorf("❌ Company update failed for ann_id %s: %v", annID, err)
		// 	continue
		// }

		personID, err := db.UpdateEntity(database, person)
		if err != nil {
			log.Errorf("❌ Person update failed for ann_id %s: %v", annID, err)
			continue
		}

		// change.CompanyID = int(companyID)
		change.PersonID = int(personID)

		if err = db.UpdateBackground(database, personID, background); err != nil {
			log.Errorf("❌ Qualifications update failed for ann_id %s: %v", annID, err)
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
