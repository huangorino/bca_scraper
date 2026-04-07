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

type Company struct {
	ID                int        `json:"id,omitempty" db:"id"`
	PrimaryPermID     int32      `json:"primary_perm_id,omitempty" db:"primary_perm_id"`
	SecondaryPermID   int32      `json:"secondary_perm_id,omitempty" db:"secondary_perm_id"`
	CompanyName       string     `json:"company_name,omitempty" db:"company_name"`
	RegNo             string     `json:"reg_no,omitempty" db:"reg_no"`
	RegNoOld          string     `json:"reg_no_old,omitempty" db:"reg_no_old"`
	DateIncorporation string     `json:"date_incorporation,omitempty" db:"date_incorporation"`
	IncorporateUnder  string     `json:"incorporate_under,omitempty" db:"incorporate_under"`
	StockCode         string     `json:"stock_code,omitempty" db:"stock_code"`
	About             string     `json:"about,omitempty" db:"about"`
	Industry          string     `json:"industry,omitempty" db:"industry"`
	Website           string     `json:"website,omitempty" db:"website"`
	Address           string     `json:"address,omitempty" db:"address"`
	Phone             string     `json:"phone,omitempty" db:"phone"`
	Fax               string     `json:"fax,omitempty" db:"fax"`
	IrContact         string     `json:"ir_contact,omitempty" db:"ir_contact"`
	GroupStructure    string     `json:"group_structure,omitempty" db:"group_structure"`
	IndustryOverview  string     `json:"industry_overview,omitempty" db:"industry_overview"`
	CreatedAt         *time.Time `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt         *time.Time `json:"updated_at,omitempty" db:"updated_at"`
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
	NewName      *string          `json:"new_name,omitempty" db:"new_name"`
	DateNewName  *time.Time       `json:"date_new_name,omitempty" db:"date_new_name"`
	Active       *bool            `json:"active,omitempty" db:"active"`
	RegNo        *string          `json:"reg_no,omitempty" db:"reg_no"`
	RegNoOld     *string          `json:"reg_no_old,omitempty" db:"reg_no_old"`
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
