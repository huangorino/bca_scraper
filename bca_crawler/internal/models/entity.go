package models

import (
	"time"
)

type Entity struct {
	ID              int       `json:"id,omitempty" db:"id"`
	PrimaryPermID   *int      `json:"primary_perm_id,omitempty" db:"primary_perm_id"`
	SecondaryPermID int       `json:"secondary_perm_id,omitempty" db:"secondary_perm_id"`
	DisplayName     string    `json:"display_name,omitempty" db:"display_name"`
	Name            *string   `json:"name,omitempty" db:"name"`
	Salutation      *string   `json:"salutation,omitempty" db:"salutation"`
	StockCode       *string   `json:"stock_code,omitempty" db:"stock_code"`
	BirthYear       *int      `json:"birth_year,omitempty" db:"birth_year"`
	Gender          *string   `json:"gender,omitempty" db:"gender"`
	Nationality     *string   `json:"nationality,omitempty" db:"nationality"`
	CreatedAt       time.Time `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at,omitempty" db:"updated_at"`
}
