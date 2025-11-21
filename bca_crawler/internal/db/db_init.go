package db

import (
	"bca_crawler/internal/utils"
	"fmt"

	"github.com/jmoiron/sqlx"
)

const schema_lite = `
PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS announcements (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ann_id INTEGER UNIQUE,
    title TEXT,
    link TEXT UNIQUE,
    company_name TEXT,
    stock_name TEXT,
    date_posted TEXT,
    category TEXT,
    ref_number TEXT,
    content TEXT,
    attachments TEXT
);
CREATE INDEX IF NOT EXISTS idx_ann_date_posted ON announcements(date_posted);

CREATE TABLE IF NOT EXISTS entities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    type TEXT NOT NULL,
    name TEXT NOT NULL,
    stock_code TEXT,
    age INTEGER,
    gender TEXT,
    nationality TEXT,
    created_at TEXT DEFAULT (DATETIME('now')),
    updated_at TEXT DEFAULT (DATETIME('now'))
);
CREATE UNIQUE INDEX IF NOT EXISTS ux_entities_type_name_stock ON entities(type, name, IFNULL(stock_code, ''));

CREATE TABLE IF NOT EXISTS boardroom_changes (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    company_id INTEGER,
    person_id INTEGER,
    ann_id INTEGER UNIQUE,
    category TEXT,
    date_announced TEXT,
    date_of_change TEXT,
    designation TEXT,
    previous_position TEXT,
    remarks TEXT,
    directorate TEXT,
    type_of_change TEXT,
    created_at TEXT DEFAULT (DATETIME('now')),
    FOREIGN KEY (company_id) REFERENCES entities(id),
    FOREIGN KEY (person_id) REFERENCES entities(id)
);

CREATE TABLE IF NOT EXISTS backgrounds (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    entity_id INTEGER,
    qualification TEXT,
    working_experience TEXT,
    directorships TEXT,
    family_relationship TEXT,
    conflict_of_interest TEXT,
    interest_in_securities TEXT,
    FOREIGN KEY (entity_id) REFERENCES entities(id) ON DELETE CASCADE
);
CREATE UNIQUE INDEX IF NOT EXISTS idx_backgrounds_entity_id ON backgrounds(entity_id);
`

const schema_pg = `
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

CREATE TABLE IF NOT EXISTS stocks (
  id SERIAL PRIMARY KEY,
  stock_code VARCHAR(50) UNIQUE,
  stock_name VARCHAR(255),
  company_name VARCHAR(255),
  market VARCHAR(100),
  sector VARCHAR(100),
  website VARCHAR(255),
  status VARCHAR(10) DEFAULT 'ACTIVE'
);

CREATE TABLE IF NOT EXISTS daily_stock_prices (
  id SERIAL PRIMARY KEY,
  stock_code VARCHAR(50) NOT NULL,
  date DATE NOT NULL,
  close_price DECIMAL(10,4),
  change DECIMAL(10,4),
  change_perc DECIMAL(5,2),
  volume BIGINT,
  recommend varchar(30),
  high_price DECIMAL(10,4),
  low_price DECIMAL(10,4),
  buy_qty BIGINT,
  sell_qty BIGINT,
  mkt_cap DECIMAL(20,4),
  pe DECIMAL(10,4),
  eps DECIMAL(10,4),
  esg_rating INTEGER,
  UNIQUE(stock_code, date)
);
CREATE INDEX IF NOT EXISTS idx_daily_stock_prices_stock_code ON daily_stock_prices(stock_code);
CREATE INDEX IF NOT EXISTS idx_daily_stock_prices_date ON daily_stock_prices(date);

`

// DriverType represents supported database drivers
type DriverType string

const (
	DriverSQLite   DriverType = "sqlite"
	DriverPostgres DriverType = "postgres"
)

// Setup initializes and verifies the database schema
func Setup(connStr string, driver DriverType) (*sqlx.DB, error) {
	var db *sqlx.DB
	var err error

	switch driver {
	case DriverSQLite:
		db, err = sqlx.Open("sqlite", connStr)
		if err != nil {
			return nil, err
		}

		if _, err := db.Exec(schema_lite); err != nil {
			db.Close()
			return nil, err
		}
	case DriverPostgres:
		db, err = sqlx.Open("postgres", connStr+"?sslmode=disable")
		if err != nil {
			return nil, err
		}

		if _, err := db.Exec(schema_pg); err != nil {
			db.Close()
			return nil, err
		}
	}

	utils.Logger.Infof("Database initialized and schema verified (%s)", connStr)
	return db, nil
}

// Connect connects to an existing PostgreSQL database without altering schema
func Connect(connStr string, driver DriverType) (*sqlx.DB, error) {
	var db *sqlx.DB
	var err error

	switch driver {
	case DriverSQLite:
		db, err = sqlx.Open("sqlite", connStr)
	case DriverPostgres:
		db, err = sqlx.Open("postgres", connStr+"?sslmode=disable")
	}

	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	utils.Logger.Infof("Connected to database: %s", connStr)
	return db, nil
}
