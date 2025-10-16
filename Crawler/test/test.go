package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/url"
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

func main() {
	startURL := flag.String("start", "https://www.bursamalaysia.com/market_information/announcements/company_announcement", "Start URL for company announcements")
	ua := flag.String("ua", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36", "User-Agent used for HTTP downloads")
	dbPath := flag.String("db", "bursa.db", "SQLite DB path")

	db, err := setupDB(*dbPath)
	if err != nil {
		log.Fatalf("db setup: %v", err)
	}
	defer db.Close()

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

	current := *startURL

	for current != "" {
		log.Printf("Navigating to %s", current)
		var body string

		actions := []chromedp.Action{
			chromedp.Navigate(current),
			chromedp.Sleep(time.Duration(500+rand.Intn(1500)) * time.Millisecond),
			chromedp.ActionFunc(func(ctx context.Context) error {
				_ = network.Enable().Do(ctx)
				chromedp.EvaluateAsDevTools(`() => { try { Object.defineProperty(navigator, 'webdriver', {get: () => undefined}); } catch(e){} }`, nil).Do(ctx)
				chromedp.EvaluateAsDevTools(`() => { try { Object.defineProperty(navigator, 'plugins', {get: () => [1,2,3,4]}); Object.defineProperty(navigator, 'languages', {get: () => ['en-US', 'en']}); } catch(e){} }`, nil).Do(ctx)
				chromedp.EvaluateAsDevTools(`() => { try { window.chrome = window.chrome || {}; window.chrome.webstore = {}; } catch(e){} }`, nil).Do(ctx)
				return nil
			}),
			chromedp.Sleep(2 * time.Second),
			chromedp.Evaluate(`window.scrollTo(0, document.body.scrollHeight * 0.5);`, nil),
			chromedp.Sleep(1 * time.Second),
			chromedp.OuterHTML("html", &body, chromedp.ByQuery),
		}

		if err := chromedp.Run(ctx, actions...); err != nil {
			log.Printf("chromedp run error: %v; stopping crawl.", err)
			break
		}

		if strings.Contains(strings.ToLower(body), "verify you are human") || strings.Contains(strings.ToLower(body), "please verify") {
			log.Println("⚠️ Detected Cloudflare/human verification. Please complete it manually in visible mode. Stopping crawl.")
			break
		}

		if !strings.Contains(strings.ToLower(body), "announcement") && !strings.Contains(body, "<table") {
			log.Println("⚠️ Announcement table not detected on this page. Stopping crawl.")
			if len(body) > 500 {
				log.Printf("Page preview: %.500s\n", body)
			}
			break
		}

		log.Println("✅ Page loaded successfully, parsing announcements...")

		doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
		if err != nil {
			log.Printf("goquery parse error: %v", err)
			break
		}

		// Robust announcement anchor selector (prioritize ann_id / announcements paths, avoid company-profile)
		findAnnouncementAnchor := func(s *goquery.Selection, baseURL string) (string, string) {
			// 1) Prefer anchor within common title cells (td index 1 or 2)
			for _, idx := range []int{1, 2} {
				td := s.Find("td").Eq(idx)
				if td.Length() == 0 {
					continue
				}
				a := td.Find("a[href]").FilterFunction(func(i int, sel *goquery.Selection) bool {
					h, _ := sel.Attr("href")
					return strings.Contains(h, "ann_id=") || strings.Contains(h, "/announcements/") || strings.Contains(h, "company_announcement")
				}).First()
				if a.Length() > 0 {
					href, _ := a.Attr("href")
					title := strings.TrimSpace(a.Text())
					return title, absoluteURL(baseURL, href)
				}
				// If no prioritized anchor, take first anchor that's not company-profile
				a = td.Find("a[href]").FilterFunction(func(i int, sel *goquery.Selection) bool {
					h, _ := sel.Attr("href")
					return !strings.Contains(h, "company-profile") && !strings.Contains(h, "/company/")
				}).First()
				if a.Length() > 0 {
					href, _ := a.Attr("href")
					title := strings.TrimSpace(a.Text())
					return title, absoluteURL(baseURL, href)
				}
			}

			// 2) Scan all anchors in the row looking for ann_id or announcements path
			var chosenHref, chosenTitle string
			s.Find("a[href]").EachWithBreak(func(i int, a *goquery.Selection) bool {
				h, _ := a.Attr("href")
				if h == "" {
					return true
				}
				if strings.Contains(h, "ann_id=") || strings.Contains(h, "/announcements/") || strings.Contains(strings.ToLower(h), "announcement") {
					chosenHref = h
					chosenTitle = strings.TrimSpace(a.Text())
					return false
				}
				return true
			})
			if chosenHref != "" {
				return chosenTitle, absoluteURL(baseURL, chosenHref)
			}

			// 3) Fallback: choose the first anchor that is NOT a company-profile
			s.Find("a[href]").EachWithBreak(func(i int, a *goquery.Selection) bool {
				h, _ := a.Attr("href")
				if h == "" {
					return true
				}
				if strings.Contains(h, "company-profile") || strings.Contains(h, "/company/") {
					return true
				}
				chosenHref = h
				chosenTitle = strings.TrimSpace(a.Text())
				return false
			})
			if chosenHref != "" {
				return chosenTitle, absoluteURL(baseURL, chosenHref)
			}

			// 4) Final fallback: text from second column, no href
			title := strings.TrimSpace(s.Find("td:nth-child(2)").Text())
			return title, ""
		}

		doc.Find("table tbody tr").Each(func(i int, s *goquery.Selection) {
			// Extract date (first cell) defensively
			date := strings.TrimSpace(s.Find("td:nth-child(1)").Text())

			// Find the correct announcement anchor (title + link)
			title, link := findAnnouncementAnchor(s, current)

			// If we didn't find a link inside the row, try to pick any href in the row as a last resort
			if link == "" {
				if a := s.Find("a[href]").First(); a.Length() > 0 {
					if h, ok := a.Attr("href"); ok {
						link = absoluteURL(current, h)
						if title == "" {
							title = strings.TrimSpace(a.Text())
						}
					}
				}
			}

			// If still empty, skip
			if title == "" && link == "" {
				return
			}

			ann := Announcement{Title: title, Link: link, DateRaw: date}
			if err := saveAnnouncement(db, &ann); err != nil {
				log.Printf("save announcement error: %v", err)
			}
		})

		nextHref, ok := findNextLink(doc)
		if !ok {
			break
		}
		current = absoluteURL(current, nextHref)
		time.Sleep(time.Duration(400+rand.Intn(1200)) * time.Millisecond)
	}

	rows, err := db.Query("SELECT title, link, date_raw, pdf_path FROM announcements ORDER BY seen_at DESC LIMIT 50")
	if err == nil {
		defer rows.Close()
		var out []Announcement
		for rows.Next() {
			var a Announcement
			var pdfPath sql.NullString
			rows.Scan(&a.Title, &a.Link, &a.DateRaw, &pdfPath)
			out = append(out, a)
		}
		b, _ := json.MarshalIndent(out, "", "  ")
		fmt.Println(string(b))
	}
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
	Title       string    `json:"title"`
	Link        string    `json:"link"`
	DateRaw     string    `json:"date_raw"`
	ParsedDate  time.Time `json:"parsed_date,omitempty"`
	Description string    `json:"description,omitempty"`
}

const schema = `
CREATE TABLE IF NOT EXISTS announcements (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT,
    link TEXT UNIQUE,
    date_raw TEXT,
    parsed_date DATETIME,
    description TEXT,
    downloaded_at DATETIME,
    pdf_path TEXT,
    pdf_checksum TEXT,
    seen_at DATETIME
);
CREATE INDEX IF NOT EXISTS idx_ann_seen_at ON announcements(seen_at);
`

func saveAnnouncement(db *sql.DB, a *Announcement) error {
	now := time.Now().UTC()
	_, err := db.Exec(`INSERT INTO announcements(title, link, date_raw, parsed_date, description, seen_at) VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT(link) DO UPDATE SET title=excluded.title, date_raw=excluded.date_raw, parsed_date=excluded.parsed_date, description=excluded.description, seen_at=?;`,
		a.Title, a.Link, a.DateRaw, nil, a.Description, now, now)
	return err
}

func findNextLink(doc *goquery.Document) (string, bool) {
	candidates := []string{
		`a[rel="next"]`,
		`a.next`,
		`ul.pagination a.next`,
		`nav a.next`,
	}
	for _, sel := range candidates {
		if s := doc.Find(sel); s.Length() > 0 {
			if href, ok := s.First().Attr("href"); ok && strings.TrimSpace(href) != "" {
				return href, true
			}
		}
	}

	var found string
	doc.Find("a").EachWithBreak(func(i int, s *goquery.Selection) bool {
		text := strings.TrimSpace(s.Text())
		if text == "Next" || text == "Next >" || text == ">" || text == ">>" || text == "›" || strings.Contains(strings.ToLower(text), "next") {
			if href, ok := s.Attr("href"); ok {
				found = href
				return false
			}
		}
		return true
	})
	if found != "" {
		return found, true
	}
	return "", false
}
