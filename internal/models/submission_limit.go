package models

import "time"

type SubmissionLimit struct {
	ID          string    `json:"id"`
	HackathonID string    `json:"hackathon_id"`
	PerDay      int       `json:"per_day"`
	Total       int       `json:"total"`
	PerTeam     int       `json:"per_team"`
	Notes       string    `json:"notes,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
