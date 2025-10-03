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

// RegisterIndividual godoc
// @Summary Register an individual participant
// @Description Register a participant without a team
// @Tags Registration
// @Accept  json
// @Produce  json
// @Param participant_id query int true "Participant ID"
// @Param hackathon_id query int true "Hackathon ID"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /registrations/individual [post]
func (h *RegistrationHandler) RegisterIndividual(c echo.Context) error {
	participantID, _ := strconv.Atoi(c.QueryParam("participant_id"))
	hackathonID, _ := strconv.Atoi(c.QueryParam("hackathon_id"))

	id, err := h.Service.RegisterIndividual(c.Request().Context(), uint(participantID), uint(hackathonID))
	if err != nil {
		return c.JSON(http.StatusBadRequest, errors.HTTPError{Code: http.StatusBadRequest, Message: err.Error()})
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{"id": id})
}

// RegisterToTeam godoc
// @Summary Register a participant to a team
// @Description Register a participant to a hackathon team
// @Tags Registration
// @Accept  json
// @Produce  json
// @Param participant_id query int true "Participant ID"
// @Param hackathon_id query int true "Hackathon ID"
// @Param team_id query int true "Team ID"
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /registrations/team [post]
func (h *RegistrationHandler) RegisterToTeam(c echo.Context) error {
	participantID, _ := strconv.Atoi(c.QueryParam("participant_id"))
	hackathonID, _ := strconv.Atoi(c.QueryParam("hackathon_id"))
	teamID, _ := strconv.Atoi(c.QueryParam("team_id"))

	id, err := h.Service.RegisterToTeam(c.Request().Context(), uint(participantID), uint(hackathonID), uint(teamID))
	if err != nil {
		return c.JSON(http.StatusBadRequest, errors.HTTPError{Code: http.StatusBadRequest, Message: err.Error()})
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{"id": id})
}

// ApproveTeamJoin godoc
// @Summary Approve a participant to join team
// @Description Team leader validates a join request
// @Tags Registration
// @Accept  json
// @Produce  json
// @Param team_lead_id query int true "Team Leader ID"
// @Param participant_id query int true "Participant ID"
// @Param team_id query int true "Team ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /registrations/team/approve [post]
func (h *RegistrationHandler) ApproveTeamJoin(c echo.Context) error {
	teamLeadID, _ := strconv.Atoi(c.QueryParam("team_lead_id"))
	participantID, _ := strconv.Atoi(c.QueryParam("participant_id"))
	teamID, _ := strconv.Atoi(c.QueryParam("team_id"))

	err := h.Service.ApproveTeamJoin(c.Request().Context(), uint(teamLeadID), uint(participantID), uint(teamID))
	if err != nil {
		return c.JSON(http.StatusBadRequest, errors.HTTPError{Code: http.StatusBadRequest, Message: err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Participant approved"})
}

// GetRegistrationsByHackathon godoc
// @Summary Get registrations by hackathon
// @Description List all registrations for a given hackathon
// @Tags Registration
// @Produce  json
// @Param hackathon_id path int true "Hackathon ID"
// @Success 200 {array} models.Registration
// @Failure 500 {object} errors.HTTPError
// @Router /registrations/hackathon/{hackathon_id} [get]
func (h *RegistrationHandler) GetRegistrationsByHackathon(c echo.Context) error {
	hackathonID, _ := strconv.Atoi(c.Param("hackathon_id"))

	rs, err := h.Service.GetRegistrationsByHackathon(c.Request().Context(), uint(hackathonID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errors.HTTPError{Code: http.StatusInternalServerError, Message: err.Error()})
	}
	return c.JSON(http.StatusOK, rs)
}

// GetRegistrationByParticipant godoc
// @Summary Get registration by participant in hackathon
// @Description Get a registration for a participant in a specific hackathon
// @Tags Registration
// @Produce  json
// @Param participant_id query int true "Participant ID"
// @Param hackathon_id query int true "Hackathon ID"
// @Success 200 {object} models.Registration
// @Failure 404 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /registrations/by-participant [get]
func (h *RegistrationHandler) GetRegistrationByParticipant(c echo.Context) error {
	participantID, _ := strconv.Atoi(c.QueryParam("participant_id"))
	hackathonID, _ := strconv.Atoi(c.QueryParam("hackathon_id"))

	r, err := h.Service.GetRegistrationByParticipant(c.Request().Context(), uint(participantID), uint(hackathonID))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errors.HTTPError{Code: http.StatusInternalServerError, Message: err.Error()})
	}
	if r == nil {
		return c.JSON(http.StatusNotFound, errors.HTTPError{Code: http.StatusNotFound, Message: "Registration not found"})
	}
	return c.JSON(http.StatusOK, r)
}
