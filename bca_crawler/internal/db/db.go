package db

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"bca_crawler/internal/models"
	"bca_crawler/internal/utils"
)

const schema = `
CREATE TABLE IF NOT EXISTS announcements (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	ann_id TEXT UNIQUE,
	title TEXT,
	link TEXT UNIQUE,
	company_name TEXT,
	stock_name TEXT,
	date_posted DATETIME,
	category TEXT,
	ref_number TEXT,
	content TEXT
);
CREATE INDEX IF NOT EXISTS idx_ann_date_posted ON announcements(date_posted);
`

// Setup initializes and verifies the database schema
func Setup(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	if _, err := db.Exec(schema); err != nil {
		db.Close()
		return nil, err
	}
	utils.Logger.Infof("✅ Database initialized and schema verified (%s)", path)
	return db, nil
}

// Connect connects to an existing SQLite database without altering schema
func Connect(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	utils.Logger.Infof("✅ Connected to database: %s", path)
	return db, nil
}

// SaveAnnouncement inserts or updates a full announcement
func SaveAnnouncement(db *sql.DB, a *models.Announcement) error {
	now := time.Now().UTC()
	_, err := db.Exec(`
	INSERT INTO announcements(
		ann_id, title, link, company_name, stock_name, date_posted, category, ref_number, content)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(ann_id)
	DO UPDATE SET
		title=excluded.title,
		link=excluded.link,
		company_name=excluded.company_name,
		stock_name=excluded.stock_name,
		category=excluded.category,
		ref_number=excluded.ref_number,
		content=excluded.content;`,
		a.AnnID, a.Title, a.Link, a.CompanyName, a.StockName, now, a.Category, a.RefNumber, a.Content)

	if err != nil {
		utils.Logger.Errorf("❌ DB insert error for ann_id %s: %v", a.AnnID, err)
	}
	return err
}

// UpdateAnnouncementInfo updates parsed fields after HTML parsing
func UpdateAnnouncement(db *sql.DB, a *models.Announcement) error {
	_, err := db.Exec(`
		UPDATE announcements
		SET company_name = ?, stock_name = ?, date_posted = ?, category = ?, ref_number = ?
		WHERE id = ?`,
		a.CompanyName, a.StockName, a.DatePosted, a.Category, a.RefNumber, a.ID)
	return err
}
