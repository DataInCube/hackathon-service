package models

import "time"

type Track struct {
	ID          string    `json:"id"`
	HackathonID string    `json:"hackathon_id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
