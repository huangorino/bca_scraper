package main

import (
	"strings"
	"time"

	"bca_crawler/internal/db"
	"bca_crawler/internal/models"
	"bca_crawler/internal/services"
	"bca_crawler/internal/utils"

	"github.com/samber/lo"
)

// main crawler for announcements

func main() {
	// Load configuration
	cfg, err := utils.LoadCfg()
	if err != nil {
		panic(err)
	}

	// Initialize logger with level from config
	utils.InitLogger()
	log := utils.Logger

	// Setup database
	database, err := db.Connect(cfg.DBPath, db.DriverType(cfg.DBDriver))
	if err != nil {
		log.Fatalf("[Error] Failed to setup DB: %v", err)
	}
	defer database.Close()

	// -------------------------------------------------------------------------
	// 3️⃣ Fetch rows to process
	// -------------------------------------------------------------------------
	data, err := db.FetchStockList(database)
	if err != nil {
		log.Fatalf("❌ Failed to fetch stock list: %v", err)
	}

	if len(data) == 0 {
		log.Info("⚠️ No stock list found. Exiting.")
		return
	}

	result := lo.Filter(data, func(stock models.Stock, _ int) bool {
		return stock.Type != nil && *stock.Type == "stocks"
	})

	// Load main page
	chromeCtx, cancel := services.InitCtx(cfg.UserAgent)
	defer cancel()

	url := "https://businessreport.ctoscredit.com.my/oneoffreport/search-result-page"

	for _, stock := range result {
		log.Infof("Processing stock: %s", stock.StockCode)

		var html string
		var err error
		maxRetries := 3
		retryTriggers := []string{
			"Web server is returning an unknown error",
			"SSL handshake failed",
			"A timeout occurred",
			"Just a moment",
			"Connection timed out",
			"Internal server error",
			"you have been blocked",
			"SYSTEM MAINTENANCE NOTICE",
			"HTML file is not found",
		}

		for attempt := 1; attempt <= maxRetries; attempt++ {
			searchTerm := ""
			if stock.Name != nil {
				searchTerm = *stock.Name
			}
			html, err = services.RunRocSearch(chromeCtx, &url, searchTerm)

			if html != "" {
				regNum := services.GetRegNum(html)
				if regNum != "" {
					log.Infof("✅ Found Registration Number: %s for %s", regNum, stock.StockCode)

					parts := strings.Split(regNum, "/")
					var oldReg, newReg string
					if len(parts) == 2 {
						oldReg = strings.TrimSpace(parts[0])
						newReg = strings.TrimSpace(parts[1])
						log.Infof("Old Reg No: %s, New Reg No: %s", oldReg, newReg)
					} else {
						// Fallback if there's no '/' separator
						newReg = strings.TrimSpace(regNum)
						log.Infof("Reg No: %s (Single part found)", newReg)
					}

					if err := db.UpdateStockRegNumbers(database, stock.ID, newReg, oldReg); err != nil {
						log.Errorf("❌ Failed to update DB for %s: %v", stock.StockCode, err)
					} else {
						log.Infof("✅ Updated DB for %s", stock.StockCode)
					}
				} else {
					log.Warnf("⚠️ Could not find Registration Number in results for %s", stock.StockCode)
				}
			}

			shouldRetry := false
			if err != nil {
				if strings.Contains(err.Error(), "net::ERR_SOCKET_NOT_CONNECTED") {
					shouldRetry = true
				} else {
					for _, trigger := range retryTriggers {
						if strings.Contains(err.Error(), trigger) {
							shouldRetry = true
							break
						}
					}
				}
			} else {
				for _, trigger := range retryTriggers {
					if strings.Contains(html, trigger) {
						shouldRetry = true
						break
					}
				}
			}

			if !shouldRetry {
				break
			}

			log.Warnf("Retrying (attempt %d/%d)...", attempt, maxRetries)
			time.Sleep(3 * time.Second)
		}
		if err != nil {
			log.Errorf("[Error] Failed to load: %v", err)
			continue
		}
	}

	log.Info("Done scraping all announcements.")
}
