package services

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"bca_crawler/internal/models"
	"bca_crawler/internal/utils"

	"github.com/PuerkitoBio/goquery"
)

type qualification struct {
	Level          string
	FieldOfStudy   string
	Institute      string
	AdditionalInfo string
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
		return fmt.Errorf("[Error] parse announcement info HTML: %w", err)
	}

	title := strings.TrimSpace(doc.Find("h3").First().Text())
	if title != "" {
		ann.Title = utils.CleanString(title)
	}

	// --------------------------------------------
	// 1. Extract Announcement Info section
	// --------------------------------------------
	section, err := extractAnnouncementInfoHTML(ann.Content)
	if err != nil {
		return fmt.Errorf("[Error] announcement info section not found")
	}

	doc, err = goquery.NewDocumentFromReader(strings.NewReader(section))
	if err != nil {
		return fmt.Errorf("[Error] parse announcement info HTML: %w", err)
	}

	// --------------------------------------------
	// 2. Parse Announcement Info fields
	// --------------------------------------------
	doc.Find("tr").Each(func(_ int, tr *goquery.Selection) {
		tds := tr.Find("td")
		if tds.Length() < 2 {
			return
		}

		label := utils.CleanString(tds.Eq(0).Text())
		value := utils.CleanString(tds.Eq(1).Text())

		switch strings.ToLower(label) {
		case "company name":
			ann.CompanyName = value
		case "stock name":
			ann.StockName = value
		case "date announced":
			ann.DatePosted = safeTimeValue(utils.ParseDate(value))
		case "category":
			ann.Category = value
		case "reference number", "reference no":
			ann.RefNumber = value
		}
	})

	// --------------------------------------------
	// 3. Parse attachment URLs (if any)
	// --------------------------------------------
	var attachments []string

	doc.Find("p.att_download_pdf a").Each(func(_ int, s *goquery.Selection) {
		href, ok := s.Attr("href")
		if !ok {
			return
		}

		href = strings.TrimSpace(utils.HtmlUnescape(href))
		if href == "" {
			return
		}

		attachments = append(attachments, href)
	})

	if len(attachments) > 0 {
		ann.Attachments = attachments
	}

	return nil
}

func extractAnnouncementInfoHTML(html string) (string, error) {
	start := strings.Index(html, `<div class="ven_announcement_info"`)
	if start == -1 {
		return "", fmt.Errorf("announcement info start not found")
	}

	slice := html[start:]

	depth := 0
	pos := 0

	for {
		openIdx := strings.Index(slice[pos:], "<div")
		closeIdx := strings.Index(slice[pos:], "</div>")

		if openIdx == -1 && closeIdx == -1 {
			break
		}

		if openIdx != -1 && (openIdx < closeIdx || closeIdx == -1) {
			depth++
			pos += openIdx + 4
		} else {
			depth--
			pos += closeIdx + 6
			if depth == 0 {
				return slice[:pos], nil
			}
		}
	}

	return "", fmt.Errorf("announcement info div not closed properly")
}

