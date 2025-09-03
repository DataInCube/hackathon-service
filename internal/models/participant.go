package models

import (
    "gorm.io/gorm"
)

type Participant struct {
    gorm.Model
    ID    uint   `gorm:"primaryKey"`
    Name        string `json:"name"`
	Email       string `json:"email" gorm:"unique"`
	HackathonID uint   `json:"hackathon_id"`
}

func (Participant) TableName() string {
    return "participants"
}