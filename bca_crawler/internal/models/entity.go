package models

import (
	"time"

	"github.com/google/uuid"
)

type EntityMaster struct {
	ID        int       `json:"id,omitempty"`
	ScID      uuid.UUID `json:"sc_id,omitempty"`
	Type      string    `json:"type"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

type Entity struct {
	ID          int       `json:"id,omitempty"`
	ScID        uuid.UUID `json:"sc_id,omitempty"`
	Prefix      string    `json:"prefix,omitempty"`
	Name        string    `json:"name"`
	Title       string    `json:"title,omitempty"`
	BirthYear   int       `json:"birth_year,omitempty"`
	Gender      string    `json:"gender,omitempty"`
	Nationality string    `json:"nationality,omitempty"`
	CreatedAt   time.Time `json:"created_at,omitempty"`
	UpdatedAt   time.Time `json:"updated_at,omitempty"`
}
