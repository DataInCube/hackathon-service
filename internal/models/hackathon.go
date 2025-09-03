package models

import (
	"time"

	"gorm.io/gorm"
)

type Hackathon struct {
	gorm.Model
	ID          uint      `gorm:"primaryKey"`
	TTitle      string    `json:"title"`
	Description string    `json:"description"`
	StartDate   time.Time `json:"start_date"`
	EndDate     time.Time `json:"end_date"`
	MaxTeams    int
	MaxMembers  int
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (Hackathon) TableName() string {
	return "hackathons"
}
