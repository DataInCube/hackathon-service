package handlers

import (
	"net/http"
	"strconv"

	"github.com/DataInCube/hackathon-service/api/services"
	"github.com/DataInCube/hackathon-service/internal/models"
	"github.com/labstack/echo/v4"
)

type RegistrationHandler struct {
	Service *services.RegistrationService
}

func NewRegistrationHandler(s *services.RegistrationService) *RegistrationHandler {
	return &RegistrationHandler{Service: s}
}

// Register godoc
// @Summary Register a participant for a hackathon
// @Description Register a participant for a hackathon with JSON payload
// @Tags Registration
// @Accept  json
// @Produce  json
// @Param registration body models.Registration true "Registration payload"
// @Success 201 {object} models.Registration
// @Failure 400 {object} models.HTTPError
// @Failure 500 {object} models.HTTPError
// @Router /registrations [post]
func (h *RegistrationHandler) Register(c echo.Context) error {
	var r models.Registration
	if err := c.Bind(&r); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if err := h.Service.Register(c.Request().Context(), r); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, r)
}

// ListRegistrations godoc
// @Summary List all registrations
// @Description Get a list of all registrations
// @Tags Registration
// @Produce  json
// @Success 200 {array} models.Registration
// @Failure 500 {object} models.HTTPError
// @Router /registrations [get]
func (h *RegistrationHandler) List(c echo.Context) error {
	rs, err := h.Service.GetAll(c.Request().Context())
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, rs)
}

// GetRegistrationByID godoc
// @Summary Get a registration by ID
// @Description Get a registration by its ID
// @Tags Registration
// @Produce  json
// @Param id path int true "Registration ID"
// @Success 200 {object} models.Registration
// @Failure 404 {object} models.HTTPError
// @Failure 500 {object} models.HTTPError
// @Router /registrations/{id} [get]
func (h *RegistrationHandler) GetByID(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	r, err := h.Service.GetByID(c.Request().Context(), uint(id))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if r == nil {
		return echo.NewHTTPError(http.StatusNotFound, "Registration not found")
	}
	return c.JSON(http.StatusOK, r)
}

// DeleteRegistration godoc
// @Summary Delete a registration by ID
// @Description Delete a registration by its ID
// @Tags Registration
// @Param id path int true "Registration ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} models.HTTPError
// @Failure 500 {object} models.HTTPError
// @Router /registrations/{id} [delete]
func (h *RegistrationHandler) Delete(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.Service.Delete(c.Request().Context(), uint(id)); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Deleted"})
}
