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
