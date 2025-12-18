package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	"bca_crawler/internal/db"
	"bca_crawler/internal/models"
	"bca_crawler/internal/services"
	"bca_crawler/internal/utils"

	"github.com/google/uuid"
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

		change, err := services.ParseBoardroomChangeHTML(ann)
		if err != nil {
			log.Warnf("⚠️ Parse failed for ann_id %s: %v", annID, err)
			continue
		}

		CompanyID, err := db.GetSCID(database, change.StockCode, "COMPANY")
		if errors.Is(err, sql.ErrNoRows) {
			CompanyID = uuid.New()
			err = db.InsertEntityMaster(database, &models.EntityMaster{
				ScID: CompanyID,
				Type: "COMPANY",
				Name: change.StockCode,
			})
		}
		if err != nil {
			log.Errorf("❌ Company ID lookup failed for ann_id %s: %v", annID, err)
			continue
		}

		PersonID, err := db.GetSCID(database, change.PersonName, "PERSON")
		if errors.Is(err, sql.ErrNoRows) {
			PersonID = uuid.New()
			err = db.InsertEntityMaster(database, &models.EntityMaster{
				ScID: PersonID,
				Type: "PERSON",
				Name: change.PersonName,
			})
		}
		if err != nil {
			log.Errorf("❌ Person ID lookup failed for ann_id %s: %v", annID, err)
			continue
		}

		entities := []models.Entity{
			{
				ScID:      CompanyID,
				Prefix:    "COMPANY",
				Name:      change.CompanyName,
				CreatedAt: *change.DateAnnounced,
			},
			{
				ScID:        PersonID,
				Prefix:      "PERSON",
				Name:        change.PersonName,
				Title:       change.PersonTitle,
				BirthYear:   change.PersonBirthYear,
				Gender:      change.PersonGender,
				Nationality: change.PersonNationality,
				CreatedAt:   *change.DateAnnounced,
			},
		}

		for _, entity := range entities {
			err = db.UpdateEntity(database, &entity)
			if err != nil {
				log.Errorf("❌ Entity update failed for ann_id %s: %v", annID, err)
				continue
			}
		}

		if err = db.UpdateBackground(database, PersonID, &change.Background); err != nil {
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
