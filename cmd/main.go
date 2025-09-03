package main

import (
	"log"
	"github.com/DataInCube/hackathon-service/internal/models"
	"github.com/DataInCube/hackathon-service/internal/handlers"
	"github.com/DataInCube/hackathon-service/internal/routes"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dsn := "host=localhost user=postgres password=yourpassword dbname=hackathondb port=5432 sslmode=disable"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	// Auto-migrate models
	db.AutoMigrate(&models.Hackathon{}, &models.Participant{})

	handler := handlers.NewHandler(db)
	r := routes.SetupRouter(handler)

	log.Println("🚀 Hackathon Service running on http://localhost:8081")
	r.Run(":8081")
}