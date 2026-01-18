package models

import "time"

type Report struct {
	ID          string    `json:"id"`
	HackathonID string    `json:"hackathon_id"`
	ReporterID  string    `json:"reporter_id,omitempty"`
	Type        string    `json:"type"`
	Content     string    `json:"content"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}
