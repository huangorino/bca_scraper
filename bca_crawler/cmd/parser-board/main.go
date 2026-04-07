package main

import (
	"fmt"
	"strconv"
	"strings"

	"bca_crawler/internal/db"
	"bca_crawler/internal/models"
	"bca_crawler/internal/services"
	"bca_crawler/internal/utils"

	"github.com/jmoiron/sqlx"
)

// parser for change in boardroom announcements

var database *sqlx.DB
var log = utils.Logger

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

		title, name := utils.SplitTitle(utils.StringValue(change.PersonName))

		entity := &models.Entity{
			DisplayName: utils.PtrString(strings.TrimSpace(title + " " + name)),
			OriName:     change.PersonName,
			Name:        &name,
			Salutation:  &title,
			StockCode:   change.StockCode,
			BirthYear:   change.PersonBirthYear,
			Gender:      change.PersonGender,
			Nationality: change.PersonNationality,
			CreatedAt:   *change.DateAnnounced,
		}

		permID, err := services.GetOrCreateEntity(log, database, entity, &change.Background)
		if err != nil {
			log.Errorf("❌ Entity lookup/creation failed for ann_id %s: %v", annID, err)
			continue
		}

		change.RelatedPerm = permID
		change.PersonTitle = &title
		change.PersonName = &name

		err = db.UpdateBoardroomChange(database, change)
		if err != nil {
			log.Errorf("❌ Boardroom change update failed for ann_id %s: %v", annID, err)
			continue
		}

		updated++

		log.Infof("🏁 Done. Updated %d records.", updated)
	}
}
