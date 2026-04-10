package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"bca_crawler/internal/services"

	"bca_crawler/internal/db"
	"bca_crawler/internal/models"
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

	ParseShareholdingChange()
	UpdateShareHoldingPerm()
}

func ParseShareholdingChange() {
	// -------------------------------------------------------------------------
	// 3️⃣ Fetch rows to process
	// -------------------------------------------------------------------------
	data, err := db.FetchAnnouncementsByShareholder(database)
	if err != nil {
		log.Fatalf("❌ Failed to fetch shareholder announcements: %v", err)
	}

	if len(data) == 0 {
		log.Info("⚠️ No shareholder announcements found. Exiting.")
		return
	}

	updated := 0
	for i := range data {
		ann := data[i]
		annID := strconv.Itoa(ann.AnnID)

		change, err := services.ParseShareholdingChange(ann)
		if err != nil {
			log.Warnf("⚠️ Parse failed for ann_id %s: %v", annID, err)
			continue
		}

		// insert into db
		if err := db.UpdateShareholdingChange(database, change); err != nil {
			log.Warnf("⚠️ DB update failed for ann_id %s: %v", annID, err)
			continue
		}

		log.Infof("🏁 Done. Processed %d records.", updated)
		updated++
	}
}

func UpdateShareHoldingPerm() {
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

	uniqueCompanies := make(map[string]struct{})
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
			"PERKESO",
			"FUND",
			"TRUST",
			"TABUNG",
			"LEMBAGA",
			"AMANAH",
			"PERBADANAN",
			"PEMERBADANAN",
			"PERTUBUHAN",
			"ASSET",
			"GROUP",
			"KUMPULAN",
			"KWSP",
			"ASIA",
			"CORP",
			"COMPANY",
			"COMPANIES",
			"EKONOMI",
			"ECONOMY",
			"FOUNDATION",
			"FINANCE",
			"FINANCIAL",
			"STATE",
			"SECRETARY",
			"HOLDING",
			"PROSPERITY",
			"ASSOCIATION",
			"ASSOCIATE",
			"AUTHORITY",
			"FEDERAL",
			"SYSTEM",
			"DEVELOPMENT",
			"PLC",
			"P L C",
			"PLT",
			"LLC",
			"L L C",
			"UCITS",
			"BANK",
			"YAYASAN",
			"B.V.",
			"BV",
			"PT",
			"ANSTALT",
			"SCHAFT",
			"PRIVATSTIFTUNG",
			"GMBH",
			"REALTY",
			"SETTLEMENT",
			"SECRETARY",
			"CREDIT",
			"EQUITY",
			"EQUITIES",
			"CAPITAL",
			"ESTATE",
			"BOARD",
			"EMPLOYEE",
			"NOMINEES",
			"ASSURANCE",
			"R.L.",
			"INVEST",
			"INTERNATIONAL",
			"INTL",
			"NASIONAL",
			"ULC",
			"UBS",
			"INC",
			"SZA",
			"AKTIEBOLAG",
			"ALLIANZ",
			"JP MORGAN",
			"JPMORGAN",
			"NIPPON",
			"ARICINQ",
			"ANIMA",
			"AXA",
			"BTS",
			"CYC",
			"CARLSBERG",
			"DEUTSCHE",
			"FINANCIERE",
			"LAFARGE",
			"NESTL",
			"PACIFIC",
			"OFFSHORE",
			"SKAGEN",
			"SOFIMA",
			"SOFIMO",
			"SOCIETE",
			"STATE",
			"TECH",
			"TELENOR",
		}

		for _, layout := range layouts {
			if strings.Contains(utils.StringValue(ann.PersonName), layout) {
				entityType = "Company"

				uniqueCompanies[utils.StringValue(ann.PersonName)] = struct{}{}

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
				Nationality: ann.PersonNationality,
				CreatedAt:   time.Now(),
			}

			permID, err := services.GetOrCreateEntity(log, database, entity, nil)
			if err != nil {
				log.Errorf("❌ Entity lookup/creation failed for ann_id %s: %v", annID, err)
				continue
			}

			// update into db
			if err := db.UpdateShareholdingChangePerm(database, ann.ID, *permID); err != nil {
				log.Errorf("❌ Failed to update shareholding change %d with permID %v: %v", ann.ID, permID, err)
			}

			log.Infof("Processing ann_id %s | Title: %s | Name: %s", annID, title, name)
		}
	}

	fmt.Println("Distinct Companies Found:")
	for c := range uniqueCompanies {
		fmt.Println("-", c)
	}
}
