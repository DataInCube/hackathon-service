package routes

import (
	"database/sql"

	"github.com/DataInCube/hackathon-service/api/handlers"
	"github.com/DataInCube/hackathon-service/api/middlewares"
	"github.com/DataInCube/hackathon-service/api/services"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func RegisterRoutes(e *echo.Echo, db *sql.DB, logger *logrus.Logger) {
	// Middleware global (logger, recover, CORS, etc. à ajouter ici si besoin)

	// Auth middleware (appelle keycloak-service)
	authMiddleware := middlewares.AuthMiddleware()

	// Injecter les services
	hackathonService := services.NewHackathonService(db)
	participantService := services.NewParticipantService(db)
	teamService := services.NewTeamService(db)
	registrationService := services.NewRegistrationService(db)

	// Injecter les handlers
	hackathonHandler := handlers.NewHackathonHandler(hackathonService)
	participantHandler := handlers.NewParticipantHandler(participantService)
	teamHandler := handlers.NewTeamHandler(teamService)
	registrationHandler := handlers.NewRegistrationHandler(registrationService)

	// Routes protégées par authentification
	api := e.Group("/api", authMiddleware)

	// Hackathon routes
	api.POST("/hackathons", hackathonHandler.Create)
	api.GET("/hackathons", hackathonHandler.List)
	api.GET("/hackathons/:id", hackathonHandler.GetByID)
	api.PUT("/hackathons/:id", hackathonHandler.Update)
	api.DELETE("/hackathons/:id", hackathonHandler.Delete)

	// Participant routes
	api.POST("/participants", participantHandler.Create)
	api.GET("/participants", participantHandler.List)
	api.GET("/participants/:id", participantHandler.GetByID)
	api.PUT("/participants/:id", participantHandler.Update)
	api.DELETE("/participants/:id", participantHandler.Delete)

	// Team routes
	api.POST("/teams", teamHandler.Create)
	api.GET("/teams", teamHandler.List)
	api.GET("/teams/:id", teamHandler.GetByID)
	api.PUT("/teams/:id", teamHandler.Update)
	api.DELETE("/teams/:id", teamHandler.Delete)

	// Registration routes
	api.POST("/registrations", registrationHandler.Register)
	api.GET("/registrations", registrationHandler.List)
	api.GET("/registrations/:id", registrationHandler.GetByID)
	api.DELETE("/registrations/:id", registrationHandler.Delete)

	// Healthcheck route
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})

	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
}
