package services

import (
	"fmt"
	"strings"
	"time"

	"bca_crawler/internal/models"
	"bca_crawler/internal/utils"

	"github.com/PuerkitoBio/goquery"
)

func parseDirectorChange219(
	doc *goquery.Document,
	ann *models.Announcement,
) ([]*models.ShareholdingChange, error) {

	var results []*models.ShareholdingChange

	person := findField(doc, "Name")

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

		change.PersonName = utils.PtrString(person)
		change.DateOfChange = parseDate(cells.Eq(1).Text())
		change.SecuritiesChanged = parseInt(cells.Eq(2).Text())
		change.TransactionType = cleanText(cells.Eq(3).Text())
		change.NatureOfInterest = cleanText(cells.Eq(4).Text())

		next1 := tr.Next()
		next2 := next1.Next()
		next3 := next2.Next()

		change.RegisteredHolder =
			cleanText(next1.Find("td").Last().Text())

		change.TransactionDesc =
			cleanText(next2.Find("td").Last().Text())

		change.PriceTransacted =
			parseDecimal(next3.Find("td").Last().Text())

		results = append(results, change)
	})

	extractTotals(doc, results)

	return results, nil
}

func parseDirectorChange135(
	doc *goquery.Document,
	ann *models.Announcement,
) ([]*models.ShareholdingChange, error) {

	var results []*models.ShareholdingChange

	person := findField(doc, "Name")

	doc.Find("table.ven_table tr").Each(func(i int, tr *goquery.Selection) {

		cells := tr.Find("td")

		if cells.Length() < 4 {
			return
		}

		date := parseDate(cells.Eq(1).Text())
		if date == nil {
			return
		}

		change := newBaseChange(ann)

		change.PersonName = utils.PtrString(person)
		change.TransactionType = cleanText(cells.Eq(0).Text())
		change.DateOfChange = date
		change.SecuritiesChanged = parseInt(cells.Eq(2).Text())
		change.PriceTransacted = parseDecimal(cells.Eq(3).Text())

		results = append(results, change)
	})

	extractTotals(doc, results)

	return results, nil
}

func parseSubstantialChange(doc *goquery.Document, ann *models.Announcement) ([]*models.ShareholdingChange, error) {
	var results []*models.ShareholdingChange

	person := findField(doc, "Name")
	address := findField(doc, "Address")
	nationality := findField(doc, "Nationality")
	companyNo := findField(doc, "Company No.")
	securityDesc := findField(doc, "Descriptions")

	table := doc.Find("table.ven_table")

	rows := table.Find("tr")

	rows.Each(func(i int, tr *goquery.Selection) {

		cells := tr.Find("td")

		if cells.Length() < 5 {
			return
		}

		// Only parse rows with rowspan (start of transaction block)
		if _, ok := cells.First().Attr("rowspan"); !ok {
			return
		}

		change := newBaseChange(ann)

		change.PersonName = utils.PtrString(person)
		change.PersonAddress = utils.PtrString(address)
		change.PersonNationality = utils.PtrString(nationality)
		change.CompanyNo = utils.PtrString(companyNo)
		change.SecurityDescription = utils.PtrString(securityDesc)

		change.DateOfChange = utils.ParseDate(cells.Eq(1).Text())
		change.SecuritiesChanged = utils.ParseInt64(cells.Eq(2).Text())
		change.TransactionType = utils.PtrString(utils.CleanString(cells.Eq(3).Text()))
		change.NatureOfInterest = utils.PtrString(utils.CleanString(cells.Eq(4).Text()))

		// ---- registered holder rows ----

		next1 := tr.Next()
		next2 := next1.Next()
		next3 := next2.Next()

		change.RegisteredHolder = utils.PtrString(utils.CleanString(
			next1.Find("td").Last().Text(),
		))

		change.RegisteredHolderAddress = utils.PtrString(utils.CleanString(
			next2.Find("td").Last().Text(),
		))

		change.TransactionDesc = utils.PtrString(utils.CleanString(
			next3.Find("td").Last().Text(),
		))

		results = append(results, change)
	})

	extractTotals(doc, results)

	return results, nil
}

