package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

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
	data, err := db.FetchShareHoldingChanges(database)
	if err != nil {
		log.Fatalf("❌ Failed to fetch shareholder announcements: %v", err)
	}

	if len(data) == 0 {
		log.Info("⚠️ No shareholder announcements found. Exiting.")
		return
	}

	for i := range data {
		ann := data[i]
		annID := strconv.Itoa(ann.AnnID)

		entityType := "Individual"

		layouts := []string{
			"BHD",
			"BERHAD",
			"LTD",
			"LIMITED",
			"LP",
			"L.P.",
			"PRIVATED",
			"FUND",
			"TRUST",
			"TABUNG",
			"LEMBAGA",
			"AMANAH",
			"GROUP",
			"KUMPULAN",
			"CORP",
			"FOUNDATION",
			"HOLDING",
			"ASSOCIATION",
			"PLC",
			"UCITS",
			"BANK",
			"YAYASAN",
			"B.V.",
			"BV",
			"SE",
			"AKTIENGESELLSCHAFT",
			"ESTATE",
			"R.L.",
			"INVESTMENTS",
			"ULC",
			"LLC",
			"INC",
			"SZA",
		}

		for _, layout := range layouts {
			if strings.Contains(utils.StringValue(ann.PersonName), layout) {
				entityType = "Company"
				break
			}
		}

		if entityType == "Individual" {
			title, name := utils.SplitTitle(utils.StringValue(ann.PersonName))

			entity := &models.Entity{
				DisplayName: utils.PtrString(strings.TrimSpace(title + " " + name)),
				OriName:     ann.PersonName,
				Name:        &name,
				Salutation:  &title,
				StockCode:   &ann.StockCode,
				Nationality: ann.PersonNationality,
				CreatedAt:   time.Now(),
			}

			permID, err := services.GetOrCreateEntity(log, database, entity, nil)
			if err != nil {
				log.Errorf("❌ Entity lookup/creation failed for ann_id %s: %v", annID, err)
				continue
			}

			ann.RelatedPerm = permID

			// insert into db
			if err := db.UpdateShareholdingChange(database, []*models.ShareholdingChange{&ann}); err != nil {
				log.Warnf("⚠️ DB update failed for ann_id %s: %v", annID, err)
				continue
			}

			log.Infof("Processing ann_id %s | Title: %s | Name: %s", annID, title, name)
		}
	}
}

// func old() {

// 	data, err := db.FetchStockList(database)
// 	if err != nil {
// 		log.Fatalf("❌ Failed to fetch stock list: %v", err)
// 	}

// 		var entities []models.Entity

// 		if stock.Management != nil {
// 			var managers []models.Management
// 			if err := json.Unmarshal(*stock.Management, &managers); err != nil {
// 				log.Warnf("⚠️ Failed to unmarshal management for %s: %v", stockName, err)
// 			} else {
// 				log.Infof("Management for %s:", stockName)
// 				for _, m := range managers {
// 					title, name := utils.SplitTitle(utils.StringValue(&m.Name))

// 					entities = append(entities, models.Entity{
// 						Name:       &name,
// 						StockCode:  &stock.StockCode,
// 						OriName:    &m.Name,
// 						Salutation: &title,
// 					})

// 					log.Infof("OriName: %s | Name: %s | Salutation: %s | Role: %s | Since: %s | Designation: %s", m.Name, name, title, m.Role, m.Since, m.Designation)
// 				}
// 			}
// 		}

// }
