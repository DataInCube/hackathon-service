package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/DataInCube/hackathon-service/api/services"
	"github.com/DataInCube/hackathon-service/internal/models"
	"github.com/DataInCube/hackathon-service/pkg/events"
	"github.com/labstack/echo/v4"
)

type HackathonHandler struct {
	Service   *services.HackathonService
	Governance *services.GovernanceService
	Publisher events.Publisher
}

func NewHackathonHandler(service *services.HackathonService, governance *services.GovernanceService, publisher events.Publisher) *HackathonHandler {
	return &HackathonHandler{Service: service, Governance: governance, Publisher: publisher}
}

func (h *HackathonHandler) Create(c echo.Context) error {
	var input models.Hackathon
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	actorID := actorIDFromContext(c)
	created, err := h.Service.Create(c.Request().Context(), input, actorID)
	if err != nil {
		return handleServiceError(err)
	}

	h.emit(c, "hackathon.created", map[string]any{
		"hackathon_id": created.ID,
		"state":        created.State,
	})
	h.audit(c, created.ID, actorID, "hackathon.created", created)

	return c.JSON(http.StatusCreated, created)
}

func (h *HackathonHandler) List(c echo.Context) error {
	limit, offset, err := parseLimitOffset(c)
	if err != nil {
		return err
	}
	items, err := h.Service.List(c.Request().Context(), limit, offset)
	if err != nil {
		return handleServiceError(err)
	}
	return c.JSON(http.StatusOK, items)
}

func (h *HackathonHandler) GetByID(c echo.Context) error {
	id, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	item, err := h.Service.GetByID(c.Request().Context(), id)
	if err != nil {
		return handleServiceError(err)
	}
	if item == nil {
		return echo.NewHTTPError(http.StatusNotFound, "hackathon not found")
	}
	return c.JSON(http.StatusOK, item)
}

func (h *HackathonHandler) Update(c echo.Context) error {
	id, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	var input models.Hackathon
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	updated, err := h.Service.Update(c.Request().Context(), id, input)
	if err != nil {
		return handleServiceError(err)
	}
	h.audit(c, id, actorIDFromContext(c), "hackathon.updated", updated)
	return c.JSON(http.StatusOK, updated)
}

func (h *HackathonHandler) Delete(c echo.Context) error {
	id, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	if err := h.Service.Delete(c.Request().Context(), id); err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "hackathon not found to delete")
		}
		return handleServiceError(err)
	}
	h.audit(c, id, actorIDFromContext(c), "hackathon.deleted", map[string]string{"id": id})
	return c.JSON(http.StatusOK, map[string]string{"message": "deleted"})
}

func (h *HackathonHandler) Publish(c echo.Context) error {
	id, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	updated, err := h.Service.Publish(c.Request().Context(), id)
	if err != nil {
		return handleServiceError(err)
	}
	h.emit(c, "hackathon.published", map[string]any{"hackathon_id": updated.ID, "state": updated.State})
	h.audit(c, id, actorIDFromContext(c), "hackathon.published", updated)
	return c.JSON(http.StatusOK, updated)
}

func (h *HackathonHandler) Transition(c echo.Context) error {
	id, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	var payload struct {
		TargetPhase string `json:"target_phase"`
	}
	if err := c.Bind(&payload); err != nil || payload.TargetPhase == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "target_phase is required")
	}
	updated, err := h.Service.Transition(c.Request().Context(), id, payload.TargetPhase)
	if err != nil {
		return handleServiceError(err)
	}
	h.emit(c, "hackathon.phase.changed", map[string]any{"hackathon_id": updated.ID, "state": updated.State})
	if updated.State == models.HackathonStateLive && updated.RequiresTeams {
		h.emit(c, "hackathon.team.required", map[string]any{"hackathon_id": updated.ID})
	}
	if updated.State == models.HackathonStateSubmissionFrozen {
		h.emit(c, "hackathon.team.locked", map[string]any{"hackathon_id": updated.ID})
	}
	if updated.State == models.HackathonStateCompleted {
		h.emit(c, "hackathon.completed", map[string]any{"hackathon_id": updated.ID})
	}
	h.audit(c, id, actorIDFromContext(c), "hackathon.phase.changed", updated)
	return c.JSON(http.StatusOK, updated)
}

