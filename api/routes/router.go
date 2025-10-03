package routes

import (
	"database/sql"

	"github.com/DataInCube/hackathon-service/api/handlers"
	"github.com/DataInCube/hackathon-service/api/middlewares"
	"github.com/DataInCube/hackathon-service/api/services"
	"github.com/DataInCube/hackathon-service/pkg/utils"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func RegisterRoutes(e *echo.Echo, db *sql.DB, logger *logrus.Logger) {
	// Auth middleware (appelle keycloak-service)
	authMiddleware := middlewares.AuthMiddleware()

	//kronos client
	kronosClient := &utils.KronosClient{
    BaseURL:   "https://kronos.example.com",
    AuthToken: "TON_PAT",
}

	// Injecter les services
	hackathonService := services.NewHackathonService(db)
	participantService := services.NewParticipantService(db)
	teamService := services.NewTeamService(db, kronosClient)
	registrationService := services.NewRegistrationService(db)
	messageService := services.NewMessageService(db)

	// Injecter les handlers
	hackathonHandler := handlers.NewHackathonHandler(hackathonService)
	participantHandler := handlers.NewParticipantHandler(participantService)
	teamHandler := handlers.NewTeamHandler(teamService)
	registrationHandler := handlers.NewRegistrationHandler(registrationService)
	messageHandler := handlers.NewMessageHandler(messageService)



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
	api.PATCH("/teams/:id/transfer-lead", teamHandler.TransferLead)


	// Registration routes
	api.POST("/registrations", registrationHandler.Register)
	api.GET("/registrations", registrationHandler.List)
	api.GET("/registrations/:id", registrationHandler.GetByID)
	api.DELETE("/registrations/:id", registrationHandler.Delete)
	api.GET("/registrations/participant/:participantID", registrationHandler.GetRegistrationByParticipant)
	api.GET("/registrations/hackathon/:hackathonID", registrationHandler.GetRegistrationsByHackathon)
	api.POST("/registrations/team", registrationHandler.RegisterToTeam)
	api.POST("/registrations/approve", registrationHandler.ApproveTeamJoin)
	api.POST("/registrations/register-individual", registrationHandler.RegisterIndividual)

	// Message routes
	api.POST("/messages", messageHandler.Create)
	api.GET("/messages", messageHandler.List)
	api.DELETE("/messages/:id", messageHandler.Delete)
	api.GET("/messages/:id", messageHandler.GetByID)
	api.GET("/messages/hackathon/:hackathonID", messageHandler.GetByHackathon)
	api.GET("/messages/team/:teamID", messageHandler.GetByTeam)



	// Healthcheck route
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})
	e.GET("/healthy", handlers.HealthCheck)
	e.GET("/ready", handlers.ReadinessCheck)

	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
}
