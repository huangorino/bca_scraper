package main

import (
	"fmt"
	"os"
	"path/filepath"

	"bca_crawler/internal/db"
	"bca_crawler/internal/utils"
)

func main() {
	// 1. Initialize Logger
	utils.InitLogger()
	log := utils.Logger

	// 2. Load Configuration
	cfg, err := utils.LoadCfg()
	if err != nil {
		log.Fatalf("❌ Failed to load config: %v", err)
	}

	// 3. Connect to Database
	database, err := db.Connect(cfg.DBPath, db.DriverType(cfg.DBDriver))
	if err != nil {
		log.Fatalf("❌ Failed to connect to DB: %v", err)
	}
	defer database.Close()

	// 4. Ingest CSV Data
	store := &DataStore{}
	inputDir := filepath.Join("input")

	log.Infof("🚀 Starting data ingestion from: %s", inputDir)
	if err := store.Ingest(inputDir); err != nil {
		log.Errorf("❌ Ingestion failed: %v", err)
		os.Exit(1)
	}

	// 5. Persist to Database
	log.Infof("💾 Starting database persistence...")
	if err := store.Persist(database); err != nil {
		log.Errorf("❌ Persistence failed: %v", err)
		os.Exit(1)
	}

	log.Infof("✨ Process complete.")
	store.PrintSummary()
}

func (s *DataStore) PrintSummary() {
	fmt.Printf("\n--- In-Memory Data Store Summary ---\n")
	fmt.Printf("IPOs:                  %d\n", len(s.IPOs))
	fmt.Printf("People Profiles:       %d\n", len(s.People))
	fmt.Printf("Corporate Directory:   %d\n", len(s.CorporateDirectory))
	fmt.Printf("Corporate Info:        %d\n", len(s.CorporateInfo))
	fmt.Printf("Company Secretaries:   %d\n", len(s.CompanySecretaries))
	fmt.Printf("Advisers:              %d\n", len(s.Advisers))
	fmt.Printf("Subsidiaries:          %d\n", len(s.Subsidiaries))
	fmt.Printf("Sub Shareholders:      %d\n", len(s.SubShareholders))
	fmt.Printf("Properties Owned:      %d\n", len(s.PropertiesOwned))
	fmt.Printf("Properties Rented:     %d\n", len(s.PropertiesRented))
	fmt.Printf("Relationships:         %d\n", len(s.Relationships))
	fmt.Printf("Major Partners:        %d\n", len(s.MajorPartners))
	fmt.Printf("------------------------------------\n")
}
