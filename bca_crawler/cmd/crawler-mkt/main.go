package main

import (
	"bca_crawler/internal/db"
	"bca_crawler/internal/services"
	"bca_crawler/internal/utils"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func main() {
	// Load configuration
	cfg, err := utils.LoadCfg()
	if err != nil {
		panic(err)
	}

	// Initialize logger
	utils.InitLogger()
	log := utils.Logger

	// Setup database
	database, err := db.Setup(cfg.DBPath)
	if err != nil {
		log.Fatalf("❌ Failed to setup DB: %v", err)
	}
	defer database.Close()

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

	// ================================================================
	// Create output directory ./ms/YYYYMMDD/
	// ================================================================
	today := time.Now().Format("20060102")
	baseDir := filepath.Join("ms", today)
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		log.Fatalf("[Error] Failed to create directory %s: %v", baseDir, err)
	}
	log.Infof("📂 Download directory: %s", baseDir)

	// ================================================================
	// Prepare HTTP client and headers
	// ================================================================
	client := &http.Client{
		Timeout: 60 * time.Second,
	}

	headers := map[string]string{
		"User-Agent":      cfg.UserAgent,
		"Referer":         cfg.StartURL,
		"Accept":          "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8",
		"Accept-Language": "en-US,en;q=0.9",
		"Connection":      "keep-alive",
	}

	// ================================================================
	// Download each file (with retry up to 3 times)
	// ================================================================
	successCount := 0
	failCount := 0
	maxRetries := 3

	for i, fileURL := range miscURLs {
		fileName := filepath.Base(fileURL)
		fileName = strings.Split(fileName, "?")[0]
		savePath := filepath.Join(baseDir, fileName)

		log.Infof("[%d/%d] ⬇️ Downloading: %s", i+1, len(miscURLs), fileName)

		var success bool
		for attempt := 1; attempt <= maxRetries; attempt++ {
			err := downloadFile(client, headers, fileURL, savePath)
			if err == nil {
				log.Infof("✅ Saved: %s (attempt %d)", savePath, attempt)
				success = true
				successCount++
				break
			} else {
				log.Warnf("⚠️ Attempt %d failed for %s: %v", attempt, fileURL, err)
				time.Sleep(2 * time.Second)
			}
		}

		if !success {
			log.Errorf("[Error] All %d attempts failed for %s", maxRetries, fileURL)
			failCount++
		}
	}

	// ================================================================
	// Summary
	// ================================================================
	log.Infof("🏁 Download complete: %d success, %d failed. Files saved under %s", successCount, failCount, baseDir)
	log.Info("✅ Done scraping and downloading Bursa attachments.")
}

// ================================================================
// downloadFile - helper with headers and file saving logic
// ================================================================
func downloadFile(client *http.Client, headers map[string]string, fileURL, savePath string) error {
	req, err := http.NewRequest("GET", fileURL, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http %d", resp.StatusCode)
	}

	out, err := os.Create(savePath)
	if err != nil {
		return fmt.Errorf("create file: %w", err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("write file: %w", err)
	}

	return nil
}
