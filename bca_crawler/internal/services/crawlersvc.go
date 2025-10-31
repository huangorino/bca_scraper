package services

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"

	"bca_crawler/internal/models"
	"bca_crawler/internal/utils"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

// InitCtx launches Chrome headless to scrape and return HTML
func InitCtx(ua string) (context.Context, func()) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		// Use the new headless mode
		chromedp.Flag("headless", "new"),
		// chromedp.Flag("headless", false),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-cache", true),
		chromedp.Flag("disk-cache-size", "0"),
		chromedp.Flag("media-cache-size", "0"),
		chromedp.Flag("disable-application-cache", true),

		chromedp.UserAgent(ua),
	)

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)

	ctx, cancel := chromedp.NewContext(allocCtx)

	cleanup := func() {
		utils.Logger.Infof("Cleaning up browser context...")
		cancel()
		cancelAlloc()
		utils.Logger.Infof("Cleanup complete.")
	}

	return ctx, cleanup
}

func RunPage(ctx context.Context, targetURL *string) (string, error) {
	utils.Logger.Infof("Navigating to %s", *targetURL)
	var body string

	if err := chromedp.Run(ctx,
		chromedp.Navigate(*targetURL),
		chromedp.Sleep(time.Duration(500+rand.Intn(1500))*time.Millisecond),
		chromedp.WaitReady("body", chromedp.ByQuery),
		chromedp.ActionFunc(func(ctx context.Context) error {
			_ = network.Enable().Do(ctx)
			chromedp.EvaluateAsDevTools(`() => { try { Object.defineProperty(navigator, 'webdriver', {get: () => undefined}); } catch(e){} }`, nil).Do(ctx)
			chromedp.EvaluateAsDevTools(`() => { try { Object.defineProperty(navigator, 'plugins', {get: () => [1,2,3,4]}); Object.defineProperty(navigator, 'languages', {get: () => ['en-US', 'en']}); } catch(e){} }`, nil).Do(ctx)
			chromedp.EvaluateAsDevTools(`() => { try { window.chrome = window.chrome || {}; window.chrome.webstore = {}; } catch(e){} }`, nil).Do(ctx)
			return nil
		}),
		chromedp.Sleep(1*time.Second),
		chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight * 0.5);`, nil),
		chromedp.Sleep(1*time.Second),
		chromedp.OuterHTML("html", &body, chromedp.ByQuery),
		LoadAndCaptureAction(&body),
	); err != nil {
		utils.Logger.Errorf("[Error] chromedp run error: %v", err)
		return "", err
	}

	if strings.Contains(strings.ToLower(body), "verify you are human") {
		utils.Logger.Warn("Cloudflare verification detected.")
		return "", fmt.Errorf("[Error] cloudflare verification detected")
	}

	return body, nil
}

func LoadAndCaptureAction(body *string) chromedp.ActionFunc {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		if err := network.Enable().Do(ctx); err != nil {
			utils.Logger.Warnf("[Error] Network enable error: %v", err)
		}
		chromedp.EvaluateAsDevTools(`() => { Object.defineProperty(navigator, 'webdriver', {get: () => undefined}); }`, nil).Do(ctx)
		chromedp.OuterHTML("html", body, chromedp.ByQuery).Do(ctx)
		return nil
	})
}

// GetMaxAnnID extracts the highest ann_id value from HTML
func GetMaxAnnID(body string) int {
	maxID := 0
	if !strings.Contains(strings.ToLower(body), "announcement") && !strings.Contains(body, "<table") {
		utils.Logger.Warn("[Error] Announcement table not detected.")
		return 0
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		utils.Logger.Errorf("[Error] goquery parse error: %v", err)
		return 0
	}

	doc.Find("table tbody tr").Each(func(i int, s *goquery.Selection) {
		s.Find("a[href]").Each(func(_ int, a *goquery.Selection) {
			href, exists := a.Attr("href")
			if !exists || !strings.Contains(href, "ann_id") {
				return
			}

			u, err := url.Parse(href)
			if err != nil {
				return
			}

			q := u.Query()
			if idStr := q.Get("ann_id"); idStr != "" {
				if idInt, err := strconv.Atoi(idStr); err == nil && idInt > maxID {
					maxID = idInt
				}
			}
		})
	})

	return maxID
}

func ParseAnnouncementHTML(ann *models.Announcement) error {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(ann.Content))
	if err != nil {
		return fmt.Errorf("[Error] parse HTML: %w", err)
	}

	found := false

	// --------------------------------------------
	// 1. Parse announcement info table
	// --------------------------------------------
	doc.Find("table").EachWithBreak(func(i int, s *goquery.Selection) bool {
		// Only parse tables containing "Company Name"
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
				case strings.EqualFold(label, "Reference Number") || strings.EqualFold(label, "Reference No"):
					ann.RefNumber = value
				}
			}
		})

		found = true
		return false
	})

	if !found {
		return fmt.Errorf("[Error] announcement info table not found")
	}

	// --------------------------------------------
	// 2. Parse attachment URLs
	// --------------------------------------------
	var attachments []string

	doc.Find("p.att_download_pdf a").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			return
		}
		href = strings.TrimSpace(href)
		if href == "" {
			return
		}

		// Convert &amp; → &
		href = utils.HtmlUnescape(href)

		attachments = append(attachments, href)
	})

	if len(attachments) > 0 {
		ann.Attachments = attachments
	}

	return nil
}

func GetURLs(body string) ([]string, error) {
	var urls []string
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("[Error] goquery parse error: %w", err)
	}

	doc.Find("a[href]").Each(func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if !exists {
			return
		}
		href = strings.TrimSpace(href)
		if href == "" {
			return
		}

		urls = append(urls, href)
	})

	return urls, nil
}
