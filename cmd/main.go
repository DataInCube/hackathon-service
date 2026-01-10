package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/DataInCube/hackathon-service/api/middlewares"
	"github.com/DataInCube/hackathon-service/api/routes"
	"github.com/DataInCube/hackathon-service/pkg/events"
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

	serviceName := getEnv("SERVICE_NAME", "hackathon-service")
	serviceVersion := getEnv("SERVICE_VERSION", "dev")

	foundationMode := isTruthy(getEnv("FOUNDATION_MODE", "false"))
	if foundationMode {
		e := echo.New()

		e.GET("/health", func(c echo.Context) error {
			return c.JSON(200, map[string]string{"status": "ok"})
		})
		e.GET("/version", func(c echo.Context) error {
			return c.JSON(200, map[string]string{"service": serviceName, "version": serviceVersion})
		})

		port := getEnv("PORT", "8081")
		logger.Infof("Hackathon Service (foundation mode) running on port %s", port)
		e.Logger.Fatal(e.Start(":" + port))
		return
	}

	// Load DB config
	driver := getEnv("DB_DRIVER", "postgres")
	dsn := getEnv("DB_DSN", "")
	if dsn == "" {
		host := getEnv("DB_HOST", "localhost")
		port := getEnv("DB_PORT", "5432")
		user := getEnv("DB_USER", "postgres")
		password := getEnv("DB_PASSWORD", "postgres")
		name := getEnv("DB_NAME", "hackathondb")
		sslMode := getEnv("DB_SSLMODE", "disable")
		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, name, sslMode)
	}

	// Init DB
	maxOpenConns := getEnvInt("DB_MAX_OPEN_CONNS", 10)
	maxIdleConns := getEnvInt("DB_MAX_IDLE_CONNS", 5)
	connMaxLifetime := time.Duration(getEnvInt("DB_CONN_MAX_LIFETIME_MINUTES", 30)) * time.Minute
	db, err := utils.NewDB(driver, dsn, maxOpenConns, maxIdleConns, connMaxLifetime, logger)
	if err != nil {
		logger.Fatal("Failed to connect to DB: ", err)
	}

	// Echo instance
	e := echo.New()

	// Middlewares
	e.Use(middlewares.MetricsMiddleware()) // Prometheus
	// e.Use(middlewares.LoggerMiddleware()) // Optionnel
	// e.Use(middlewares.RecoverMiddleware(logger)) // Recover from panics

	authCfg := middlewares.JWTConfig{
		JWKSURL:  getEnv("AUTH_JWKS_URL", ""),
		Issuer:   getEnv("AUTH_ISSUER", ""),
		Audience: getEnv("AUTH_AUDIENCE", ""),
		ClientID: getEnv("AUTH_CLIENT_ID", ""),
		Required: isTruthy(getEnv("AUTH_REQUIRED", "true")),
	}

	authMiddleware, err := middlewares.AuthMiddleware(authCfg, logger)
	if err != nil {
		logger.Fatal("Failed to initialize auth middleware: ", err)
	}

	var publisher events.Publisher
	if isTruthy(getEnv("EVENTS_ENABLED", "true")) {
		natsURL := getEnv("NATS_URL", "nats://nats:4222")
		stream := getEnv("NATS_STREAM", "SENTIO_EVENTS")
		subjects := []string{
			getEnv("NATS_SUBJECT_HACKATHON_CREATED", "hackathon.created"),
			getEnv("NATS_SUBJECT_HACKATHON_PUBLISHED", "hackathon.published"),
			getEnv("NATS_SUBJECT_HACKATHON_PHASE_CHANGED", "hackathon.phase.changed"),
			getEnv("NATS_SUBJECT_HACKATHON_COMPLETED", "hackathon.completed"),
			getEnv("NATS_SUBJECT_SUBMISSION_CREATED", "submission.created"),
			getEnv("NATS_SUBJECT_SUBMISSION_LOCKED", "submission.locked"),
			getEnv("NATS_SUBJECT_SUBMISSION_INVALIDATED", "submission.invalidated"),
			getEnv("NATS_SUBJECT_LEADERBOARD_FREEZE", "leaderboard.freeze.requested"),
			getEnv("NATS_SUBJECT_LEADERBOARD_UNFREEZE", "leaderboard.unfreeze.requested"),
			getEnv("NATS_SUBJECT_LEADERBOARD_PUBLISH", "leaderboard.publish.requested"),
			getEnv("NATS_SUBJECT_TEAM_REQUIRED", "hackathon.team.required"),
			getEnv("NATS_SUBJECT_TEAM_LOCKED", "hackathon.team.locked"),
			getEnv("NATS_SUBJECT_RULE_CREATED", "hackathon.rule.created"),
			getEnv("NATS_SUBJECT_RULE_VERSION_LOCKED", "hackathon.rule.version.locked"),
			getEnv("NATS_SUBJECT_RULE_ACTIVATED", "hackathon.rule.activated"),
		}
		natsPublisher, err := events.NewNatsPublisher(natsURL, stream, uniqueStrings(subjects))
		if err != nil {
			logger.Fatal("Failed to connect to NATS: ", err)
		}
		defer natsPublisher.Close()
		publisher = natsPublisher
	}

	// Register API routes
	routes.RegisterRoutes(e, db, logger, authMiddleware, serviceName, serviceVersion, publisher)

	// Swagger documentation
	e.GET("/swagger/*", echoSwagger.WrapHandler)
	// e.GET("/api/v1/swagger/*", echoSwagger.WrapHandler(echoSwagger.Name("Hackathon API")))

	// Start server
	port := getEnv("PORT", "8081")
	logger.Infof("Hackathon Service running on port %s", port)
	server := &http.Server{
		Addr:         ":" + port,
		Handler:      e,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  60 * time.Second,
	}
	e.Logger.Fatal(e.StartServer(server))

}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			return parsed
		}
	}
	return fallback
}

func isTruthy(val string) bool {
	switch strings.ToLower(strings.TrimSpace(val)) {
	case "1", "true", "yes", "y", "on":
		return true
	default:
		return false
	}
}

func uniqueStrings(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, v := range in {
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}
