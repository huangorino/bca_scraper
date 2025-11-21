package services

import (
	"context"
	"fmt"
	"math/rand"
	"strings"
	"time"

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