func (h *HackathonHandler) GetState(c echo.Context) error {
	id, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	state, err := h.Service.GetState(c.Request().Context(), id)
	if err != nil {
		return handleServiceError(err)
	}
	return c.JSON(http.StatusOK, map[string]string{"state": state})
}

func (h *HackathonHandler) TeamPolicy(c echo.Context) error {
	id, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	policy, err := h.Service.GetTeamPolicy(c.Request().Context(), id)
	if err != nil {
		return handleServiceError(err)
	}
	return c.JSON(http.StatusOK, policy)
}

func (h *HackathonHandler) LeaderboardPolicy(c echo.Context) error {
	id, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	policy, err := h.Service.GetLeaderboardPolicy(c.Request().Context(), id)
	if err != nil {
		return handleServiceError(err)
	}
	return c.JSON(http.StatusOK, policy)
}

func (h *HackathonHandler) ValidateTeam(c echo.Context) error {
	id, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	var payload struct {
		TeamID      string `json:"team_id"`
		MemberCount *int   `json:"member_count,omitempty"`
	}
	if err := c.Bind(&payload); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	policy, err := h.Service.GetTeamPolicy(c.Request().Context(), id)
	if err != nil {
		return handleServiceError(err)
	}
	allowed := true
	reason := ""
	if policy.RequiresTeams && payload.TeamID == "" {
		allowed = false
		reason = "team required"
	}
	if !policy.AllowsTeams && payload.TeamID != "" {
		allowed = false
		reason = "teams not allowed"
	}
	if payload.MemberCount != nil {
		if policy.MinTeamSize > 0 && *payload.MemberCount < policy.MinTeamSize {
			allowed = false
			reason = "team too small"
		}
		if policy.MaxTeamSize > 0 && *payload.MemberCount > policy.MaxTeamSize {
			allowed = false
			reason = "team too large"
		}
	}
	return c.JSON(http.StatusOK, map[string]any{
		"allowed": allowed,
		"reason":  reason,
		"policy":  policy,
	})
}

func (h *HackathonHandler) FreezeLeaderboard(c echo.Context) error {
	id, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	updated, err := h.Service.FreezeLeaderboard(c.Request().Context(), id)
	if err != nil {
		return handleServiceError(err)
	}
	h.emit(c, "leaderboard.freeze.requested", map[string]any{"hackathon_id": updated.ID})
	h.audit(c, id, actorIDFromContext(c), "leaderboard.freeze.requested", updated)
	return c.JSON(http.StatusOK, updated)
}

func (h *HackathonHandler) PublishLeaderboard(c echo.Context) error {
	id, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	updated, err := h.Service.PublishLeaderboard(c.Request().Context(), id)
	if err != nil {
		return handleServiceError(err)
	}
	h.emit(c, "leaderboard.publish.requested", map[string]any{"hackathon_id": updated.ID})
	h.audit(c, id, actorIDFromContext(c), "leaderboard.publish.requested", updated)
	return c.JSON(http.StatusOK, updated)
}

func (h *HackathonHandler) UnfreezeLeaderboard(c echo.Context) error {
	id, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	updated, err := h.Service.UnfreezeLeaderboard(c.Request().Context(), id)
	if err != nil {
		return handleServiceError(err)
	}
	h.emit(c, "leaderboard.unfreeze.requested", map[string]any{"hackathon_id": updated.ID})
	h.audit(c, id, actorIDFromContext(c), "leaderboard.unfreeze.requested", updated)
	return c.JSON(http.StatusOK, updated)
}

func (h *HackathonHandler) emit(c echo.Context, subject string, payload any) {
	if h.Publisher == nil {
		return
	}
	if err := h.Publisher.Publish(c.Request().Context(), subject, payload); err != nil {
		c.Logger().Error(err)
	}
}

func (h *HackathonHandler) audit(c echo.Context, hackathonID, actorID, action string, payload any) {
	if h.Governance == nil {
		return
	}
	raw, _ := json.Marshal(payload)
	_ = h.Governance.AppendAudit(c.Request().Context(), models.AuditLog{
		HackathonID: hackathonID,
		ActorID:     actorID,
		Action:      action,
		Payload:     raw,
	})
}

func actorIDFromContext(c echo.Context) string {
	if v, ok := c.Get("user_id").(string); ok {
		return v
	}
	return ""
}
