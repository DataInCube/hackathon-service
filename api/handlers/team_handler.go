package handlers

import (
	"net/http"
	"strconv"

	"github.com/DataInCube/hackathon-service/api/services"
	"github.com/DataInCube/hackathon-service/internal/models"
	"github.com/DataInCube/hackathon-service/pkg/errors"
	"github.com/labstack/echo/v4"
)

type TeamHandler struct {
	Service *services.TeamService
}

func NewTeamHandler(s *services.TeamService) *TeamHandler {
	return &TeamHandler{Service: s}
}

// Create godoc
// @Summary Create a team
// @Description Create a team with JSON payload
// @Tags Team
// @Accept  json
// @Produce  json
// @Param team body models.Team true "Team payload"
// @Success 201 {object} models.Team
// @Failure 400 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /teams [post]
func (h *TeamHandler) Create(c echo.Context) error {
	type Req struct {
		Team               models.Team `json:"team"`
		MemberGitUsernames []string    `json:"members"` // GitHub usernames
	}
	var req Req
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, errors.HTTPError{Code: http.StatusBadRequest, Message: err.Error()})
	}
	team, err := h.Service.CreateTeam(c.Request().Context(), req.Team, req.MemberGitUsernames)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errors.HTTPError{Code: http.StatusInternalServerError, Message: err.Error()})
	}
	return c.JSON(http.StatusCreated, team)
}

// List godoc
// @Summary List all teams
// @Description Get a list of all teams
// @Tags Team
// @Produce  json
// @Success 200 {array} models.Team
// @Failure 500 {object} errors.HTTPError
// @Router /teams [get]
func (h *TeamHandler) List(c echo.Context) error {
	ts, err := h.Service.GetAllTeams(c.Request().Context())
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errors.HTTPError{Code: http.StatusInternalServerError, Message: err.Error()})
	}
	return c.JSON(http.StatusOK, ts)
}

// GetByID godoc
// @Summary Get a team by ID
// @Description Get a team by its ID
// @Tags Team
// @Produce  json
// @Param id path int true "Team ID"
// @Success 200 {object} models.Team
// @Failure 404 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /teams/{id} [get]
func (h *TeamHandler) GetByID(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	t, err := h.Service.GetTeamByID(c.Request().Context(), uint(id))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, errors.HTTPError{Code: http.StatusInternalServerError, Message: err.Error()})
	}
	if t == nil {
		return c.JSON(http.StatusNotFound, errors.HTTPError{Code: http.StatusNotFound, Message: "Team not found"})
	}
	return c.JSON(http.StatusOK, t)
}

// Update godoc
// @Summary Update a team by ID
// @Description Update a team by its ID
// @Tags Team
// @Accept  json
// @Produce  json
// @Param id path int true "Team ID"
// @Param team body models.Team true "Team payload"
// @Success 200 {object} models.Team
// @Failure 400 {object} errors.HTTPError
// @Failure 404 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /teams/{id} [put]
func (h *TeamHandler) Update(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	var t models.Team
	if err := c.Bind(&t); err != nil {
		return c.JSON(http.StatusBadRequest, errors.HTTPError{Code: http.StatusBadRequest, Message: err.Error()})
	}
	t.ID = uint(id)
	if err := h.Service.UpdateTeam(c.Request().Context(), t); err != nil {
		return c.JSON(http.StatusInternalServerError, errors.HTTPError{Code: http.StatusInternalServerError, Message: err.Error()})
	}
	return c.JSON(http.StatusOK, t)
}

// Delete godoc
// @Summary Delete a team by ID
// @Description Delete a team by its ID
// @Tags Team
// @Param id path int true "Team ID"
// @Success 200 {object} map[string]string
// @Failure 404 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /teams/{id} [delete]
func (h *TeamHandler) Delete(c echo.Context) error {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.Service.DeleteTeam(c.Request().Context(), uint(id)); err != nil {
		return c.JSON(http.StatusInternalServerError, errors.HTTPError{Code: http.StatusInternalServerError, Message: err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Deleted"})
}

// TransferLead godoc
// @Summary Transfer team lead to another member
// @Description Change the lead of a team
// @Tags Team
// @Accept  json
// @Produce  json
// @Param id path int true "Team ID"
// @Param new_lead body map[string]uint true "New lead ID"
// @Success 200 {object} map[string]string
// @Failure 400 {object} errors.HTTPError
// @Failure 500 {object} errors.HTTPError
// @Router /teams/{id}/transfer-lead [patch]
func (h *TeamHandler) TransferLead(c echo.Context) error {
    teamID, _ := strconv.Atoi(c.Param("id"))

    var payload struct {
        NewLeadID uint `json:"new_lead_id"`
    }
    if err := c.Bind(&payload); err != nil {
        return c.JSON(http.StatusBadRequest, errors.HTTPError{Code: http.StatusBadRequest, Message: err.Error()})
    }

    if err := h.Service.TransferLead(c.Request().Context(), uint(teamID), payload.NewLeadID); err != nil {
        return c.JSON(http.StatusInternalServerError, errors.HTTPError{Code: http.StatusInternalServerError, Message: err.Error()})
    }

    return c.JSON(http.StatusOK, map[string]string{"message": "Team lead transferred successfully"})
}