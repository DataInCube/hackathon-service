package handlers

import (
	"net/http"
	"strconv"

	"github.com/DataInCube/hackathon-service/api/services"
	"github.com/DataInCube/hackathon-service/internal/models"
	"github.com/DataInCube/hackathon-service/pkg/errors"
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
// @Failure 400 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /registrations [post]
func (h *RegistrationHandler) Register(c echo.Context) error {
	var r models.Registration
	if err := c.Bind(&r); err != nil {
		return c.JSON(http.StatusBadRequest, errors.HTTPError{Code: http.StatusBadRequest, Message: err.Error()})
	}
	if _, err := h.Service.Register(c.Request().Context(), r); err != nil {
		return c.JSON(http.StatusInternalServerError, errors.HTTPError{Code: http.StatusInternalServerError, Message: err.Error()})
	}
	return c.JSON(http.StatusCreated, r)
}

// ListRegistrations godoc
// @Summary List all registrations
// @Description Get a list of all registrations
// @Tags Registration
// @Produce  json
// @Success 200 {array} models.Registration
// @Failure 500 {object} errors.HTTPError
// @Router /registrations [get]
func (h *RegistrationHandler) List(c echo.Context) error {
	rs, err := h.Service.GetAllRegistrations(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errors.HTTPError{Code: http.StatusInternalServerError, Message: err.Error()})
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
// @Failure 404 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /registrations/{id} [get]
func (h *RegistrationHandler) GetByID(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	r, err := h.Service.GetRegistrationByID(c.Request().Context(), uint(id))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errors.HTTPError{Code: http.StatusInternalServerError, Message: err.Error()})
	}
	if r == nil {
		return c.JSON(http.StatusNotFound, errors.HTTPError{Code: http.StatusNotFound, Message: "Registration not found"})	
	}
	return c.JSON(http.StatusOK, r)
}

// DeleteRegistration godoc
// @Summary Delete a registration by ID
// @Description Delete a registration by its ID
// @Tags Registration
// @Param id path int true "Registration ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /registrations/{id} [delete]
func (h *RegistrationHandler) Delete(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.Service.DeleteRegistration(c.Request().Context(), uint(id)); err != nil {
		return c.JSON(http.StatusInternalServerError, errors.HTTPError{Code: http.StatusInternalServerError, Message: err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Deleted"})
}
