package models

import (
	"encoding/json"
	"time"
)

type Resource struct {
	ID          string          `json:"id"`
	HackathonID string          `json:"hackathon_id"`
	Type        string          `json:"type"`
	Title       string          `json:"title"`
	URL         string          `json:"url"`
	Metadata    json.RawMessage `json:"metadata,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
}
