package models

import (
	"time"
)

type Team struct {
    ID          uint      `json:"id"`
	Name        string    `json:"name"`
	HackathonID uint      `json:"hackathon_id"`
	LeadID      uint      `json:"lead_id"` // FK vers Participant
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Team) TableName() string {
    return "teams"
}
