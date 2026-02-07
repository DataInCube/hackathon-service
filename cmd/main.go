package main

import (
	"fmt"
	"net/http"
	"time"

	dbutils "github.com/DataInCube/go-utils/db"
	"github.com/DataInCube/go-utils/env"
	"github.com/DataInCube/go-utils/stringsx"
	"github.com/DataInCube/hackathon-service/api/middlewares"
	"github.com/DataInCube/hackathon-service/api/routes"
	"github.com/DataInCube/hackathon-service/pkg/events"

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

	serviceName := env.GetString("SERVICE_NAME", "hackathon-service")
	serviceVersion := env.GetString("SERVICE_VERSION", "dev")

	foundationMode := env.GetBool("FOUNDATION_MODE", false)
	if foundationMode {
		e := echo.New()

		e.GET("/health", func(c echo.Context) error {
			return c.JSON(200, map[string]string{"status": "ok"})
		})
		e.GET("/version", func(c echo.Context) error {
			return c.JSON(200, map[string]string{"service": serviceName, "version": serviceVersion})
		})

		port := env.GetString("PORT", "8081")
		logger.Infof("Hackathon Service (foundation mode) running on port %s", port)
		e.Logger.Fatal(e.Start(":" + port))
		return
	}

	// Load DB config
	driver := env.GetString("DB_DRIVER", "postgres")
	dsn := env.GetString("DB_DSN", "")
	if dsn == "" {
		host := env.GetString("DB_HOST", "localhost")
		port := env.GetString("DB_PORT", "5432")
		user := env.GetString("DB_USER", "postgres")
		password := env.GetString("DB_PASSWORD", "postgres")
		name := env.GetString("DB_NAME", "hackathondb")
		sslMode := env.GetString("DB_SSLMODE", "disable")
		dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s", host, port, user, password, name, sslMode)
	}

	// Init DB
	maxOpenConns := env.GetInt("DB_MAX_OPEN_CONNS", 10)
	maxIdleConns := env.GetInt("DB_MAX_IDLE_CONNS", 5)
	connMaxLifetime := time.Duration(env.GetInt("DB_CONN_MAX_LIFETIME_MINUTES", 30)) * time.Minute
	db, err := dbutils.Connect(driver, dsn, dbutils.PoolConfig{
		MaxOpenConns:    maxOpenConns,
		MaxIdleConns:    maxIdleConns,
		ConnMaxLifetime: connMaxLifetime,
	})
	if err != nil {
		logger.Fatal("Failed to connect to DB: ", err)
	}
	logger.Println("✅ Connected to PostgreSQL database")

	// Echo instance
	e := echo.New()

	// Middlewares
	e.Use(middlewares.MetricsMiddleware()) // Prometheus
	// e.Use(middlewares.LoggerMiddleware()) // Optionnel
	// e.Use(middlewares.RecoverMiddleware(logger)) // Recover from panics

	authCfg := middlewares.JWTConfig{
		JWKSURL:  env.GetString("AUTH_JWKS_URL", ""),
		Issuer:   env.GetString("AUTH_ISSUER", ""),
		Audience: env.GetString("AUTH_AUDIENCE", ""),
		ClientID: env.GetString("AUTH_CLIENT_ID", ""),
		Required: env.GetBool("AUTH_REQUIRED", true),
	}

	authMiddleware, err := middlewares.AuthMiddleware(authCfg, logger)
	if err != nil {
		logger.Fatal("Failed to initialize auth middleware: ", err)
	}

	var publisher events.Publisher
	if env.GetBool("EVENTS_ENABLED", true) {
		natsURL := env.GetString("NATS_URL", "nats://nats:4222")
		stream := env.GetString("NATS_STREAM", "SENTIO_EVENTS")
		subjects := []string{
			env.GetString("NATS_SUBJECT_HACKATHON_CREATED", "hackathon.created"),
			env.GetString("NATS_SUBJECT_HACKATHON_PUBLISHED", "hackathon.published"),
			env.GetString("NATS_SUBJECT_HACKATHON_PHASE_CHANGED", "hackathon.phase.changed"),
			env.GetString("NATS_SUBJECT_HACKATHON_COMPLETED", "hackathon.completed"),
			env.GetString("NATS_SUBJECT_HACKATHON_DATA_CREATED", "hackathon.data.created"),
			env.GetString("NATS_SUBJECT_HACKATHON_DATA_UPDATED", "hackathon.data.updated"),
			env.GetString("NATS_SUBJECT_HACKATHON_DATA_DELETED", "hackathon.data.deleted"),
			env.GetString("NATS_SUBJECT_HACKATHON_DATA_FILE_CREATED", "hackathon.data.file.created"),
			env.GetString("NATS_SUBJECT_HACKATHON_DATA_FILE_UPDATED", "hackathon.data.file.updated"),
			env.GetString("NATS_SUBJECT_HACKATHON_DATA_FILE_DELETED", "hackathon.data.file.deleted"),
			env.GetString("NATS_SUBJECT_HACKATHON_DATA_VARIABLE_CREATED", "hackathon.data.variable.created"),
			env.GetString("NATS_SUBJECT_HACKATHON_DATA_VARIABLE_UPDATED", "hackathon.data.variable.updated"),
			env.GetString("NATS_SUBJECT_HACKATHON_DATA_VARIABLE_DELETED", "hackathon.data.variable.deleted"),
			env.GetString("NATS_SUBJECT_HACKATHON_METRIC_CREATED", "hackathon.metric.created"),
			env.GetString("NATS_SUBJECT_HACKATHON_METRIC_UPDATED", "hackathon.metric.updated"),
			env.GetString("NATS_SUBJECT_HACKATHON_METRIC_DELETED", "hackathon.metric.deleted"),
			env.GetString("NATS_SUBJECT_SUBMISSION_LIMITS_CREATED", "hackathon.submission_limits.created"),
			env.GetString("NATS_SUBJECT_SUBMISSION_LIMITS_UPDATED", "hackathon.submission_limits.updated"),
			env.GetString("NATS_SUBJECT_SUBMISSION_LIMITS_DELETED", "hackathon.submission_limits.deleted"),
			env.GetString("NATS_SUBJECT_SUBMISSION_CREATED", "submission.created"),
			env.GetString("NATS_SUBJECT_SUBMISSION_LOCKED", "submission.locked"),
			env.GetString("NATS_SUBJECT_SUBMISSION_INVALIDATED", "submission.invalidated"),
			env.GetString("NATS_SUBJECT_EVALUATION_COMPLETED", "evaluation.completed"),
			env.GetString("NATS_SUBJECT_LEADERBOARD_FREEZE", "leaderboard.freeze.requested"),
			env.GetString("NATS_SUBJECT_LEADERBOARD_UNFREEZE", "leaderboard.unfreeze.requested"),
			env.GetString("NATS_SUBJECT_LEADERBOARD_PUBLISH", "leaderboard.publish.requested"),
			env.GetString("NATS_SUBJECT_TEAM_REQUIRED", "hackathon.team.required"),
			env.GetString("NATS_SUBJECT_TEAM_LOCKED", "hackathon.team.locked"),
			env.GetString("NATS_SUBJECT_RULE_CREATED", "hackathon.rule.created"),
			env.GetString("NATS_SUBJECT_RULE_VERSION_LOCKED", "hackathon.rule.version.locked"),
			env.GetString("NATS_SUBJECT_RULE_ACTIVATED", "hackathon.rule.activated"),
		}
		natsPublisher, err := events.NewNatsPublisher(natsURL, stream, stringsx.UniqueStrings(subjects))
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
	port := env.GetString("PORT", "8081")
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
