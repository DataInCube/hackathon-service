package models

import (
	"encoding/json"
	"time"
)

type Submission struct {
	ID            string          `json:"id"`
	HackathonID   string          `json:"hackathon_id"`
	TrackID       *string         `json:"track_id,omitempty"`
	RuleVersionID string          `json:"rule_version_id"`
	SubmittedBy   string          `json:"submitted_by"`
	TeamID        *string         `json:"team_id,omitempty"`
	Status        string          `json:"status"`
	Phase         string          `json:"phase"`
	Metadata      json.RawMessage `json:"metadata,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`
	LockedAt      *time.Time      `json:"locked_at,omitempty"`
	InvalidatedAt *time.Time      `json:"invalidated_at,omitempty"`
}