func ParseBoardroomChangeHTML(ann *models.Announcement) (*models.BoardroomChange, *models.Entity, *models.Entity, *models.Background, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(ann.Content))
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("[Error] parse HTML: %w", err)
	}

	// detect if it's old-style Bursa HTML (pre-2015)
	isOld := doc.Find("table.InputTable2").Length() > 0 ||
		strings.Contains(strings.ToLower(doc.Text()), "previous position") ||
		strings.Contains(strings.ToLower(doc.Text()), "remarks :")

	// --- Boardroom change core fields ---
	change := &models.BoardroomChange{
		AnnID:         ann.AnnID,
		Category:      tidy(findValueByLabel(doc, "Category")),
		DateAnnounced: utils.ParseDate(findValueByLabel(doc, "Date Announced")),
		DateOfChange:  utils.ParseDate(findValueByLabel(doc, "Date of change")),
		Directorate:   tidy(findValueByLabel(doc, "Directorate")),
		TypeOfChange:  tidy(findValueByLabel(doc, "Type of change")),
	}

	if change.DateAnnounced == nil {
		change.DateAnnounced = change.DateOfChange
	}

	change.Designation = tidy(findValueByLabel(doc, "New Position"))
	change.PreviousPosition = tidy(findValueByLabel(doc, "Previous Position"))

	if change.Designation == "" {
		if change.TypeOfChange == "Appointment" {
			change.Designation = tidy(findValueByLabel(doc, "Designation"))
		} else {
			change.PreviousPosition = tidy(findValueByLabel(doc, "Designation"))
		}
	}

	change.Remarks = tidy(findValueByLabel(doc, "Remarks :"))
	if change.Remarks == "" {
		// try remarks from InputTable2
		change.Remarks = tidy(doc.Find("table.InputTable2 td.FootNote").Text())
	}

	// --- Company fields ---
	companyName := tidy(findValueByLabel(doc, "Company Name"))
	if companyName == "" {
		companyName = tidy(doc.Find("td.company_name").First().Text())
	}
	stockCode := tidy(findValueByLabel(doc, "Stock Name"))
	if stockCode == "" {
		stockCode = tidy(findValueByLabel(doc, "Stock Code"))
	}

	company := &models.Entity{
		Type:      strings.ToUpper("company"),
		Name:      strings.ToUpper(companyName),
		StockCode: strings.ToUpper(stockCode),
		CreatedAt: safeTimeValue(change.DateAnnounced),
	}

	// --- Person fields ---
	personName := tidy(findValueByLabel(doc, "Name"))
	// name, title := utils.TrimAbbreviation(personName)
	ageStr := tidy(findValueByLabel(doc, "Age"))
	age, _ := strconv.Atoi(ageStr)
	birthYear := change.DateAnnounced.Year() - age
	gender := tidy(findValueByLabel(doc, "Gender"))
	nationality := tidy(findValueByLabel(doc, "Nationality"))

	// Extract first character of gender, safely
	genderCode := ""
	if len(gender) > 0 {
		genderCode = strings.ToUpper(gender[:1])
	}

	person := &models.Entity{
		Type:        strings.ToUpper("person"),
		Name:        strings.ToUpper(personName),
		Title:       strings.ToUpper(""),
		StockCode:   strings.ToUpper(stockCode),
		BirthYear:   birthYear,
		Gender:      genderCode,
		Nationality: strings.ToUpper(nationality),
		CreatedAt:   safeTimeValue(change.DateAnnounced),
	}

	background := &models.Background{
		WorkingExperience:    tidy(findValueByLabel(doc, "Working experience and occupation")),
		Directorships:        tidy(findValueByLabel(doc, "Directorships in public companies and listed issuers (if any)")),
		FamilyRelationship:   tidy(findValueByLabel(doc, "Family relationship with any director and/or major shareholder of the listed issuer")),
		ConflictOfInterest:   tidy(findValueByLabel(doc, "Any conflict of interests that he/she has with the listed issuer")),
		InterestInSecurities: tidy(findValueByLabel(doc, "Details of any interest in the securities of the listed issuer or its subsidiaries")),
	}

	// handle alternate labels found in old HTML
	if background.Directorships == "" {
		background.Directorships = tidy(findValueByLabel(doc, "Directorship of public companies (if any)"))
	}

	if isOld {
		qInline := tidy(findValueByLabel(doc, "Qualifications"))
		if qInline != "" {
			// parsed := parseInlineQualification(qInline)
			background.Qualification = qInline
		}
	} else {
		background.Qualification = extractQualifications(doc)
	}

	return change, company, person, background, nil
}

func findValueByLabel(doc *goquery.Document, label string) string {
	var value string
	doc.Find("table").EachWithBreak(func(i int, s *goquery.Selection) bool {
		s.Find("tr").EachWithBreak(func(j int, tr *goquery.Selection) bool {
			tds := tr.Find("td")
			if tds.Length() >= 2 {
				labelText := utils.CleanString(tds.Eq(0).Text())
				if strings.EqualFold(labelText, label) {
					value = utils.CleanString(tds.Eq(1).Text())
					return false // break row loop
				}
			}
			return true
		})
		if value != "" {
			return false // break table loop
		}
		return true
	})

	if strings.ToLower(value) == "nil" {
		return ""
	}

	return value
}

func tidy(s string) string {
	return utils.CleanString(s)
}

// safeTimeValue safely dereferences a *time.Time pointer.
// If the pointer is nil, returns time.Now() as a fallback.
func safeTimeValue(t *time.Time) time.Time {
	if t == nil {
		return time.Now()
	}
	return *t
}

// detectLevel guesses the degree level from inline qualification text
func detectLevel(text string) string {
	t := strings.ToLower(text)
	switch {
	case strings.Contains(t, "master"):
		return "Masters"
	case strings.Contains(t, "bachelor"):
		return "Degree"
	case strings.Contains(t, "diploma"):
		return "Diploma"
	case strings.Contains(t, "phd"), strings.Contains(t, "doctor"):
		return "PhD"
	default:
		return "Other"
	}
}

