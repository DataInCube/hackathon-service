package main

import (
	"os"
	"time"

	"github.com/DataInCube/hackathon-service/api/middlewares"
	"github.com/DataInCube/hackathon-service/api/routes"
	"github.com/DataInCube/hackathon-service/pkg/utils"

	_ "github.com/DataInCube/hackathon-service/docs"
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
	echoSwagger "github.com/swaggo/echo-swagger"
)

// @title Hackathon API
// @version 1.0
// @description Hackathon microservice API documentation with Swagger
// @host localhost:8080
// @BasePath /api/v1

func main() {
	// Load .env file if present
	_ = godotenv.Load(".env")

	// Logger setup
	logger := logrus.New()
	logger.SetFormatter(&logrus.TextFormatter{
		FullTimestamp: true,
	})

	// Load DB config
	driver := getEnv("DB_DRIVER", "postgres")
	dsn := getEnv("DB_DSN", "host=localhost user=charme password= dbname=hackathondb port=5432 sslmode=disable")

	// Init DB
	db, err := utils.NewDB(driver, dsn, 10, 5, time.Minute*30, logger)
	if err != nil {
		logger.Fatal("Failed to connect to DB: ", err)
	}

	// Echo instance
	e := echo.New()

	// Middlewares
	e.Use(middlewares.MetricsMiddleware()) // Prometheus
	// e.Use(middlewares.LoggerMiddleware()) // Optionnel
	// e.Use(middlewares.RecoverMiddleware(logger)) // Recover from panics

	// Register API routes
	routes.RegisterRoutes(e, db, logger)

	// Swagger documentation
	e.GET("/swagger/*", echoSwagger.WrapHandler)
	// e.GET("/api/v1/swagger/*", echoSwagger.WrapHandler(echoSwagger.Name("Hackathon API")))

	// Start server
	port := getEnv("PORT", "8081")
	logger.Infof("Hackathon Service running on port %s", port)
	e.Logger.Fatal(e.Start(":" + port))

}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
