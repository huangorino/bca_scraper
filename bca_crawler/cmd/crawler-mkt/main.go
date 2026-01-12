package main

import (
	"bca_crawler/internal/services"
	"bca_crawler/internal/utils"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// crawler for market statistics

func main() {
	// Load configuration
	cfg, err := utils.LoadCfg()
	if err != nil {
		panic(err)
	}

	// Initialize logger
	utils.InitLogger()
	log := utils.Logger

	// Load Bursa main page using Chromedp
	chromeCtx, cancel := services.InitCtx(cfg.UserAgent)
	defer cancel()

	body, err := services.RunPage(chromeCtx, &cfg.StartURL)
	if err != nil {
		log.Fatalf("[Error] Failed to load start page: %v", err)
		return
	}

	// Parse HTML with GoQuery
	const basePrefix = "/misc/missftp"

	var miscURLs []string
	allURLs, err := services.GetURLs(body)
	if err != nil {
		log.Fatalf("[Error] Failed to parse HTML: %v", err)
		return
	}

	// Deduplicate allURLs
	uniqueURLs := make(map[string]struct{})
	for _, url := range allURLs {
		uniqueURLs[url] = struct{}{}
	}
	allURLs = make([]string, 0, len(uniqueURLs))
	for url := range uniqueURLs {
		allURLs = append(allURLs, url)
	}

	if len(allURLs) == 0 {
		log.Warn("[Error] No Bursa attachment URLs found.")
		return
	}

	for _, href := range allURLs {
		if strings.HasPrefix(href, basePrefix) {
			fullURL := "https://www.bursamalaysia.com" + href
			miscURLs = append(miscURLs, fullURL)
		}
	}

	today := time.Now().Format("20060102")
	baseDir := filepath.Join("ms", today)
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		log.Fatalf("[Error] Failed to create directory %s: %v", baseDir, err)
	}
	log.Infof("📂 Download directory: %s", baseDir)

	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	for i, fileURL := range miscURLs {
		err := utils.DownloadFile(client, cfg, fileURL, baseDir)
		if err == nil {
			log.Infof("[%d/%d] ⬇️ Downloaded: %s", i+1, len(miscURLs), fileURL)
		} else {
			log.Warnf("⚠️ Failed to download %s: %v", fileURL, err)
		}
	}

	log.Infof("🏁 Download complete. Files saved under %s", baseDir)
}
