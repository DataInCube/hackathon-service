package handlers

import (
	"net/http"
	"strconv"

	"github.com/DataInCube/hackathon-service/api/services"
	"github.com/DataInCube/hackathon-service/internal/models"
	"github.com/DataInCube/hackathon-service/pkg/errors"
	"github.com/labstack/echo/v4"
)

type HackathonHandler struct {
	Service *services.HackathonService
}

func NewHackathonHandler(s *services.HackathonService) *HackathonHandler {
	return &HackathonHandler{Service: s}
}

// CreateHackathon godoc
// @Summary Create a new hackathon
// @Description Create a new hackathon with JSON payload
// @Tags Hackathon
// @Accept  json
// @Produce  json
// @Param hackathon body models.Hackathon true "Hackathon payload"
// @Success 201 {object} models.Hackathon
// @Failure 400 {object} errors.HTTPError
// @Router /hackathons [post]
func (h *HackathonHandler) Create(c echo.Context) error {
	var input models.Hackathon
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, errors.HTTPError{Code: http.StatusBadRequest, Message: err.Error()})
	}
	if _, err := h.Service.CreateHackathon(c.Request().Context(), input); err != nil {
		return c.JSON(http.StatusInternalServerError, errors.HTTPError{Code: http.StatusInternalServerError, Message: err.Error()})
	}
	return c.JSON(http.StatusCreated, input)
}

// ListHackathons godoc
// @Summary List all hackathons
// @Description Get a list of all hackathons
// @Tags Hackathon
// @Produce  json
// @Success 200 {array} models.Hackathon
// @Failure 500 {object} models.HTTPError
// @Router /hackathons [get]
func (h *HackathonHandler) List(c echo.Context) error {
	hackathons, err := h.Service.GetAllHackathons(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errors.HTTPError{Code: http.StatusInternalServerError, Message: err.Error()})
	}
	return c.JSON(http.StatusOK, hackathons)
}

// GetHackathonByID godoc
// @Summary Get a hackathon by ID
// @Description Get a hackathon by its ID
// @Tags Hackathon
// @Produce  json
// @Param id path int true "Hackathon ID"
// @Success 200 {object} models.Hackathon
// @Failure 404 {object} models.HTTPError
// @Failure 500 {object} models.HTTPError
// @Router /hackathons/{id} [get]
func (h *HackathonHandler) GetByID(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	hackathon, err := h.Service.GetHackathonByID(c.Request().Context(), uint(id))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errors.HTTPError{Code: http.StatusInternalServerError, Message: err.Error()})
	}
	if hackathon == nil {
		return c.JSON(http.StatusNotFound, errors.HTTPError{Code: http.StatusNotFound, Message: "Hackathon not found"})
	}
	return c.JSON(http.StatusOK, hackathon)
}

// UpdateHackathon godoc
// @Summary Update a hackathon by ID
// @Description Update a hackathon by its ID with JSON payload
// @Tags Hackathon
// @Accept  json
// @Produce  json
// @Param id path int true "Hackathon ID"
// @Param hackathon body models.Hackathon true "Hackathon payload"
// @Success 200 {object} models.Hackathon
// @Failure 400 {object} models.HTTPError
// @Failure 404 {object} models.HTTPError
// @Failure 500 {object} models.HTTPError
// @Router /hackathons/{id} [put]
func (h *HackathonHandler) Update(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var input models.Hackathon
	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, errors.HTTPError{Code: http.StatusBadRequest, Message: err.Error()})
	}
	input.ID = uint(id)
	if err := h.Service.UpdateHackathon(c.Request().Context(), input); err != nil {
		return c.JSON(http.StatusInternalServerError, errors.HTTPError{Code: http.StatusInternalServerError, Message: err.Error()})
	}
	return c.JSON(http.StatusOK, input)
}

// DeleteHackathon godoc
// @Summary Delete a hackathon by ID
// @Description Delete a hackathon by its ID
// @Tags Hackathon
// @Param id path int true "Hackathon ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} models.HTTPError
// @Failure 500 {object} models.HTTPError
// @Router /hackathons/{id} [delete]
func (h *HackathonHandler) Delete(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.Service.DeleteHackathon(c.Request().Context(), uint(id)); err != nil {
		return c.JSON(http.StatusInternalServerError, errors.HTTPError{Code: http.StatusInternalServerError, Message: err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Deleted"})
}
