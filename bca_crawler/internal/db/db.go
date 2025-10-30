package db

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "github.com/lib/pq"

	"bca_crawler/internal/models"
	"bca_crawler/internal/utils"
)

const schema = `
CREATE TABLE IF NOT EXISTS announcements (
	id SERIAL PRIMARY KEY,
	ann_id INTEGER UNIQUE,
	title TEXT,
	link TEXT UNIQUE,
	company_name TEXT,
	stock_name TEXT,
	date_posted TIMESTAMP,
	category TEXT,
	ref_number TEXT,
	content TEXT,
	attachments TEXT
);
CREATE INDEX IF NOT EXISTS idx_ann_date_posted ON announcements(date_posted);
`

// Setup initializes and verifies the database schema
func Setup(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr+"?sslmode=disable")
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, err
	}
	utils.Logger.Infof("Database initialized and schema verified (%s)", connStr)
	return db, nil
}

// Connect connects to an existing PostgreSQL database without altering schema
func Connect(connStr string) (*sql.DB, error) {
	db, err := sql.Open("postgres", connStr+"?sslmode=disable")
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	utils.Logger.Infof("Connected to database: %s", connStr)
	return db, nil
}

// SaveAnnouncement inserts or updates a full announcement
func SaveAnnouncement(db *sql.DB, a *models.Announcement) error {
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
func UpdateAnnouncement(db *sql.DB, a *models.Announcement) error {
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

func FetchUnparsedAnnouncements(db *sql.DB) ([]*models.Announcement, error) {
	rows, err := db.Query(`
	SELECT id, ann_id, content 
	FROM announcements 
	WHERE ref_number IS NULL ORDER BY ann_id ASC`)

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

func GetMaxAnnID(db *sql.DB) (int, error) {
	var maxID int
	err := db.QueryRow(`SELECT MAX(ann_id) FROM announcements`).Scan(&maxID)
	if err != nil {
		return 0, fmt.Errorf("query max ann_id: %w", err)
	}
	return maxID, nil
}
