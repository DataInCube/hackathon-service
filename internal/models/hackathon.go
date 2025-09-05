package models

import (
	"time"
)

type Hackathon struct {
	ID          uint      `json:"id"`
	Title      string    `json:"title"`
	Description string    `json:"description"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	MaxTeams    int       `json:"max_teams"`
	MaxMembers  int       `json:"max_members"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func (Hackathon) TableName() string {
	return "hackathons"
}
