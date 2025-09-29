package models

import (
	"time"
)

type Registration struct {
	ID            uint      `json:"id"`
	ParticipantID uint      `json:"participant_id"`
	HackathonID   uint      `json:"hackathon_id"`
	TeamID        *uint     `json:"team_id"` // Peut être nul si pas encore assigné
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func (Registration) TableName() string {
	return "registrations"
}
