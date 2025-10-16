package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	_ "github.com/mattn/go-sqlite3"
)

func setupDB(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}

func initCtx(startURL *string) (string, error) {
	ua := flag.String("ua", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36", "User-Agent used for HTTP downloads")

	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", false),
		chromedp.Flag("disable-blink-features", "AutomationControlled"),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.UserAgent(*ua),
	)
	allocCtx, cancelAlloc := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancelAlloc()

	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()

	ctx, cancelTimeout := context.WithTimeout(ctx, 300*time.Second)
	defer cancelTimeout()

	log.Printf("Navigating to %s", *startURL)
	var body string

	actions := []chromedp.Action{
		chromedp.Navigate(*startURL),

		// Optional extra wait for Vue rendering (just to be safe)
		chromedp.Sleep(2 * time.Second),

		loadAndCaptureAction(&body),
	}

	if err := chromedp.Run(ctx, actions...); err != nil {
		log.Printf("chromedp run error: %v; stopping crawl.", err)
		return "", err
	}

	if strings.Contains(strings.ToLower(body), "verify you are human") || strings.Contains(strings.ToLower(body), "please verify") {
		log.Println("⚠️ Detected Cloudflare/human verification. Please complete it manually in visible mode. Stopping crawl.")
		return "", fmt.Errorf("cloudflare/human verification detected")
	}

	return body, nil
}

func getMaxAnnID(body string) int {
	maxID := 0

	if !strings.Contains(strings.ToLower(body), "announcement") && !strings.Contains(body, "<table") {
		log.Println("⚠️ Announcement table not detected on this page. Stopping crawl.")
		if len(body) > 500 {
			log.Printf("Page preview: %.500s\n", body)
		}
		return 0
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		log.Printf("goquery parse error: %v", err)
		return 0
	}

	// extract announcements on this page
	doc.Find("table tbody tr").Each(func(i int, s *goquery.Selection) {
		// Skip rows that are inside a footer (either the row itself or its table ancestor)
		if s.ParentsFiltered("footer").Length() > 0 || tableAncestorInFooter(s) {
			return
		}

		var link string
		if a := s.Find("a[href]").First(); a.Length() > 0 {
			if h, ok := a.Attr("href"); ok {
				link = h
			}
		}

		// If no link, skip
		if link == "" {
			return
		}

		if u, err := url.Parse(link); err == nil {
			q := u.Query()
			if idInt, err := strconv.Atoi(q.Get("ann_id")); err == nil && idInt > maxID {
				maxID = idInt
			}
		}
	})

	return maxID
}

func extractAnnouncement(body, detailURL string) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		log.Printf("goquery parse error: %v", err)
	}

	title := strings.TrimSpace(doc.Find("#main > h3").First().Text())
	if title == "" {
		log.Println("⚠️ Announcement title not found.")
	}

	var annID string
	u, err := url.Parse(detailURL)
	if err == nil {
		q := u.Query()
		annID = q.Get("ann_id")
	}
	log.Printf("Announcement ID: %s, Title: %s", annID, title)

	// extract announcements on this page
	doc.Find("div.ven_announcement_info table tbody tr").Each(func(i int, s *goquery.Selection) {
		s.Find("td:nth-child(2)").Text()
	})
}

func main() {
	maxID := 0
	startURL := flag.String("start", "https://www.bursamalaysia.com/market_information/announcements/company_announcement", "Start URL for company announcements")
	detailURL := flag.String("detail", "https://www.bursamalaysia.com/market_information/announcements/company_announcement/announcement_details?ann_id=", "Announcements Details")
	dbPath := flag.String("db", "bursa.db", "SQLite DB path")

	db, err := setupDB(*dbPath)
	if err != nil {
		log.Fatalf("db setup: %v", err)
	}
	defer db.Close()

	body, err := initCtx(startURL)
	if err != nil {
		log.Fatalf("failed to initialize context and load page: %v", err)
		return
	}

	log.Println("✅ Page loaded successfully, parsing announcements...")
	maxID = getMaxAnnID(body)
	if maxID == 0 {
		log.Println("No announcements found on the page. Exiting.")
		return
	}
	log.Printf("✅ Parsed announcements on page. Max ann_id seen: %d", maxID)

	for i := 3598874; i <= maxID; i++ {
		url := *detailURL + strconv.Itoa(i)
		log.Printf("Processing announcement ID: %s", url)

		body, err := initCtx(&url)
		if err != nil {
			log.Printf("failed to load announcement ID %d: %v", i, err)
			continue
		}

		extractAnnouncement(body, url)
	}

	// rows, err := db.Query("SELECT title, link, date_raw, pdf_path FROM announcements ORDER BY seen_at DESC LIMIT 50")
	// if err == nil {
	// 	defer rows.Close()
	// 	var out []Announcement
	// 	for rows.Next() {
	// 		var a Announcement
	// 		var pdfPath sql.NullString
	// 		rows.Scan(&a.Title, &a.Link, &a.DateRaw, &pdfPath)
	// 		out = append(out, a)
	// 	}
	// 	b, _ := json.MarshalIndent(out, "", "  ")
	// 	fmt.Println(string(b))
	// }
	log.Println("✅ Done.")
}

