package models

import (
	"encoding/json"
	"time"
)

type BursaStockResponse struct {
	RecordsTotal    string          `json:"recordsTotal"`
	RecordsFiltered string          `json:"recordsFiltered"`
	Data            [][]interface{} `json:"data"`
	Board           interface{}     `json:"board"`
}

type Investor struct {
	Name          string `json:"Name"`
	ChangePercent string `json:"ChangePercent"`
	SharesMillion string `json:"SharesMillion"`
}

type Management struct {
	Name        string `json:"Name"`
	Role        string `json:"Role"`
	Since       string `json:"Since"`
	Designation string `json:"Designation"`
}

type Stock struct {
	ID           int              `json:"id,omitempty" db:"id"`
	StockCode    string           `json:"stock_code,omitempty" db:"stock_code"`
	BursaID      *int32           `json:"bursa_id,omitempty" db:"bursa_id"`
	BursaCode    *string          `json:"bursa_code,omitempty" db:"bursa_code"`
	Name         *string          `json:"name,omitempty" db:"name"`
	Sector       *string          `json:"sector,omitempty" db:"sector"`
	SubSector    *string          `json:"sub_sector,omitempty" db:"sub_sector"`
	Type         *string          `json:"type,omitempty" db:"type"`
	Shariah      *bool            `json:"shariah,omitempty" db:"shariah"`
	CreatedAt    *time.Time       `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt    *time.Time       `json:"updated_at,omitempty" db:"updated_at"`
	StockName    *string          `json:"stock_name,omitempty" db:"stock_name"`
	Board        *string          `json:"board,omitempty" db:"board"`
	About        *string          `json:"about,omitempty" db:"about"`
	Website      *string          `json:"website,omitempty" db:"website"`
	Address      *string          `json:"address,omitempty" db:"address"`
	Phone        *string          `json:"phone,omitempty" db:"phone"`
	Fax          *string          `json:"fax,omitempty" db:"fax"`
	Management   *json.RawMessage `json:"management,omitempty" db:"management"`
	Ownership    *json.RawMessage `json:"ownership,omitempty" db:"ownership"`
	TopInvestors *json.RawMessage `json:"top_investors,omitempty" db:"top_investors"`
	Insiders     *json.RawMessage `json:"insiders,omitempty" db:"insiders"`
}

type StockRow struct {
	Index         int
	Name          string
	Code          string
	Market        string
	LastPrice     string
	ChangePrice   string
	ChangeValue   string
	ChangePercent string
	Volume        string
	Value         string
	Bid           string
	Ask           string
	BidVolume     string
	High          string
	Low           string
	Misc          string
}
