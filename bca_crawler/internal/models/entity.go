package models

import (
	"time"
)

type Entity struct {
	ID              int       `json:"id,omitempty" db:"id"`
	PrimaryPermID   *int      `json:"primary_perm_id,omitempty" db:"primary_perm_id"`
	SecondaryPermID int       `json:"secondary_perm_id,omitempty" db:"secondary_perm_id"`
	DisplayName     *string   `json:"display_name,omitempty" db:"display_name"`
	Name            *string   `json:"name,omitempty" db:"name"`
	AliasName       *string   `json:"alias_name,omitempty" db:"alias_name"`
	OriName         *string   `json:"ori_name,omitempty" db:"ori_name"`
	Salutation      *string   `json:"salutation,omitempty" db:"salutation"`
	StockCode       *string   `json:"stock_code,omitempty" db:"stock_code"`
	BirthYear       *int      `json:"birth_year,omitempty" db:"birth_year"`
	Gender          *string   `json:"gender,omitempty" db:"gender"`
	Nationality     *string   `json:"nationality,omitempty" db:"nationality"`
	NewIC           *string   `json:"new_ic,omitempty" db:"new_ic"`
	ImgPath         *string   `json:"img_path,omitempty" db:"img_path"`
	CreatedAt       time.Time `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at,omitempty" db:"updated_at"`
}

type EntityRole struct {
	ID            int        `json:"id,omitempty" db:"id"`
	PermID        int        `json:"perm_id,omitempty" db:"perm_id"`
	CompanyName   string     `json:"company_name,omitempty" db:"company_name"`
	StockName     string     `json:"stock_name,omitempty" db:"stock_name"`
	DateAppointed *time.Time `json:"date_appointed,omitempty" db:"date_appointed"`
	DateResigned  *time.Time `json:"date_resigned,omitempty" db:"date_resigned"`
	Category      string     `json:"category,omitempty" db:"category"`
	RoleName      string     `json:"role_name,omitempty" db:"role_name"`
	Promoted      bool       `json:"promoted,omitempty" db:"promoted"`
	Alternative   bool       `json:"alternative,omitempty" db:"alternative"`
	Independentcy bool       `json:"independentcy,omitempty" db:"independentcy"`
	Executive     bool       `json:"executive,omitempty" db:"executive"`
	Chairmanship  bool       `json:"chairmanship,omitempty" db:"chairmanship"`
	DateCreated   time.Time  `json:"date_created,omitempty" db:"date_created"`
	DateUpdated   time.Time  `json:"date_updated,omitempty" db:"date_updated"`
}

type RoleChangeInput struct {
	PermID           int
	StockCode        string
	CompanyName      string
	Designation      string
	PreviousPosition string
	TypeOfChange     string
	DateOfChange     *time.Time
	Category         string
}

type ExHistEntity struct {
	StockCode    *string    `json:"stock_code" db:"stock_code"`
	CompanyName  *string    `json:"company_name" db:"company_name"`
	Title        *string    `json:"title" db:"title"`
	DirectorName *string    `json:"director_name" db:"director_name"`
	AppDate      *time.Time `json:"pst_app_date" db:"pst_app_date"`
	ResDate      *time.Time `json:"pst_res_date" db:"pst_res_date"`
	Designation  *string    `json:"designation" db:"designation"`
	BirthDate    *time.Time `json:"birth_date" db:"birth_date"`
	Gender       *string    `json:"gender" db:"gender"`
	Nationality  *string    `json:"nationality" db:"nationality"`
	NewIC        *string    `json:"new_ic" db:"new_ic"`
}
