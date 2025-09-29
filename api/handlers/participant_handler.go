package handlers

import (
	"net/http"
	"strconv"

	"github.com/DataInCube/hackathon-service/api/services"
	"github.com/DataInCube/hackathon-service/internal/models"
	"github.com/DataInCube/hackathon-service/pkg/errors"
	"github.com/labstack/echo/v4"
)

type ParticipantHandler struct {
	Service *services.ParticipantService
}

func NewParticipantHandler(s *services.ParticipantService) *ParticipantHandler {
	return &ParticipantHandler{Service: s}
}

// CreateParticipant godoc
// @Summary Create a new participant
// @Description Create a new participant with JSON payload
// @Tags Participant
// @Accept  json
// @Produce  json
// @Param participant body models.Participant true "Participant payload"
// @Success 201 {object} models.Participant
// @Failure 400 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /participants [post]
func (h *ParticipantHandler) Create(c echo.Context) error {
	var p models.Participant
	if err := c.Bind(&p); err != nil {
		return c.JSON(http.StatusBadRequest, errors.HTTPError{Code: http.StatusBadRequest, Message: err.Error()})
	}
	if _, err := h.Service.Create(c.Request().Context(), p); err != nil {
		return c.JSON(http.StatusInternalServerError, errors.HTTPError{Code: http.StatusInternalServerError, Message: err.Error()})
	}
	return c.JSON(http.StatusCreated, p)
}

// ListParticipants godoc
// @Summary List all participants
// @Description Get a list of all participants
// @Tags Participant
// @Produce  json
// @Success 200 {array} models.Participant
// @Failure 500 {object} errors.HTTPError
// @Router /participants [get]
func (h *ParticipantHandler) List(c echo.Context) error {
	ps, err := h.Service.GetAll(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errors.HTTPError{Code: http.StatusInternalServerError, Message: err.Error()})
	}
	return c.JSON(http.StatusOK, ps)
}

// GetParticipantByID godoc
// @Summary Get a participant by ID
// @Description Get a participant by its ID
// @Tags Participant
// @Produce  json
// @Param id path int true "Participant ID"
// @Success 200 {object} models.Participant
// @Failure 404 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /participants/{id} [get]
func (h *ParticipantHandler) GetByID(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	p, err := h.Service.GetByID(c.Request().Context(), uint(id))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errors.HTTPError{Code: http.StatusInternalServerError, Message: err.Error()})
	}
	if p == nil {
		return c.JSON(http.StatusNotFound, errors.HTTPError{Code: http.StatusNotFound, Message: "Participant not found"})
	}
	return c.JSON(http.StatusOK, p)
}

// UpdateParticipant godoc
// @Summary Update a participant by ID
// @Description Update a participant by its ID with JSON payload
// @Tags Participant
// @Accept  json
// @Produce  json
// @Param id path int true "Participant ID"
// @Param participant body models.Participant true "Participant payload"
// @Success 200 {object} models.Participant
// @Failure 400 {object} errors.HTTPError
// @Failure 404 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /participants/{id} [put]
func (h *ParticipantHandler) Update(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var p models.Participant
	if err := c.Bind(&p); err != nil {
		return c.JSON(http.StatusBadRequest, errors.HTTPError{Code: http.StatusBadRequest, Message: err.Error()})
	}
	p.ID = uint(id)
	if err := h.Service.Update(c.Request().Context(), p); err != nil {
		return c.JSON(http.StatusInternalServerError, errors.HTTPError{Code: http.StatusInternalServerError, Message: err.Error()})
	}
	return c.JSON(http.StatusOK, p)
}

// DeleteParticipant godoc
// @Summary Delete a participant by ID
// @Description Delete a participant by its ID
// @Tags Participant
// @Param id path int true "Participant ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /participants/{id} [delete]
func (h *ParticipantHandler) Delete(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.Service.Delete(c.Request().Context(), uint(id)); err != nil {
		return c.JSON(http.StatusInternalServerError, errors.HTTPError{Code: http.StatusInternalServerError, Message: err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Deleted"})
}
