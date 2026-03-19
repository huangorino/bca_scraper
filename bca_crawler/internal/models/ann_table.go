package models

import (
	"time"
)

type BoardroomChange struct {
	ID                int        `json:"id,omitempty" db:"id"`
	AnnID             *int       `json:"ann_id,omitempty" db:"ann_id"`
	CompanyName       *string    `json:"company_name,omitempty" db:"company_name"`
	StockCode         *string    `json:"stock_code,omitempty" db:"stock_code"`
	PersonName        *string    `json:"person_name,omitempty" db:"person_name"`
	PersonTitle       *string    `json:"person_title,omitempty" db:"person_title"`
	PersonBirthYear   *int       `json:"person_birth_year,omitempty" db:"person_birth_year"`
	PersonGender      *string    `json:"person_gender,omitempty" db:"person_gender"`
	PersonNationality *string    `json:"person_nationality,omitempty" db:"person_nationality"`
	DateAnnounced     *time.Time `json:"date_announced,omitempty" db:"date_announced"`
	DateOfChange      *time.Time `json:"date_of_change,omitempty" db:"date_of_change"`
	Designation       *string    `json:"designation,omitempty" db:"designation"`
	PreviousPosition  *string    `json:"previous_position,omitempty" db:"previous_position"`
	Remarks           *string    `json:"remarks,omitempty" db:"remarks"`
	Directorate       *string    `json:"directorate,omitempty" db:"directorate"`
	TypeOfChange      *string    `json:"type_of_change,omitempty" db:"type_of_change"`
	Background        Background `json:"background,omitempty" db:"background"`
	RelatedPerm       *int       `json:"related_perm,omitempty" db:"related_perm"`
	CreatedAt         time.Time  `json:"created_at,omitempty" db:"created_at"`
}

type ShareholdingChange struct {
	ID                      int        `json:"id,omitempty" db:"id"`
	AnnID                   int        `json:"ann_id" db:"ann_id"`
	StockCode               string     `json:"stock_code" db:"stock_code"`
	CompanyName             *string    `json:"company_name,omitempty" db:"company_name"`
	ChangeType              *string    `json:"change_type,omitempty" db:"change_type"`
	PersonName              *string    `json:"person_name,omitempty" db:"person_name"`
	PersonAddress           *string    `json:"person_address,omitempty" db:"person_address"`
	PersonNationality       *string    `json:"person_nationality,omitempty" db:"person_nationality"`
	CompanyNo               *string    `json:"company_no,omitempty" db:"company_no"`
	SecurityDescription     *string    `json:"security_description,omitempty" db:"security_description"`
	RegisteredHolder        *string    `json:"registered_holder,omitempty" db:"registered_holder"`
	RegisteredHolderAddress *string    `json:"registered_holder_address,omitempty" db:"registered_holder_address"`
	TransactionType         *string    `json:"transaction_type,omitempty" db:"transaction_type"`
	TransactionDesc         *string    `json:"transaction_desc,omitempty" db:"transaction_desc"`
	Currency                *string    `json:"currency,omitempty" db:"currency"`
	DateOfChange            *time.Time `json:"date_of_change,omitempty" db:"date_of_change"`
	DateInterestAcquired    *time.Time `json:"date_interest_acquired,omitempty" db:"date_interest_acquired"`
	DateOfCessation         *time.Time `json:"date_of_cessation,omitempty" db:"date_of_cessation"`
	SecuritiesChanged       *int64     `json:"securities_changed,omitempty" db:"securities_changed"`
	PriceTransacted         *float64   `json:"price_transacted,omitempty" db:"price_transacted"`
	NatureOfInterest        *string    `json:"nature_of_interest,omitempty" db:"nature_of_interest"`
	Circumstances           *string    `json:"circumstances,omitempty" db:"circumstances"`
	Consideration           *string    `json:"consideration,omitempty" db:"consideration"`
	DirectUnits             *int64     `json:"direct_units,omitempty" db:"direct_units"`
	DirectPercent           *float64   `json:"direct_percent,omitempty" db:"direct_percent"`
	IndirectUnits           *int64     `json:"indirect_units,omitempty" db:"indirect_units"`
	IndirectPercent         *float64   `json:"indirect_percent,omitempty" db:"indirect_percent"`
	TotalSecurities         *int64     `json:"total_securities,omitempty" db:"total_securities"`
	DateOfNotice            *time.Time `json:"date_of_notice,omitempty" db:"date_of_notice"`
	DateNoticeReceived      *time.Time `json:"date_notice_received,omitempty" db:"date_notice_received"`
	Remarks                 *string    `json:"remarks,omitempty" db:"remarks"`
	CreatedAt               time.Time  `json:"created_at,omitempty" db:"created_at"`
	RelatedPerm             *int       `json:"related_perm,omitempty" db:"related_perm"`
}
