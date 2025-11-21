package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	_ "modernc.org/sqlite"

	"bca_crawler/internal/models"
)

// SaveAnnouncement inserts or updates a full announcement
func SaveAnnouncement(db *sqlx.DB, a *models.Announcement) error {
	now := time.Now().UTC()

	attachmentsJSON, err := json.Marshal(a.Attachments)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
	INSERT INTO announcements(
		ann_id, title, link, company_name, stock_name, date_posted, category, ref_number, content, attachments)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
	ON CONFLICT(ann_id)
	DO UPDATE SET
		title = EXCLUDED.title,
		link = EXCLUDED.link,
		company_name = EXCLUDED.company_name,
		stock_name = EXCLUDED.stock_name,
		category = EXCLUDED.category,
		ref_number = EXCLUDED.ref_number,
		content = EXCLUDED.content,
		attachments = EXCLUDED.attachments;`,
		a.AnnID, a.Title, a.Link, a.CompanyName, a.StockName, now, a.Category, a.RefNumber, a.Content, attachmentsJSON)

	return err
}

// UpdateAnnouncementInfo updates parsed fields after HTML parsing
func UpdateAnnouncement(db *sqlx.DB, a *models.Announcement) error {
	attachmentsJSON, err := json.Marshal(a.Attachments)
	if err != nil {
		return err
	}

	_, err = db.Exec(`
	INSERT INTO announcements (
		ann_id, company_name, stock_name, date_posted, category, ref_number, attachments, content
	) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	ON CONFLICT(ann_id)
	DO UPDATE SET
		company_name = EXCLUDED.company_name,
		stock_name = EXCLUDED.stock_name,
		date_posted = EXCLUDED.date_posted,
		category = EXCLUDED.category,
		ref_number = EXCLUDED.ref_number,
		attachments = EXCLUDED.attachments,
		content = EXCLUDED.content;`,
		a.AnnID, a.CompanyName, a.StockName, a.DatePosted, a.Category, a.RefNumber, attachmentsJSON, a.Content)
	return err
}

func FetchUnparsedAnnouncements(db *sqlx.DB) ([]*models.Announcement, error) {
	rows, err := db.Query(`
	SELECT id, ann_id, content 
	FROM announcements 
	WHERE ref_number = '' ORDER BY ann_id ASC`)

	if err != nil {
		return nil, fmt.Errorf("query announcements: %w", err)
	}
	defer rows.Close()

	var announcements []*models.Announcement
	for rows.Next() {
		var ann models.Announcement
		if err := rows.Scan(&ann.ID, &ann.AnnID, &ann.Content); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		announcements = append(announcements, &ann)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("rows error: %w", err)
	}

	return announcements, nil
}

func SaveStock(db *sqlx.DB, stock *models.Stock) error {
	_, err := db.Exec(`
	INSERT INTO stocks(
		stock_code, stock_name, market, status)
	VALUES ($1, $2, $3, $4)
	ON CONFLICT(stock_code)
	DO UPDATE SET
		stock_name = EXCLUDED.stock_name,
		market = EXCLUDED.market,
		status = EXCLUDED.status;`,
		stock.StockCode, stock.Name, stock.Market, stock.Status)
	return err
}

func SaveEntity(db *sqlx.DB, e *models.Entity) error {
	_, err := db.Exec(`
	INSERT INTO entities(
		type, name, stock_code, ic_number, age, gender, nationality, created_at)
	VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	ON CONFLICT(name, IFNULL(ic_number, ''))
	DO UPDATE SET
		age = EXCLUDED.age,
		gender = EXCLUDED.gender,
		nationality = EXCLUDED.nationality,
		updated_at = DATETIME('now');`,
		e.Type, e.Name, e.StockCode, e.ICNumber, e.Age, e.Gender, e.Nationality, e.CreatedAt)

	return err
}

func GetMaxAnnID(db *sqlx.DB) (int, error) {
	var maxID int
	err := db.QueryRow(`SELECT MAX(ann_id) FROM announcements`).Scan(&maxID)
	if err != nil {
		return 0, fmt.Errorf("query max ann_id: %w", err)
	}
	return maxID, nil
}

func FetchAnnouncementsByCategory(db *sqlx.DB, category string) ([]*models.Announcement, error) {
	sqlQuery := `SELECT id, ann_id,
      link, company_name, stock_name,
      date_posted, category, ref_number,
      attachments, content 
      FROM announcements`

	var args []interface{}
	if category != "" {
		if category == "attachments" {
			sqlQuery += " WHERE attachments != 'null'"
		} else {
			sqlQuery += " WHERE category = $1"
			args = append(args, category)
		}
	}

	sqlQuery += " ORDER BY ann_id ASC"

	var announcements []models.AnnouncementDB
	err := db.Select(&announcements, sqlQuery, args...)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("sqlx select: %w", err)
	}

	var result []*models.Announcement
	for i := range announcements {
		ann, err := ConvertAnnouncementDBToAnnouncement(announcements[i])
		if err != nil {
			return nil, fmt.Errorf("convert announcement db to announcement: %w", err)
		}

		result = append(result, ann)
	}

	return result, nil
}

func ConvertAnnouncementDBToAnnouncement(a models.AnnouncementDB) (*models.Announcement, error) {
	attachments := []string{}
	if err := json.Unmarshal([]byte(a.Attachments.String), &attachments); err != nil {
		return nil, fmt.Errorf("json unmarshal: %w", err)
	}

	return &models.Announcement{
		ID:          a.ID,
		AnnID:       a.AnnID,
		Title:       a.Title.String,
		Link:        a.Link.String,
		CompanyName: a.CompanyName.String,
		StockName:   a.StockName.String,
		DatePosted:  a.DatePosted.String,
		Category:    a.Category.String,
		RefNumber:   a.RefNumber.String,
		Content:     a.Content.String,
		Attachments: attachments,
	}, nil
}
