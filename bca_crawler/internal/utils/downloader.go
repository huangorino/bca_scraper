package utils

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func DownloadFile(client *http.Client, cfg *Config, url string, savePath string) error {
	headers := map[string]string{
		"User-Agent":                cfg.UserAgent,
		"Referer":                   cfg.StartURL,
		"Accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
		"Accept-Language":           "en-US,en;q=0.9",
		"Connection":                "keep-alive",
		"Sec-Ch-Ua":                 `"Google Chrome";v="123", "Not:A-Brand";v="8", "Chromium";v="123"`,
		"Sec-Ch-Ua-Mobile":          "?0",
		"Sec-Ch-Ua-Platform":        `"Windows"`,
		"Sec-Fetch-Dest":            "document",
		"Sec-Fetch-Mode":            "navigate",
		"Sec-Fetch-Site":            "none",
		"Sec-Fetch-User":            "?1",
		"Upgrade-Insecure-Requests": "1",
		"Cache-Control":             "max-age=0",
	}

	maxRetries := 3
	retryDelay := time.Second * 2

	var downloadErr error
	for i := 0; i < maxRetries; i++ {
		if i > 0 {
			time.Sleep(retryDelay)
		}

		downloadErr = func() error {
			req, err := http.NewRequest("GET", url, nil)
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

			// Determine filename
			contentDisposition := resp.Header.Get("Content-Disposition")
			_, params, err := mime.ParseMediaType(contentDisposition)
			filename := ""
			if err == nil {
				filename = params["filename"]
			}
			if filename == "" {
				filename = filepath.Base(resp.Request.URL.Path)
			}
			if filename == "" || filename == "." || filename == "/" {
				filename = "attachment.bin"
			}

			// Sanitize filename (basic)
			filename = filepath.Base(filename)
			filename = strings.ReplaceAll(filename, "%20", "_")

			fullPath := filepath.Join(savePath, filename)

			if mkdirErr := os.MkdirAll(savePath, 0755); mkdirErr != nil {
				return fmt.Errorf("failed to create directory %s: %w", savePath, mkdirErr)
			}

			out, err := os.Create(fullPath)
			if err != nil {
				return fmt.Errorf("create file: %w", err)
			}
			defer out.Close()

			_, err = io.Copy(out, resp.Body)
			if err != nil {
				return fmt.Errorf("write file: %w", err)
			}

			return nil
		}()

		if downloadErr == nil {
			return nil
		}
	}

	if downloadErr == nil {
		return nil
	}

	return fmt.Errorf("failed after %d attempts: %w", maxRetries, downloadErr)
}

// GetFileNameFromURL extracts the filename from a URL.
// Kept for compatibility if used elsewhere, though DownloadFile now handles it.
func GetFileNameFromURL(rawURL string) string {
	// This might be redundant now but keeping it safe.
	return filepath.Base(rawURL)
}
