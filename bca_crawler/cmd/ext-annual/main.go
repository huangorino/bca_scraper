package main

import (
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

}
