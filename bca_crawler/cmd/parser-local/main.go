package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/KEDigitalMY/kedai_models/db"
	"github.com/KEDigitalMY/kedai_models/models"
	"bca_crawler/internal/services"
	"github.com/KEDigitalMY/kedai_models/utils"
)

func main() {
	// -------------------------------------------------------------------------
	// 1️⃣ Load Configuration
	// -------------------------------------------------------------------------
	cfg, err := utils.LoadCfg()
	if err != nil {
		panic(fmt.Sprintf("❌ Failed to load config: %v", err))
	}

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
	// 3️⃣ Determine Input Folder (Default: ./CA)
	// -------------------------------------------------------------------------
	inputDir := "./CA"

	log.Infof("📁 Scanning directory recursively: %s", inputDir)

	// -------------------------------------------------------------------------
	// 4️⃣ Walk through all subfolders and collect .html files
	// -------------------------------------------------------------------------
	var htmlFiles []string

	err = filepath.WalkDir(inputDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(strings.ToLower(d.Name()), ".htm") {
			htmlFiles = append(htmlFiles, path)
		}
		return nil
	})
	if err != nil {
		log.Fatalf("❌ Failed to scan directory %s: %v", inputDir, err)
	}

	if len(htmlFiles) == 0 {
		log.Warnf("⚠️ No HTML files found in %s. Exiting.", inputDir)
		return
	}

	log.Infof("📦 Found %d HTML files under %s", len(htmlFiles), inputDir)

	// -------------------------------------------------------------------------
	// 5️⃣ Process Each HTML File
	// -------------------------------------------------------------------------
	parsed := 0

	for _, path := range htmlFiles {
		fileName := filepath.Base(path)
		annID := strings.TrimSuffix(fileName, ".htm")

		content, err := os.ReadFile(path)
		if err != nil {
			log.Warnf("⚠️ Failed to read file %s: %v", path, err)
			continue
		}

		// Prepare Announcement struct for parser
		annIDInt, err := strconv.Atoi(annID)
		if err != nil {
			log.Warnf("⚠️ Invalid AnnID %s: %v", annID, err)
			continue
		}

		ann := &models.Announcement{
			AnnID:   annIDInt,
			Content: string(content),
		}

		// ---------------------------------------------------------------------
		// 6️⃣ Parse HTML
		// ---------------------------------------------------------------------
		if err := services.ParseAnnouncementHTML(ann); err != nil {
			log.Warnf("⚠️ Parse failed for %s: %v", annID, err)
			continue
		}

		// -------------------------------------------------------------------------
		// 5️⃣ Update Announcement in DB
		// -------------------------------------------------------------------------
		if err := db.UpdateAnnouncement(database, ann); err != nil {
			log.Errorf("❌ Update failed for ann_id %s: %v", annID, err)
			continue
		}

		parsed++
	}

	log.Infof("🏁 Done. Updated %d records.", parsed)
}
