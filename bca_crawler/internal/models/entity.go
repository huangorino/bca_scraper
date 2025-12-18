package models

import (
	"time"
)

type EntityMaster struct {
	ScID      int       `json:"sc_id,omitempty"`
	Type      string    `json:"type"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}

type Entity struct {
	ScID      int       `json:"sc_id,omitempty"`
	Prefix    string    `json:"prefix,omitempty"`
	Value     string    `json:"value"`
	CreatedAt time.Time `json:"created_at,omitempty"`
	UpdatedAt time.Time `json:"updated_at,omitempty"`
}
