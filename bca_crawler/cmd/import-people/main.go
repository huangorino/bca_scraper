package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"path/filepath"
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
		log.Fatalf("❌ Failed to connect to DB: %v", err)
	}
	defer database.Close()

	// -------------------------------------------------------------------------
	// 3️⃣ Define CSV Directory
	// -------------------------------------------------------------------------
	csvDir := `C:\Users\user\Desktop\csv`
	files, err := filepath.Glob(filepath.Join(csvDir, "*.csv"))
	if err != nil {
		log.Fatalf("❌ Failed to scan CSV directory: %v", err)
	}

	if len(files) == 0 {
		log.Warnf("⚠️ No CSV files found in %s", csvDir)
		return
	}

	log.Infof("📂 Found %d CSV files to process", len(files))

	currentRoleTracker := make(map[string]*models.EntityRole)

	// -------------------------------------------------------------------------
	// 4️⃣ Process Each File
	// -------------------------------------------------------------------------
	for _, filePath := range files {
		log.Infof("📄 Processing file: %s", filepath.Base(filePath))
		
		f, err := os.Open(filePath)
		if err != nil {
			log.Errorf("❌ Failed to open file %s: %v", filePath, err)
			continue
		}

		reader := csv.NewReader(f)
		reader.LazyQuotes = true
		reader.TrimLeadingSpace = true

		// Skip header
		_, err = reader.Read()
		if err != nil {
			log.Errorf("❌ Failed to read header from %s: %v", filePath, err)
			f.Close()
			continue
		}

		var rolePtrs []*models.EntityRole
		rowCount := 0

		for {
			row, err := reader.Read()
			if err == io.EOF {
				break
			}
			if err != nil {
				log.Errorf("⚠️ Error reading row in %s: %v", filePath, err)
				continue
			}

			rowCount++

			// Ensure row has enough columns (mapping expects up to index 21)
			fullRow := make([]string, 22)
			copy(fullRow, row)

			name := strings.TrimSpace(fullRow[6])
			if name == "" {
				name = strings.TrimSpace(fullRow[4]) // Fallback to Display Name
			}
			if name == "" {
				log.Warnf("⚠️ %s | Row %d: Skipping because name is empty", filepath.Base(filePath), rowCount)
				continue
			}

			entity := &models.Entity{
				PrimaryPermID: utils.ParseInt(fullRow[2]),
				DisplayName:   utils.PtrString(strings.TrimSpace(fullRow[4])),
				Salutation:    utils.PtrString(strings.TrimSpace(fullRow[5])),
				Name:          utils.PtrString(name),
				OriName:       utils.PtrString(name),
				BirthYear:     utils.ParseInt(fullRow[7]),
				Gender:        utils.PtrString(strings.TrimSpace(fullRow[8])),
				Nationality:   utils.PtrString(strings.TrimSpace(fullRow[9])),
				ImgPath:       utils.PtrString(strings.TrimSpace(fullRow[3])),
			}

			// Background info
			qual := strings.TrimSpace(fullRow[11])
			if len(fullRow) > 12 && fullRow[12] != "" {
				qual += " (" + strings.TrimSpace(fullRow[12]) + ")"
			}
			if len(fullRow) > 13 && fullRow[13] != "" {
				qual += " - " + strings.TrimSpace(fullRow[13])
			}
			if len(fullRow) > 14 && fullRow[14] != "" {
				qual += "\nProfessional: " + strings.TrimSpace(fullRow[14])
			}

			background := &models.Background{
				Qualification:        qual,
				WorkingExperience:    strings.TrimSpace(fullRow[10]),
				Directorships:        strings.TrimSpace(fullRow[18]),
				ConflictOfInterest:   strings.TrimSpace(fullRow[19]),
				InterestInSecurities: strings.TrimSpace(fullRow[20]),
			}

			permID, err := services.GetOrCreateEntity(log, database, entity, background)
			if err != nil {
				log.Errorf("❌ %s | Row %d: Entity lookup/creation failed for %s: %v", filepath.Base(filePath), rowCount, name, err)
				continue
			}

			input := models.RoleChangeInput{
				PermID:       utils.IntValue(permID),
				StockCode:    strings.TrimSpace(fullRow[1]),
				CompanyName:  strings.TrimSpace(fullRow[0]),
				Designation:  strings.TrimSpace(fullRow[15]),
				TypeOfChange: "APPOINTMENT",
				DateOfChange: utils.ParseDate(fullRow[17]),
				Category:     "DIRECTOR",
			}
			newRoles := services.ProcessSingleRoleChange(input, currentRoleTracker)
			rolePtrs = append(rolePtrs, newRoles...)

			log.Infof("✅ %s | Row %d: Processed %s (PermID: %v)", filepath.Base(filePath), rowCount, name, *permID)
		}
		f.Close()
		log.Infof("📊 Finished processing %s: %d records", filepath.Base(filePath), rowCount)
	}

	log.Info("🏁 All imports complete.")
}
