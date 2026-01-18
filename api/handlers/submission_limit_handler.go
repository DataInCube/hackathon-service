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

type SubmissionLimitHandler struct {
	Service    *services.SubmissionLimitService
	Governance *services.GovernanceService
	Publisher  events.Publisher
}

func NewSubmissionLimitHandler(service *services.SubmissionLimitService, governance *services.GovernanceService, publisher events.Publisher) *SubmissionLimitHandler {
	return &SubmissionLimitHandler{Service: service, Governance: governance, Publisher: publisher}
}

func (h *SubmissionLimitHandler) Create(c echo.Context) error {
	hackathonID, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	var input models.SubmissionLimit
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	created, err := h.Service.Create(c.Request().Context(), hackathonID, input)
	if err != nil {
		return handleServiceError(err)
	}
	h.emit(c, "hackathon.submission_limits.created", map[string]any{"hackathon_id": hackathonID})
	h.audit(c, hackathonID, actorIDFromContext(c), "hackathon.submission_limits.created", created)
	return c.JSON(http.StatusCreated, created)
}

func (h *SubmissionLimitHandler) Get(c echo.Context) error {
	hackathonID, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	item, err := h.Service.Get(c.Request().Context(), hackathonID)
	if err != nil {
		return handleServiceError(err)
	}
	if item == nil {
		return echo.NewHTTPError(http.StatusNotFound, "submission limits not found")
	}
	return c.JSON(http.StatusOK, item)
}

func (h *SubmissionLimitHandler) Update(c echo.Context) error {
	hackathonID, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	var input services.SubmissionLimitUpdateInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	updated, err := h.Service.Update(c.Request().Context(), hackathonID, input)
	if err != nil {
		return handleServiceError(err)
	}
	h.emit(c, "hackathon.submission_limits.updated", map[string]any{"hackathon_id": hackathonID})
	h.audit(c, hackathonID, actorIDFromContext(c), "hackathon.submission_limits.updated", updated)
	return c.JSON(http.StatusOK, updated)
}

func (h *SubmissionLimitHandler) Delete(c echo.Context) error {
	hackathonID, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	if err := h.Service.Delete(c.Request().Context(), hackathonID); err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "submission limits not found to delete")
		}
		return handleServiceError(err)
	}
	h.emit(c, "hackathon.submission_limits.deleted", map[string]any{"hackathon_id": hackathonID})
	h.audit(c, hackathonID, actorIDFromContext(c), "hackathon.submission_limits.deleted", map[string]string{"hackathon_id": hackathonID})
	return c.JSON(http.StatusOK, map[string]string{"message": "deleted"})
}

func (h *SubmissionLimitHandler) emit(c echo.Context, subject string, payload any) {
	if h.Publisher == nil {
		return
	}
	if err := h.Publisher.Publish(c.Request().Context(), subject, payload); err != nil {
		c.Logger().Error(err)
	}
}

func (h *SubmissionLimitHandler) audit(c echo.Context, hackathonID, actorID, action string, payload any) {
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
