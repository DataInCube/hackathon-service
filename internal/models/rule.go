package models

import "time"

type Rule struct {
	ID          string    `json:"id"`
	HackathonID string    `json:"hackathon_id"`
	TrackID     *string   `json:"track_id,omitempty"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
