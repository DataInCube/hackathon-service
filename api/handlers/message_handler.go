package handlers

import (
	"net/http"
	"strconv"

	"github.com/DataInCube/hackathon-service/api/services"
	"github.com/DataInCube/hackathon-service/internal/models"
	"github.com/labstack/echo/v4"
)

type MessageHandler struct {
	Service *services.MessageService
}

func NewMessageHandler(service *services.MessageService) *MessageHandler {
	return &MessageHandler{Service: service}
}

// POST /messages
func (h *MessageHandler) Create(c echo.Context) error {
	var m models.Message
	if err := c.Bind(&m); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid request"})
	}

	id, err := h.Service.Create(c.Request().Context(), m)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "could not create message"})
	}

	m.ID = uint(id)
	return c.JSON(http.StatusCreated, m)
}

// GET /messages/:id
func (h *MessageHandler) GetByID(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid message id"})
	}

	msg, err := h.Service.GetByID(c.Request().Context(), uint(id))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "error fetching message"})
	}
	if msg == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "message not found"})
	}
	return c.JSON(http.StatusOK, msg)
}

// GET /messages/hackathon/:hackathonID
func (h *MessageHandler) GetByHackathon(c echo.Context) error {
	hackathonID, err := strconv.Atoi(c.Param("hackathonID"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid hackathon id"})
	}

	messages, err := h.Service.GetByHackathon(c.Request().Context(), uint(hackathonID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "could not fetch messages"})
	}
	return c.JSON(http.StatusOK, messages)
}

// GET /messages/team/:teamID
func (h *MessageHandler) GetByTeam(c echo.Context) error {
	teamID, err := strconv.Atoi(c.Param("teamID"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid team id"})
	}

	messages, err := h.Service.GetByTeam(c.Request().Context(), uint(teamID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "could not fetch messages"})
	}
	return c.JSON(http.StatusOK, messages)
}

// GET /messages
func (h *MessageHandler) List(c echo.Context) error {
	messages, err := h.Service.List(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "could not fetch messages"})
	}
	return c.JSON(http.StatusOK, messages)
}

// DELETE /messages/:id
func (h *MessageHandler) Delete(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "invalid message id"})
	}

	if err := h.Service.Delete(c.Request().Context(), uint(id)); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "could not delete message"})
	}
	return c.NoContent(http.StatusNoContent)
}