func absoluteURL(base, href string) string {
	if href == "" {
		return href
	}
	u, err := url.Parse(href)
	if err == nil && u.IsAbs() {
		return u.String()
	}
	b, err := url.Parse(base)
	if err != nil {
		return href
	}
	return b.ResolveReference(u).String()
}

// Announcement represents a single announcement entry
type Announcement struct {
	AnnID       string `json:"ann_id"`
	Title       string `json:"title"`
	Link        string `json:"link"`
	CompanyName string `json:"company_name,omitempty"`
	StockName   string `json:"stock_name,omitempty"`
	DatePosted  string `json:"date_posted,omitempty"`
	Category    string `json:"category,omitempty"`
	RefNumber   string `json:"ref_number,omitempty"`
	Content     string `json:"description,omitempty"`
}

const schema = `
CREATE TABLE IF NOT EXISTS announcements (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
	ann_id TEXT UNIQUE,
    title TEXT,
    link TEXT UNIQUE,
    date_raw TEXT,
    description TEXT,
    date_updated DATETIME
);
CREATE INDEX IF NOT EXISTS idx_ann_date_updated ON announcements(date_updated);
`

// func saveAnnouncement(db *sql.DB, a *Announcement) error {
// 	now := time.Now().UTC()
// 	_, err := db.Exec(`
// 		INSERT INTO announcements(
// 		ann_id, title, link, date_raw, description, date_updated)
// 		VALUES (?, ?, ?, ?, ?, ?)
// 		ON CONFLICT(ann_id)
// 		DO UPDATE SET
// 		ann_id=excluded.ann_id,
// 		title=excluded.title,
// 		date_raw=excluded.date_raw,
// 		description=excluded.description,
// 		date_updated=?;`,
// 		a.AnnID, a.Title, a.Link, a.DateRaw, a.Description, now)
// 	return err
// }

// tableAncestorInFooter returns true if the nearest table ancestor of sel
// is contained within a <footer> element.
func tableAncestorInFooter(sel *goquery.Selection) bool {
	// find the nearest table ancestor
	tbl := sel.ParentsFiltered("table").First()
	if tbl.Length() == 0 {
		return false
	}
	// check if that table has a footer ancestor
	return tbl.ParentsFiltered("footer").Length() > 0
}

// loadAndCapture runs the small anti-detection steps (enable network, spoof
// navigator properties), waits for readyState, probes for expected content,
// scrolls, and captures the page HTML into the provided body pointer.
func loadAndCaptureAction(body *string) chromedp.ActionFunc {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		if err := network.Enable().Do(ctx); err != nil {
			// continue even if network enable fails
		}
		chromedp.EvaluateAsDevTools(`() => { try { Object.defineProperty(navigator, 'webdriver', {get: () => undefined}); } catch(e){} }`, nil).Do(ctx)
		chromedp.EvaluateAsDevTools(`() => { try { Object.defineProperty(navigator, 'plugins', {get: () => [1,2,3,4]}); Object.defineProperty(navigator, 'languages', {get: () => ['en-US', 'en']}); } catch(e){} }`, nil).Do(ctx)
		chromedp.EvaluateAsDevTools(`() => { try { window.chrome = window.chrome || {}; window.chrome.webstore = {}; } catch(e){} }`, nil).Do(ctx)

		var readyState string
		for i := 0; i < 20; i++ {
			if err := chromedp.Evaluate(`document.readyState`, &readyState).Do(ctx); err != nil {
			} else if readyState == "complete" {
				break
			}
			time.Sleep(200 * time.Millisecond)
		}

		var bodyProbe string
		for i := 0; i < 25; i++ {
			if err := chromedp.OuterHTML("html", &bodyProbe, chromedp.ByQuery).Do(ctx); err == nil {
				lower := strings.ToLower(bodyProbe)
				if strings.Contains(lower, "announcement") || strings.Contains(lower, "<table") {
					break
				}
			}
			time.Sleep(300 * time.Millisecond)
		}

		chromedp.Sleep(500 * time.Millisecond).Do(ctx)
		chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight * 0.5);`, nil).Do(ctx)
		chromedp.Sleep(800 * time.Millisecond).Do(ctx)
		chromedp.OuterHTML("html", body, chromedp.ByQuery).Do(ctx)
		return nil
	})
}
