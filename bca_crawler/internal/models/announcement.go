package models

import (
	"database/sql"
	"time"
)

type Announcement struct {
	ID          int       `json:"id,omitempty"`
	AnnID       int       `json:"ann_id"`
	Title       string    `json:"title,omitempty"`
	Link        string    `json:"link"`
	CompanyName string    `json:"company_name,omitempty"`
	StockName   string    `json:"stock_name,omitempty"`
	DatePosted  time.Time `json:"date_posted,omitempty"`
	Category    string    `json:"category,omitempty"`
	RefNumber   string    `json:"ref_number,omitempty"`
	Content     string    `json:"content,omitempty"`
	Attachments []string  `json:"attachments,omitempty"`
}

type AnnouncementDB struct {
	ID          int            `db:"id"`
	AnnID       int            `db:"ann_id"`
	Title       sql.NullString `db:"title"`
	Link        sql.NullString `db:"link"`
	CompanyName sql.NullString `db:"company_name"`
	StockName   sql.NullString `db:"stock_name"`
	DatePosted  sql.NullTime   `db:"date_posted"`
	Category    sql.NullString `db:"category"`
	RefNumber   sql.NullString `db:"ref_number"`
	Content     sql.NullString `db:"content"`
	Attachments sql.NullString `db:"attachments"`
}
