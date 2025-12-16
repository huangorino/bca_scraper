package models

import "time"

type Entity struct {
	ID          int       `json:"id,omitempty"`
	Type        string    `json:"type"`
	Name        string    `json:"name"`
	Title       string    `json:"title,omitempty"`
	StockCode   string    `json:"stock_code,omitempty"`
	ICNumber    string    `json:"ic_number,omitempty"`
	BirthYear   int       `json:"birth_year,omitempty"`
	Gender      string    `json:"gender,omitempty"`
	Nationality string    `json:"nationality,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}
