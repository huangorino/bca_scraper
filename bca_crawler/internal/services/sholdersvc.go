package services

import (
	"fmt"
	"strings"
	"time"

	"bca_crawler/internal/models"
	"bca_crawler/internal/utils"

	"github.com/PuerkitoBio/goquery"
)

func parseDirectorChange219(doc *goquery.Document, ann *models.Announcement) ([]*models.ShareholdingChange, error) {
	var results []*models.ShareholdingChange

	table := doc.Find("table.ven_table")

	table.Find("tr").Each(func(i int, tr *goquery.Selection) {
		cells := tr.Find("td")

		if cells.Length() < 5 {
			return
		}

		if _, ok := cells.First().Attr("rowspan"); !ok {
			return
		}

		change := newBaseChange(ann)
		change.ChangeType = utils.PtrString("Changes in Director's Interest Pursuant")

		change.DateOfChange = parseDate(cells.Eq(1).Text())
		change.SecuritiesChanged = parseInt(cells.Eq(2).Text())
		change.TransactionType = cleanText(cells.Eq(3).Text())
		change.NatureOfInterest = cleanText(cells.Eq(4).Text())

		next1 := tr.Next()
		next2 := next1.Next()
		next3 := next2.Next()

		change.RegisteredHolder = cleanText(next1.Find("td").Last().Text())
		change.TransactionDesc = cleanText(next2.Find("td").Last().Text())
		change.Consideration = cleanText(next3.Find("td").Last().Text())

		results = append(results, change)
	})

	extractTotals(doc, results)

	return results, nil
}

func parseDirectorChange135(doc *goquery.Document, ann *models.Announcement) ([]*models.ShareholdingChange, error) {
	var results []*models.ShareholdingChange

	table := doc.Find("table.ven_table")

	table.Find("tr").Each(func(i int, tr *goquery.Selection) {
		cells := tr.Find("td")

		if cells.Length() < 4 {
			return
		}

		date := parseDate(cells.Eq(1).Text())
		if date == nil {
			return
		}

		change := newBaseChange(ann)
		change.ChangeType = utils.PtrString("Changes in Director's Interest Pursuant")

		change.TransactionType = cleanText(cells.Eq(0).Text())
		change.DateOfChange = date
		change.SecuritiesChanged = parseInt(cells.Eq(2).Text())
		change.PriceTransacted = parseDecimal(cells.Eq(3).Text())

		results = append(results, change)
	})

	extractTotals(doc, results)

	return results, nil
}

func parseChangesInSub138(doc *goquery.Document, ann *models.Announcement) ([]*models.ShareholdingChange, error) {
	var results []*models.ShareholdingChange

	table := doc.Find("table.ven_table")

	table.Find("tr").Each(func(i int, tr *goquery.Selection) {
		cells := tr.Find("td")

		if cells.Length() < 5 {
			return
		}

		if _, ok := cells.First().Attr("rowspan"); !ok {
			return
		}

		change := newBaseChange(ann)
		change.ChangeType = utils.PtrString("Changes in Substantial Shareholder's Interest Pursuant")

		change.DateOfChange = parseDate(cells.Eq(1).Text())
		change.SecuritiesChanged = parseInt(cells.Eq(2).Text())
		change.TransactionType = cleanText(cells.Eq(3).Text())
		change.NatureOfInterest = cleanText(cells.Eq(4).Text())

		next1 := tr.Next()
		next2 := next1.Next()

		change.RegisteredHolder = cleanText(next1.Find("td").Last().Text())
		change.TransactionDesc = cleanText(next2.Find("td").Last().Text())

		results = append(results, change)
	})

	extractTotals(doc, results)

	return results, nil
}

