package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/DataInCube/hackathon-service/internal/handlers"
)

func SetupRouter(h *handlers.Handler) *gin.Engine {
	r := gin.Default()

	r.GET("/hackathons", h.GetHackathons)
	r.POST("/hackathons", h.CreateHackathon)
	r.POST("/hackathons/:id/register", h.RegisterParticipant)

	return r
}