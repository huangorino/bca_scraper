package db

import (
	"bca_crawler/internal/utils"
	"fmt"

	"github.com/jmoiron/sqlx"
)

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


CREATE TABLE IF NOT EXISTS entities (
    id SERIAL PRIMARY KEY,
    type TEXT NOT NULL,
    name TEXT NOT NULL,
    title TEXT,
    stock_code TEXT,
    birth_year INTEGER,
    gender TEXT,
    nationality TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

DROP INDEX IF EXISTS uq_entities_company;
DROP INDEX IF EXISTS uq_entities_person;
CREATE UNIQUE INDEX uq_entities_company ON entities (type, name, stock_code) WHERE type = 'COMPANY';
CREATE UNIQUE INDEX uq_entities_person ON entities (type, name, birth_year) WHERE type = 'PERSON';


CREATE TABLE IF NOT EXISTS entities_level1 (
    id SERIAL PRIMARY KEY,
    type TEXT NOT NULL,
    name TEXT NOT NULL,
    title TEXT,
    company TEXT,
    birth_year INTEGER,
    gender TEXT,
    nationality TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

DROP INDEX IF EXISTS uq_entities_level1_person;
CREATE UNIQUE INDEX uq_entities_level1_person ON entities_level1 (name, company, birth_year);


CREATE TABLE IF NOT EXISTS boardroom_changes (
    id SERIAL PRIMARY KEY,
    company_id INTEGER REFERENCES entities(id),
    person_id INTEGER REFERENCES entities(id),
    ann_id INTEGER UNIQUE,
    category TEXT,
    date_announced DATE,
    date_of_change DATE,
    designation TEXT,
    previous_position TEXT,
    remarks TEXT,
    directorate TEXT,
    type_of_change TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS backgrounds (
    id SERIAL PRIMARY KEY,
    entity_id INTEGER NOT NULL REFERENCES entities(id) ON DELETE CASCADE,
    qualification TEXT,
    working_experience TEXT,
    directorships TEXT,
    family_relationship TEXT,
    conflict_of_interest TEXT,
    interest_in_securities TEXT
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_backgrounds_entity_id
ON backgrounds(entity_id);
`

// DriverType represents supported database drivers
type DriverType string

const (
	DriverPostgres DriverType = "postgres"
)

// Setup initializes and verifies the database schema
func Setup(connStr string, driver DriverType) (*sqlx.DB, error) {
	var db *sqlx.DB
	var err error

	switch driver {
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
