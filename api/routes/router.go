package routes

import (
	"database/sql"

	"github.com/DataInCube/hackathon-service/api/handlers"
	"github.com/DataInCube/hackathon-service/api/middlewares"
	"github.com/DataInCube/hackathon-service/api/services"
	"github.com/DataInCube/hackathon-service/pkg/events"

	"github.com/labstack/echo/v4"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sirupsen/logrus"
)

func RegisterRoutes(e *echo.Echo, db *sql.DB, logger *logrus.Logger, authMiddleware echo.MiddlewareFunc, serviceName, serviceVersion string, publisher events.Publisher) {
	// Middleware global (logger, recover, CORS, etc. à ajouter ici si besoin)

	// Injecter les services
	hackathonService := services.NewHackathonService(db)
	trackService := services.NewTrackService(db)
	ruleService := services.NewRuleService(db)
	submissionService := services.NewSubmissionService(db, trackService)
	resourceService := services.NewResourceService(db)
	datasetService := services.NewDatasetService(db)
	metricService := services.NewMetricService(db)
	submissionLimitService := services.NewSubmissionLimitService(db)
	governanceService := services.NewGovernanceService(db)

	// Injecter les handlers
	hackathonHandler := handlers.NewHackathonHandler(hackathonService, governanceService, publisher)
	trackHandler := handlers.NewTrackHandler(trackService, governanceService)
	ruleHandler := handlers.NewRuleHandler(ruleService, hackathonService, governanceService, publisher)
	submissionHandler := handlers.NewSubmissionHandler(submissionService, governanceService, publisher)
	resourceHandler := handlers.NewResourceHandler(resourceService, governanceService)
	dataHandler := handlers.NewDataHandler(datasetService, governanceService, publisher)
	metricHandler := handlers.NewMetricHandler(metricService, governanceService, publisher)
	submissionLimitHandler := handlers.NewSubmissionLimitHandler(submissionLimitService, governanceService, publisher)
	governanceHandler := handlers.NewGovernanceHandler(governanceService)

	// Routes protégées par authentification
	api := e.Group("/api/v1")
	if authMiddleware != nil {
		api.Use(authMiddleware)
	}

	adminOrOrganizer := middlewares.RequireAnyRole("hackathon_admin", "hackathon_organizer")

	// Hackathon routes
	api.POST("/hackathons", hackathonHandler.Create, adminOrOrganizer)
	api.GET("/hackathons", hackathonHandler.List)
	api.GET("/hackathons/:hackathonId", hackathonHandler.GetByID)
	api.PUT("/hackathons/:hackathonId", hackathonHandler.Update, adminOrOrganizer)
	api.DELETE("/hackathons/:hackathonId", hackathonHandler.Delete, adminOrOrganizer)
	api.POST("/hackathons/:hackathonId/publish", hackathonHandler.Publish, adminOrOrganizer)
	api.POST("/hackathons/:hackathonId/transition", hackathonHandler.Transition, adminOrOrganizer)
	api.GET("/hackathons/:hackathonId/state", hackathonHandler.GetState)

	// Tracks & rules
	api.POST("/hackathons/:hackathonId/tracks", trackHandler.Create, adminOrOrganizer)
	api.GET("/hackathons/:hackathonId/tracks", trackHandler.List)
	api.GET("/hackathons/:hackathonId/tracks/:trackId", trackHandler.GetByID)
	api.PUT("/hackathons/:hackathonId/tracks/:trackId", trackHandler.Update, adminOrOrganizer)
	api.DELETE("/hackathons/:hackathonId/tracks/:trackId", trackHandler.Delete, adminOrOrganizer)
	api.GET("/hackathons/:hackathonId/rules", ruleHandler.ListByHackathon)
	api.POST("/hackathons/:hackathonId/rules", ruleHandler.Create, adminOrOrganizer)
	api.GET("/rules/:ruleId", ruleHandler.GetByID)
	api.PUT("/rules/:ruleId", ruleHandler.Update, adminOrOrganizer)
	api.DELETE("/rules/:ruleId", ruleHandler.Delete, adminOrOrganizer)
	api.POST("/rules/:ruleId/version", ruleHandler.CreateVersion, adminOrOrganizer)
	api.POST("/rules/:ruleId/versions", ruleHandler.CreateVersion, adminOrOrganizer)
	api.POST("/rules/versions/:ruleVersionId/lock", ruleHandler.LockVersion, adminOrOrganizer)
	api.GET("/rules/:ruleId/history", ruleHandler.History)
	api.POST("/hackathons/:hackathonId/rules/:ruleVersionId/activate", ruleHandler.Activate, adminOrOrganizer)

	// Team policy
	api.GET("/hackathons/:hackathonId/team-policy", hackathonHandler.TeamPolicy)
	api.POST("/hackathons/:hackathonId/teams/validate", hackathonHandler.ValidateTeam)

	// Submissions
	api.POST("/hackathons/:hackathonId/submissions", submissionHandler.Create)
	api.GET("/hackathons/:hackathonId/submissions", submissionHandler.ListByHackathon)
	api.GET("/submissions/:submissionId", submissionHandler.GetByID)
	api.PUT("/submissions/:submissionId", submissionHandler.Update)
	api.DELETE("/submissions/:submissionId", submissionHandler.Delete)
	api.POST("/submissions/:submissionId/lock", submissionHandler.Lock, adminOrOrganizer)
	evaluationRole := middlewares.RequireAnyRole("hackathon_admin", "hackathon_organizer", "platform_admin", "evaluation_executor")
	api.POST("/submissions/:submissionId/evaluation/start", submissionHandler.MarkEvaluationRunning, evaluationRole)
	api.POST("/submissions/:submissionId/evaluation/fail", submissionHandler.MarkEvaluationFailed, evaluationRole)
	api.POST("/submissions/:submissionId/evaluation/score", submissionHandler.MarkScored, evaluationRole)
	api.POST("/submissions/:submissionId/invalidate", submissionHandler.Invalidate, adminOrOrganizer)

	// Leaderboard policy
	api.GET("/hackathons/:hackathonId/leaderboard-policy", hackathonHandler.LeaderboardPolicy)
	api.POST("/hackathons/:hackathonId/leaderboard/freeze", hackathonHandler.FreezeLeaderboard, adminOrOrganizer)
	api.POST("/hackathons/:hackathonId/leaderboard/unfreeze", hackathonHandler.UnfreezeLeaderboard, adminOrOrganizer)
	api.POST("/hackathons/:hackathonId/leaderboard/publish", hackathonHandler.PublishLeaderboard, adminOrOrganizer)

	// Resources
	api.GET("/hackathons/:hackathonId/resources", resourceHandler.List)
	api.POST("/hackathons/:hackathonId/resources", resourceHandler.Create, adminOrOrganizer)
	api.GET("/hackathons/:hackathonId/resources/:resourceId", resourceHandler.GetByID)
	api.PUT("/hackathons/:hackathonId/resources/:resourceId", resourceHandler.Update, adminOrOrganizer)
	api.DELETE("/hackathons/:hackathonId/resources/:resourceId", resourceHandler.Delete, adminOrOrganizer)

	// Data (datasets, files, variables)
	api.POST("/hackathons/:hackathonId/data", dataHandler.Create, adminOrOrganizer)
	api.GET("/hackathons/:hackathonId/data", dataHandler.Get)
	api.PUT("/hackathons/:hackathonId/data", dataHandler.Update, adminOrOrganizer)
	api.DELETE("/hackathons/:hackathonId/data", dataHandler.Delete, adminOrOrganizer)
	api.POST("/hackathons/:hackathonId/data/files", dataHandler.CreateFile, adminOrOrganizer)
	api.GET("/hackathons/:hackathonId/data/files", dataHandler.ListFiles)
	api.GET("/hackathons/:hackathonId/data/files/:fileId", dataHandler.GetFile)
	api.PUT("/hackathons/:hackathonId/data/files/:fileId", dataHandler.UpdateFile, adminOrOrganizer)
	api.DELETE("/hackathons/:hackathonId/data/files/:fileId", dataHandler.DeleteFile, adminOrOrganizer)
	api.POST("/hackathons/:hackathonId/data/variables", dataHandler.CreateVariable, adminOrOrganizer)
	api.GET("/hackathons/:hackathonId/data/variables", dataHandler.ListVariables)
	api.GET("/hackathons/:hackathonId/data/variables/:variableId", dataHandler.GetVariable)
	api.PUT("/hackathons/:hackathonId/data/variables/:variableId", dataHandler.UpdateVariable, adminOrOrganizer)
	api.DELETE("/hackathons/:hackathonId/data/variables/:variableId", dataHandler.DeleteVariable, adminOrOrganizer)

	// Evaluation metrics
	api.POST("/hackathons/:hackathonId/metrics", metricHandler.Create, adminOrOrganizer)
	api.GET("/hackathons/:hackathonId/metrics", metricHandler.List)
	api.GET("/hackathons/:hackathonId/metrics/:metricId", metricHandler.GetByID)
	api.PUT("/hackathons/:hackathonId/metrics/:metricId", metricHandler.Update, adminOrOrganizer)
	api.DELETE("/hackathons/:hackathonId/metrics/:metricId", metricHandler.Delete, adminOrOrganizer)

	// Submission limits
	api.POST("/hackathons/:hackathonId/submission-limits", submissionLimitHandler.Create, adminOrOrganizer)
	api.GET("/hackathons/:hackathonId/submission-limits", submissionLimitHandler.Get)
	api.PUT("/hackathons/:hackathonId/submission-limits", submissionLimitHandler.Update, adminOrOrganizer)
	api.DELETE("/hackathons/:hackathonId/submission-limits", submissionLimitHandler.Delete, adminOrOrganizer)

	// Governance & audit
	api.POST("/hackathons/:hackathonId/reports", governanceHandler.CreateReport)
	api.POST("/appeals", governanceHandler.CreateAppeal)
	api.GET("/audit/hackathons/:hackathonId", governanceHandler.AuditHackathon, adminOrOrganizer)

	// Healthcheck route
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"status": "ok"})
	})
	e.GET("/version", func(c echo.Context) error {
		return c.JSON(200, map[string]string{"service": serviceName, "version": serviceVersion})
	})
	e.GET("/ready", func(c echo.Context) error {
		if err := db.Ping(); err != nil {
			return c.JSON(500, map[string]string{"status": "not_ready"})
		}
		return c.JSON(200, map[string]string{"status": "ready"})
	})

	e.GET("/metrics", echo.WrapHandler(promhttp.Handler()))
}
