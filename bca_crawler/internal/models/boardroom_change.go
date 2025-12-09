package models

import "time"

type BoardroomChange struct {
	ID               int       `json:"id,omitempty"`
	CompanyID        int       `json:"company_id"`
	PersonID         int       `json:"person_id"`
	AnnID            int       `json:"ann_id"`
	Category         string    `json:"category,omitempty"`
	DateAnnounced    time.Time `json:"date_announced,omitempty"`
	DateOfChange     time.Time `json:"date_of_change,omitempty"`
	Designation      string    `json:"designation,omitempty"`
	PreviousPosition string    `json:"previous_position,omitempty"`
	Remarks          string    `json:"remarks,omitempty"`
	Directorate      string    `json:"directorate,omitempty"`
	TypeOfChange     string    `json:"type_of_change,omitempty"`
	CreatedAt        string    `json:"created_at,omitempty"`
}