func parseChangesInSub29B(doc *goquery.Document, ann *models.Announcement) ([]*models.ShareholdingChange, error) {
	var results []*models.ShareholdingChange

	table := doc.Find("table.ven_table")

	table.Find("tr").Each(func(i int, tr *goquery.Selection) {
		cells := tr.Find("td")

		if cells.Length() < 4 {
			return
		}

		// Old format (5 cols): TransactionType | Description | Date | Securities | Price
		// New format (4 cols): TransactionType | Date | Securities | Price
		// Detect by trying the date column for each format.
		var desc, dateIdx, securitiesIdx, priceIdx int
		if cells.Length() >= 5 && parseDate(cells.Eq(2).Text()) != nil {
			desc, dateIdx, securitiesIdx, priceIdx = 1, 2, 3, 4
		} else {
			dateIdx, securitiesIdx, priceIdx = 1, 2, 3
		}

		date := parseDate(cells.Eq(dateIdx).Text())
		if date == nil {
			return
		}

		change := newBaseChange(ann)
		change.ChangeType = utils.PtrString("Changes in Substantial Shareholder's Interest Pursuant")

		change.TransactionType = cleanText(cells.Eq(0).Text())
		change.TransactionDesc = cleanText(cells.Eq(desc).Text())
		change.DateOfChange = date
		change.SecuritiesChanged = parseInt(cells.Eq(securitiesIdx).Text())
		change.PriceTransacted = parseDecimal(cells.Eq(priceIdx).Text())

		results = append(results, change)
	})

	extractTotals(doc, results)

	return results, nil
}

func parseNoticeInterest(doc *goquery.Document, ann *models.Announcement) ([]*models.ShareholdingChange, error) {
	change := newBaseChange(ann)
	change.ChangeType = utils.PtrString("Notice of Interest of Substantial Shareholders Pursuant")

	extractTotals(doc, []*models.ShareholdingChange{change})

	return []*models.ShareholdingChange{change}, nil
}

func parseNoticeCeasing(doc *goquery.Document, ann *models.Announcement) ([]*models.ShareholdingChange, error) {
	change := newBaseChange(ann)
	change.ChangeType = utils.PtrString("Notice of Person Ceasing Substantial Shareholders Pursuant")

	extractTotals(doc, []*models.ShareholdingChange{change})

	return []*models.ShareholdingChange{change}, nil
}

func findField(doc *goquery.Document, label string) string {
	var result string

	doc.Find("td.formContentLabel").EachWithBreak(
		func(i int, s *goquery.Selection) bool {

			if strings.Contains(strings.ToLower(s.Text()), strings.ToLower(label)) {
				result = utils.CleanString(s.Next().Text())
				return false
			}

			return true
		},
	)

	return result
}

func extractTotals(doc *goquery.Document, changes []*models.ShareholdingChange) {
	var securitiesLabels = []string{
		"Securities disposed",
		"No of securities",
		"No. of securities",
		"Number of securities",
	}

	for _, c := range changes {
		c.PersonName = utils.PtrString(findField(doc, "Name"))
		c.PersonAddress = utils.PtrString(findField(doc, "Address"))
		c.CompanyNo = utils.PtrString(findField(doc, "Company No"))
		c.PersonNationality = utils.PtrString(findField(doc, "Nationality"))
		c.SecurityDescription = utils.PtrString(findField(doc, "Descriptions"))
		c.RegisteredHolder = utils.PtrString(findField(doc, "Name of registered holder"))
		c.RegisteredHolderAddress = utils.PtrString(findField(doc, "Address of registered holder"))

		c.DateInterestAcquired = utils.ParseDate(findField(doc, "Date interest acquired"))
		c.DateOfCessation = utils.ParseDate(findField(doc, "Date of cessation"))
		c.Currency = utils.PtrString(findField(doc, "Currency"))
		c.PriceTransacted = utils.ParseFloat(findField(doc, "Price Transacted"))
		c.Circumstances = utils.PtrString(findField(doc, "Circumstances"))
		c.NatureOfInterest = utils.PtrString(findField(doc, "Nature of interest"))
		c.DateOfNotice = utils.ParseDate(findField(doc, "Date of notice"))
		c.DateNoticeReceived = utils.ParseDate(findField(doc, "Date notice received"))

		if c.SecuritiesChanged == nil {
			for _, label := range securitiesLabels {
				if v := utils.ParseInt64(findField(doc, label)); v != nil {
					c.SecuritiesChanged = v
					break
				}
			}
		}

		if c.Consideration == nil {
			c.Consideration = utils.PtrString(findField(doc, "Consideration"))
		}

		c.DirectUnits = utils.ParseInt64(findField(doc, "Direct (units)"))
		c.DirectPercent = utils.ParseFloat(findField(doc, "Direct (%)"))
		c.IndirectUnits = utils.ParseInt64(findField(doc, "Indirect/deemed interest (units)"))
		c.IndirectPercent = utils.ParseFloat(findField(doc, "Indirect/deemed interest (%)"))
		c.TotalSecurities = utils.ParseInt64(findField(doc, "Total no of securities after change"))
		c.DateOfNotice = utils.ParseDate(findField(doc, "Date of notice"))
		c.DateNoticeReceived = utils.ParseDate(findField(doc, "Date notice received"))

		c.Remarks = extractRemarks(doc)
	}
}

