package models

import "time"

type Message struct {
	ID          uint      `json:"id"`
	SenderID    uint      `json:"sender_id"`
	TeamID      *uint     `json:"team_id,omitempty"`      // Null si message global hackathon
	HackathonID *uint     `json:"hackathon_id,omitempty"` // Null si message d’équipe
	Content     string    `json:"content"`
	CreatedAt   time.Time `json:"created_at"`
}

func (Message) TableName() string {
	return "messages"
}