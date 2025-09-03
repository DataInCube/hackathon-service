package models

type Registration struct {
    ID           uint `gorm:"primaryKey"`
    ParticipantID uint
    HackathonID  uint
    TeamID       *uint // Peut être nul si pas encore assigné
}

func (Registration) TableName() string {
    return "registrations"
}