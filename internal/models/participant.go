package models

import (
    "time"
)

type Participant struct {
    ID    uint   `json:"id"`
    Name        string `json:"name"`
	Email       string `json:"email" gorm:"unique"`
	HackathonID uint   `json:"hackathon_id"`
	UserID    string    `json:"user_id"` // ID venant de Keycloak
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Participant) TableName() string {
    return "participants"
}