func parseInlineQualification(text string) qualification {
	original := strings.TrimSpace(text)
	out := qualification{
		Level: detectLevel(original),
	}

	// Remove common prefix like "Holds a degree in " (case-insensitive)
	lower := strings.ToLower(original)
	prefix := "holds a degree in "
	if strings.HasPrefix(lower, prefix) {
		original = strings.TrimSpace(original[len(prefix):])
	}

	// Try to split by " from " to obtain institute.
	parts := strings.Split(original, " from the ")

	if len(parts) >= 2 {
		out.FieldOfStudy = strings.TrimSpace(parts[0])
		partsparts := strings.Split(parts[1], " in ")

		if len(partsparts) >= 2 {
			out.Institute = strings.TrimSpace(partsparts[0])
		} else {
			out.Institute = strings.TrimSpace(parts[1])
		}
	}

	return out
}

func extractQualifications(doc *goquery.Document) string {
	var q []qualification

	doc.Find("table").Each(func(i int, table *goquery.Selection) {
		table.Find("tr").Each(func(j int, tr *goquery.Selection) {
			if tr.Find(".formTableColumnHeader").Length() > 0 {
				return
			}
			tds := tr.Find("td")
			if tds.Length() < 4 {
				return
			}
			level := tidy(tds.Eq(1).Text())
			field := tidy(tds.Eq(2).Text())
			institute := tidy(tds.Eq(3).Text())
			add := ""
			if tds.Length() >= 5 {
				add = tidy(tds.Eq(4).Text())
			}
			if level == "" && field == "" && institute == "" {
				return
			}

			q = append(q, qualification{
				Level:          level,
				FieldOfStudy:   field,
				Institute:      institute,
				AdditionalInfo: add,
			})
		})
	})

	jsonBytes, err := json.Marshal(q)
	if err != nil {
		utils.Logger.Errorf("failed to marshal qualification to JSON: %v", err)
		return ""
	}

	return string(jsonBytes)
}

func MapToStockRows(body string) ([]models.StockRow, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(body))
	if err != nil {
		utils.Logger.Errorf("[Error] goquery parse error: %v", err)
		return nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	jsonStr := strings.TrimSpace(doc.Find("pre").Text())
	if jsonStr == "" {
		return nil, fmt.Errorf("JSON not found inside <pre> tag")
	}

	var result models.BursaStockResponse
	if err := json.Unmarshal([]byte(jsonStr), &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	if len(result.Data) == 0 {
		return []models.StockRow{}, nil
	}

	stocks := make([]models.StockRow, 0, len(result.Data))

	for i, row := range result.Data {
		if len(row) != 16 {
			utils.Logger.Warnf("[Warning] Row %d has unexpected column count: %d (expected 16)", i, len(row))
			continue
		}

		index, err := strconv.Atoi(fmt.Sprintf("%v", row[0]))
		if err != nil {
			utils.Logger.Warnf("[Warning] Row %d: invalid index value '%v': %v", i, row[0], err)
			index = 0
		}

		// Clean and validate stock name
		nameStr := fmt.Sprintf("%v", row[1])
		cleanName := strings.TrimSpace(strings.ReplaceAll(nameStr, "[S]", ""))
		if cleanName == "" {
			utils.Logger.Warnf("[Warning] Row %d: empty stock name", i)
		}

		stocks = append(stocks, models.StockRow{
			Index:         index,
			Name:          HTMLToText(cleanName),
			Code:          strings.TrimSpace(fmt.Sprintf("%v", row[2])),
			Market:        strings.TrimSpace(fmt.Sprintf("%v", row[3])),
			LastPrice:     strings.TrimSpace(fmt.Sprintf("%v", row[4])),
			ChangePrice:   strings.TrimSpace(fmt.Sprintf("%v", row[5])),
			ChangeValue:   HTMLToText(strings.TrimSpace(fmt.Sprintf("%v", row[6]))),
			ChangePercent: HTMLToText(strings.TrimSpace(fmt.Sprintf("%v", row[7]))),
			Volume:        strings.TrimSpace(fmt.Sprintf("%v", row[8])),
			Value:         strings.TrimSpace(fmt.Sprintf("%v", row[9])),
			Bid:           strings.TrimSpace(fmt.Sprintf("%v", row[10])),
			Ask:           strings.TrimSpace(fmt.Sprintf("%v", row[11])),
			BidVolume:     strings.TrimSpace(fmt.Sprintf("%v", row[12])),
			High:          strings.TrimSpace(fmt.Sprintf("%v", row[13])),
			Low:           strings.TrimSpace(fmt.Sprintf("%v", row[14])),
			Misc:          strings.TrimSpace(fmt.Sprintf("%v", row[15])),
		})
	}

	return stocks, nil
}

func HTMLToText(html string) string {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		utils.Logger.Warnf("[Warning] HTML parsing error in HTMLToText: %v", err)
		return strings.TrimSpace(html)
	}
	return strings.TrimSpace(doc.Text())
}
