package models

type Team struct {
    ID          uint   `gorm:"primaryKey"`
    Name        string `gorm:"not null"`
    HackathonID uint
    LeadID      uint // ID du participant leader
}

func (Team) TableName() string {
    return "teams"
}
