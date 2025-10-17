package services

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"
	"strconv"
	"strings"
	"time"

	"bca_crawler/internal/utils"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

// InitCtx launches Chrome headless to scrape and return HTML
func InitCtx(targetURL *string, ua string) (string, error) {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.UserAgent(ua),
	)

	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancelTimeout := context.WithTimeout(ctx, 300*time.Second)
	defer cancelTimeout()

	utils.Logger.Infof("🌐 Navigating to %s", *targetURL)
	var body string

	if err := chromedp.Run(ctx,
		chromedp.Navigate(*targetURL),
		chromedp.Sleep(time.Duration(500+rand.Intn(1500))*time.Millisecond),
		chromedp.ActionFunc(func(ctx context.Context) error {
			_ = network.Enable().Do(ctx)
			chromedp.EvaluateAsDevTools(`() => { try { Object.defineProperty(navigator, 'webdriver', {get: () => undefined}); } catch(e){} }`, nil).Do(ctx)
			chromedp.EvaluateAsDevTools(`() => { try { Object.defineProperty(navigator, 'plugins', {get: () => [1,2,3,4]}); Object.defineProperty(navigator, 'languages', {get: () => ['en-US', 'en']}); } catch(e){} }`, nil).Do(ctx)
			chromedp.EvaluateAsDevTools(`() => { try { window.chrome = window.chrome || {}; window.chrome.webstore = {}; } catch(e){} }`, nil).Do(ctx)
			return nil
		}),
		chromedp.Sleep(2*time.Second),
		chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight * 0.5);`, nil),
		chromedp.Sleep(1*time.Second),
		chromedp.OuterHTML("html", &body, chromedp.ByQuery),
		loadAndCaptureAction(&body),
	); err != nil {
		utils.Logger.Errorf("chromedp run error: %v", err)
		return "", err
	}

	if strings.Contains(strings.ToLower(body), "verify you are human") {
		utils.Logger.Warn("⚠️ Cloudflare verification detected.")
		return "", fmt.Errorf("cloudflare verification detected")
	}
	return body, nil
}

// GetMaxAnnID extracts the highest ann_id value from HTML
func GetMaxAnnID(body string) int {
	maxID := 0
	if !strings.Contains(strings.ToLower(body), "announcement") && !strings.Contains(body, "<table") {
		utils.Logger.Warn("⚠️ Announcement table not detected.")
		return 0
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		utils.Logger.Errorf("goquery parse error: %v", err)
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

func loadAndCaptureAction(body *string) chromedp.ActionFunc {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		if err := network.Enable().Do(ctx); err != nil {
			utils.Logger.Warnf("Network enable error: %v", err)
		}
		chromedp.EvaluateAsDevTools(`() => { Object.defineProperty(navigator, 'webdriver', {get: () => undefined}); }`, nil).Do(ctx)
		chromedp.OuterHTML("html", body, chromedp.ByQuery).Do(ctx)
		return nil
	})
}
