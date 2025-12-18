package models

import (
	"time"
)

type BoardroomChange struct {
	ID                int        `json:"id,omitempty"`
	AnnID             int        `json:"ann_id"`
	CompanyName       string     `json:"company_name"`
	StockCode         string     `json:"stock_code"`
	PersonName        string     `json:"person_name"`
	PersonTitle       string     `json:"person_title"`
	PersonBirthYear   int        `json:"person_birth_year"`
	PersonGender      string     `json:"person_gender"`
	PersonNationality string     `json:"person_nationality"`
	DateAnnounced     *time.Time `json:"date_announced,omitempty"`
	DateOfChange      *time.Time `json:"date_of_change,omitempty"`
	Designation       string     `json:"designation,omitempty"`
	PreviousPosition  string     `json:"previous_position,omitempty"`
	Remarks           string     `json:"remarks,omitempty"`
	Directorate       string     `json:"directorate,omitempty"`
	TypeOfChange      string     `json:"type_of_change,omitempty"`
	Background        Background `json:"background,omitempty"`
	CreatedAt         time.Time  `json:"created_at,omitempty"`
}
