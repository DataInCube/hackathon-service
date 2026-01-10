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

type RuleHandler struct {
	Service          *services.RuleService
	Hackathons       *services.HackathonService
	Governance       *services.GovernanceService
	Publisher        events.Publisher
}

func NewRuleHandler(service *services.RuleService, hackathons *services.HackathonService, governance *services.GovernanceService, publisher events.Publisher) *RuleHandler {
	return &RuleHandler{Service: service, Hackathons: hackathons, Governance: governance, Publisher: publisher}
}

func (h *RuleHandler) Create(c echo.Context) error {
	hackathonID, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	var payload struct {
		Name        string          `json:"name"`
		Description string          `json:"description"`
		TrackID     *string         `json:"track_id,omitempty"`
		Content     json.RawMessage `json:"content"`
	}
	if err := c.Bind(&payload); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	ruleInput := models.Rule{
		Name:        payload.Name,
		Description: payload.Description,
		TrackID:     payload.TrackID,
	}
	actorID := actorIDFromContext(c)
	rule, version, err := h.Service.CreateRule(c.Request().Context(), hackathonID, ruleInput, payload.Content, actorID)
	if err != nil {
		return handleServiceError(err)
	}

	h.emit(c, "hackathon.rule.created", map[string]any{
		"hackathon_id": hackathonID,
		"rule_id":      rule.ID,
		"version_id":   version.ID,
	})
	h.audit(c, hackathonID, actorID, "hackathon.rule.created", rule)

	return c.JSON(http.StatusCreated, map[string]any{
		"rule":    rule,
		"version": version,
	})
}

func (h *RuleHandler) ListByHackathon(c echo.Context) error {
	hackathonID, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	limit, offset, err := parseLimitOffset(c)
	if err != nil {
		return err
	}
	items, err := h.Service.ListByHackathon(c.Request().Context(), hackathonID, limit, offset)
	if err != nil {
		return handleServiceError(err)
	}
	return c.JSON(http.StatusOK, items)
}

func (h *RuleHandler) GetByID(c echo.Context) error {
	ruleID, err := parseUUIDParam(c, "ruleId")
	if err != nil {
		return err
	}
	item, err := h.Service.GetByID(c.Request().Context(), ruleID)
	if err != nil {
		return handleServiceError(err)
	}
	if item == nil {
		return echo.NewHTTPError(http.StatusNotFound, "rule not found")
	}
	return c.JSON(http.StatusOK, item)
}

func (h *RuleHandler) CreateVersion(c echo.Context) error {
	ruleID, err := parseUUIDParam(c, "ruleId")
	if err != nil {
		return err
	}
	var payload struct {
		Content json.RawMessage `json:"content"`
	}
	if err := c.Bind(&payload); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	version, err := h.Service.CreateVersion(c.Request().Context(), ruleID, payload.Content, actorIDFromContext(c))
	if err != nil {
		return handleServiceError(err)
	}
	hackathonID := ""
	if rule, err := h.Service.GetByID(c.Request().Context(), ruleID); err == nil && rule != nil {
		hackathonID = rule.HackathonID
	}
	h.audit(c, hackathonID, actorIDFromContext(c), "hackathon.rule.versioned", version)
	return c.JSON(http.StatusCreated, version)
}

func (h *RuleHandler) History(c echo.Context) error {
	ruleID, err := parseUUIDParam(c, "ruleId")
	if err != nil {
		return err
	}
	limit, offset, err := parseLimitOffset(c)
	if err != nil {
		return err
	}
	items, err := h.Service.History(c.Request().Context(), ruleID, limit, offset)
	if err != nil {
		return handleServiceError(err)
	}
	return c.JSON(http.StatusOK, items)
}

func (h *RuleHandler) Update(c echo.Context) error {
	ruleID, err := parseUUIDParam(c, "ruleId")
	if err != nil {
		return err
	}
	var input services.RuleUpdateInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	updated, err := h.Service.Update(c.Request().Context(), ruleID, input)
	if err != nil {
		return handleServiceError(err)
	}
	h.audit(c, updated.HackathonID, actorIDFromContext(c), "hackathon.rule.updated", updated)
	return c.JSON(http.StatusOK, updated)
}

func (h *RuleHandler) Delete(c echo.Context) error {
	ruleID, err := parseUUIDParam(c, "ruleId")
	if err != nil {
		return err
	}
	rule, err := h.Service.GetByID(c.Request().Context(), ruleID)
	if err != nil {
		return handleServiceError(err)
	}
	if rule == nil {
		return echo.NewHTTPError(http.StatusNotFound, "rule not found to delete")
	}
	if err := h.Service.Delete(c.Request().Context(), ruleID); err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "rule not found to delete")
		}
		return handleServiceError(err)
	}
	h.audit(c, rule.HackathonID, actorIDFromContext(c), "hackathon.rule.deleted", map[string]string{"id": ruleID})
	return c.JSON(http.StatusOK, map[string]string{"message": "deleted"})
}

func (h *RuleHandler) Activate(c echo.Context) error {
	hackathonID, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	ruleVersionID, err := parseUUIDParam(c, "ruleVersionId")
	if err != nil {
		return err
	}
	version, err := h.Service.GetVersionByID(c.Request().Context(), ruleVersionID)
	if err != nil {
		return handleServiceError(err)
	}
	if version == nil {
		return echo.NewHTTPError(http.StatusNotFound, "rule version not found")
	}
	if version.Status != models.RuleStatusLocked {
		return echo.NewHTTPError(http.StatusBadRequest, "rule version must be locked")
	}
	ok, err := h.Service.RuleVersionBelongsToHackathon(c.Request().Context(), ruleVersionID, hackathonID)
	if err != nil {
		return handleServiceError(err)
	}
	if !ok {
		return echo.NewHTTPError(http.StatusBadRequest, "rule version does not belong to hackathon")
	}
	if err := h.Hackathons.SetActiveRuleVersion(c.Request().Context(), hackathonID, ruleVersionID); err != nil {
		return handleServiceError(err)
	}

	h.emit(c, "hackathon.rule.activated", map[string]any{
		"hackathon_id":     hackathonID,
		"rule_version_id":  ruleVersionID,
	})
	h.audit(c, hackathonID, actorIDFromContext(c), "hackathon.rule.activated", map[string]any{
		"rule_version_id": ruleVersionID,
	})

	return c.JSON(http.StatusOK, map[string]string{"message": "rule activated"})
}

func (h *RuleHandler) LockVersion(c echo.Context) error {
	ruleVersionID, err := parseUUIDParam(c, "ruleVersionId")
	if err != nil {
		return err
	}
	version, err := h.Service.LockVersion(c.Request().Context(), ruleVersionID)
	if err != nil {
		return handleServiceError(err)
	}

	hackathonID := ""
	if rule, err := h.Service.GetByID(c.Request().Context(), version.RuleID); err == nil && rule != nil {
		hackathonID = rule.HackathonID
	}

	h.emit(c, "hackathon.rule.version.locked", map[string]any{
		"rule_id":        version.RuleID,
		"rule_version_id": version.ID,
	})
	h.audit(c, hackathonID, actorIDFromContext(c), "hackathon.rule.version.locked", version)

	return c.JSON(http.StatusOK, version)
}

func (h *RuleHandler) emit(c echo.Context, subject string, payload any) {
	if h.Publisher == nil {
		return
	}
	if err := h.Publisher.Publish(c.Request().Context(), subject, payload); err != nil {
		c.Logger().Error(err)
	}
}

func (h *RuleHandler) audit(c echo.Context, hackathonID, actorID, action string, payload any) {
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