func parseNoticeInterest(doc *goquery.Document, ann *models.Announcement) ([]*models.ShareholdingChange, error) {
	change := newBaseChange(ann)

	change.PersonName = utils.PtrString(findField(doc, "Name"))
	change.PersonAddress = utils.PtrString(findField(doc, "Address"))
	change.PersonNationality = utils.PtrString(findField(doc, "Nationality"))
	change.CompanyNo = utils.PtrString(findField(doc, "NRIC"))

	change.DateInterestAcquired = utils.ParseDate(findField(doc, "Date interest acquired"))

	change.SecuritiesChanged = utils.ParseInt64(findField(doc, "No of securities"))

	change.PriceTransacted = utils.ParseFloat(findField(doc, "Price Transacted"))

	change.NatureOfInterest = utils.PtrString(findField(doc, "Nature of interest"))

	extractTotals(doc, []*models.ShareholdingChange{change})

	return []*models.ShareholdingChange{change}, nil
}

func parseNoticeCeasing(doc *goquery.Document, ann *models.Announcement) ([]*models.ShareholdingChange, error) {
	change := newBaseChange(ann)

	change.PersonName = utils.PtrString(findField(doc, "Name"))
	change.PersonAddress = utils.PtrString(findField(doc, "Address"))
	change.PersonNationality = utils.PtrString(findField(doc, "Nationality"))

	change.DateOfCessation = utils.ParseDate(findField(doc, "Date of cessation"))

	change.SecuritiesChanged = utils.ParseInt64(findField(doc, "No of securities disposed"))

	change.PriceTransacted = utils.ParseFloat(findField(doc, "Price Transacted"))

	change.NatureOfInterest = utils.PtrString(findField(doc, "Nature of interest"))

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
	directUnits := utils.ParseInt64(findField(doc, "Direct (units)"))
	directPct := utils.ParseFloat(findField(doc, "Direct (%)"))

	indirectUnits := utils.ParseInt64(findField(doc, "Indirect/deemed interest (units)"))
	indirectPct := utils.ParseFloat(findField(doc, "Indirect/deemed interest (%)"))

	total := utils.ParseInt64(findField(doc, "Total no of securities after change"))

	notice := utils.ParseDate(findField(doc, "Date of notice"))
	received := utils.ParseDate(findField(doc, "Date notice received"))

	for _, c := range changes {

		c.DirectUnits = directUnits
		c.DirectPercent = directPct
		c.IndirectUnits = indirectUnits
		c.IndirectPercent = indirectPct
		c.TotalSecurities = total
		c.DateOfNotice = notice
		c.DateNoticeReceived = received
	}
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
	TypeSubstantialChange
	TypeNoticeInterest
	TypeNoticeCeasing
)

func detectAnnouncementType(doc *goquery.Document) AnnouncementType {

	title := strings.ToLower(strings.TrimSpace(doc.Find("h3").First().Text()))

	switch {

	case strings.Contains(title, "director") &&
		strings.Contains(title, "219"):
		return TypeDirector219

	case strings.Contains(title, "director"):
		return TypeDirector135

	case strings.Contains(title, "sub. s-hldr"):
		return TypeSubstantialChange

	case strings.Contains(title, "notice of interest"):
		return TypeNoticeInterest

	case strings.Contains(title, "notice of person ceasing"):
		return TypeNoticeCeasing
	}

	return TypeUnknown
}

func ParseShareholdingChange(
	ann *models.Announcement,
) ([]*models.ShareholdingChange, error) {

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(ann.Content))
	if err != nil {
		return nil, err
	}

	switch detectAnnouncementType(doc) {

	case TypeDirector135:
		return parseDirectorChange135(doc, ann)

	case TypeDirector219:
		return parseDirectorChange219(doc, ann)

	case TypeSubstantialChange:
		return parseSubstantialChange(doc, ann)

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
