package main

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"

	"bca_crawler/internal/db"
	"bca_crawler/internal/models"
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
	database, err := db.Connect(cfg.DBPath)
	if err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}
	defer database.Close()

	// -------------------------------------------------------------------------
	// 3️⃣ Fetch rows to process
	// -------------------------------------------------------------------------
	rows, err := database.Query(`
		SELECT id, ann_id, content 
		FROM announcements 
		WHERE content IS NOT NULL AND content != ''`)
	if err != nil {
		log.Fatalf("❌ Failed to query announcements: %v", err)
	}
	defer rows.Close()

	updated := 0
	for rows.Next() {
		var ann models.Announcement
		if err := rows.Scan(&ann.ID, &ann.AnnID, &ann.Content); err != nil {
			log.Errorf("Row scan error: %v", err)
			continue
		}

		// ---------------------------------------------------------------------
		// 4️⃣ Parse HTML content and extract announcement info
		// ---------------------------------------------------------------------
		if err := parseAnnouncementHTML(&ann); err != nil {
			log.Warnf("⚠️ Parse failed for ann_id %s: %v", ann.AnnID, err)
			continue
		}

		// ---------------------------------------------------------------------
		// 5️⃣ Update DB with parsed fields
		// ---------------------------------------------------------------------
		if err := db.UpdateAnnouncement(database, &ann); err != nil {
			log.Errorf("❌ Update failed for ann_id %s: %v", ann.AnnID, err)
			continue
		}

		updated++
		log.Infof("✅ Updated ann_id %s — %s | %s", ann.AnnID, ann.CompanyName, ann.StockName)
	}

	log.Infof("🏁 Done. Updated %d records.", updated)
}

// -----------------------------------------------------------------------------
// HTML Parser
// -----------------------------------------------------------------------------

func parseAnnouncementHTML(ann *models.Announcement) error {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(ann.Content))
	if err != nil {
		return fmt.Errorf("parse HTML: %w", err)
	}

	found := false
	doc.Find("table").EachWithBreak(func(i int, s *goquery.Selection) bool {
		// only parse tables containing "Company Name"
		if !strings.Contains(strings.ToLower(s.Text()), "company name") {
			return true
		}

		s.Find("tr").Each(func(_ int, tr *goquery.Selection) {
			tds := tr.Find("td")
			if tds.Length() >= 2 {
				label := utils.CleanString(tds.Eq(0).Text())
				value := utils.CleanString(tds.Eq(1).Text())

				switch {
				case strings.EqualFold(label, "Company Name"):
					ann.CompanyName = value
				case strings.EqualFold(label, "Stock Name"):
					ann.StockName = value
				case strings.EqualFold(label, "Date Announced"):
					ann.DatePosted = utils.ParseDate(value)
				case strings.EqualFold(label, "Category"):
					ann.Category = value
				case strings.EqualFold(label, "Reference Number"):
					ann.RefNumber = value
				}
			}
		})

		found = true
		return false
	})

	if !found {
		return fmt.Errorf("announcement info table not found")
	}
	return nil
}