func extractRemarks(doc *goquery.Document) *string {
	remarks := ""

	// modern layout
	remarks = strings.TrimSpace(
		doc.Find("#divRemarks td.FootNote").Text(),
	)

	if remarks != "" {
		return cleanText(remarks)
	}

	// fallback for some announcements using <pre>
	remarks = strings.TrimSpace(
		doc.Find("#divRemarks pre").Text(),
	)

	return cleanText(remarks)
}

func newBaseChange(ann *models.Announcement) *models.ShareholdingChange {
	return &models.ShareholdingChange{
		AnnID:       ann.AnnID,
		StockCode:   ann.StockName,
		CompanyName: &ann.CompanyName,
	}
}

type AnnouncementType int

const (
	TypeUnknown AnnouncementType = iota
	TypeDirector135
	TypeDirector219
	TypeChangesInSub138
	TypeChangesInSub29B
	TypeNoticeInterest
	TypeNoticeCeasing
)

func detectAnnouncementType(doc *goquery.Document) AnnouncementType {

	title := strings.ToLower(strings.TrimSpace(doc.Find("h3").First().Text()))

	switch {

	case strings.Contains(title, "director") && strings.Contains(title, "219"):
		return TypeDirector219

	case strings.Contains(title, "director") && strings.Contains(title, "135"):
		return TypeDirector135

	case strings.Contains(title, "changes in sub") && strings.Contains(title, "29b"):
		return TypeChangesInSub29B

	case strings.Contains(title, "changes in sub") && strings.Contains(title, "138"):
		return TypeChangesInSub138

	case strings.Contains(title, "notice of interest"):
		return TypeNoticeInterest

	case strings.Contains(title, "notice of person ceasing"):
		return TypeNoticeCeasing
	}

	return TypeUnknown
}

func ParseShareholdingChange(ann *models.Announcement) ([]*models.ShareholdingChange, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(ann.Content))
	if err != nil {
		return nil, err
	}

	switch detectAnnouncementType(doc) {

	case TypeDirector135:
		return parseDirectorChange135(doc, ann)

	case TypeDirector219:
		return parseDirectorChange219(doc, ann)

	case TypeChangesInSub138:
		return parseChangesInSub138(doc, ann)

	case TypeChangesInSub29B:
		return parseChangesInSub29B(doc, ann)

	case TypeNoticeInterest:
		return parseNoticeInterest(doc, ann)

	case TypeNoticeCeasing:
		return parseNoticeCeasing(doc, ann)
	}

	return nil, fmt.Errorf("unsupported announcement type")
}

func parseDate(s string) *time.Time  { return utils.ParseDate(s) }
func parseInt(s string) *int64       { return utils.ParseInt64(s) }
func parseDecimal(s string) *float64 { return utils.ParseFloat(s) }
func cleanText(s string) *string     { return utils.PtrString(utils.CleanString(s)) }
