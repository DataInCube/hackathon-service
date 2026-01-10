package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/DataInCube/hackathon-service/api/services"
	"github.com/DataInCube/hackathon-service/internal/models"
	"github.com/DataInCube/hackathon-service/pkg/events"
	"github.com/labstack/echo/v4"
)

type SubmissionHandler struct {
	Service    *services.SubmissionService
	Governance *services.GovernanceService
	Publisher  events.Publisher
}

func NewSubmissionHandler(service *services.SubmissionService, governance *services.GovernanceService, publisher events.Publisher) *SubmissionHandler {
	return &SubmissionHandler{Service: service, Governance: governance, Publisher: publisher}
}

func (h *SubmissionHandler) Create(c echo.Context) error {
	hackathonID, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	var input services.SubmissionInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	sub, err := h.Service.Create(c.Request().Context(), hackathonID, input, actorIDFromContext(c))
	if err != nil {
		return handleServiceError(err)
	}
	h.emit(c, "submission.created", map[string]any{
		"submission_id": sub.ID,
		"hackathon_id":  sub.HackathonID,
		"status":        sub.Status,
	})
	h.audit(c, sub.HackathonID, actorIDFromContext(c), "submission.created", sub)
	return c.JSON(http.StatusCreated, sub)
}

func (h *SubmissionHandler) Update(c echo.Context) error {
	id, err := parseUUIDParam(c, "submissionId")
	if err != nil {
		return err
	}
	existing, err := h.Service.GetByID(c.Request().Context(), id)
	if err != nil {
		return handleServiceError(err)
	}
	if existing == nil {
		return echo.NewHTTPError(http.StatusNotFound, "submission not found to update")
	}
	if !isAdminOrOrganizer(c) && existing.SubmittedBy != actorIDFromContext(c) {
		return echo.NewHTTPError(http.StatusForbidden, "forbidden")
	}

	var input services.SubmissionUpdateInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	updated, err := h.Service.Update(c.Request().Context(), id, input)
	if err != nil {
		return handleServiceError(err)
	}
	h.audit(c, updated.HackathonID, actorIDFromContext(c), "submission.updated", updated)
	return c.JSON(http.StatusOK, updated)
}

func (h *SubmissionHandler) GetByID(c echo.Context) error {
	id, err := parseUUIDParam(c, "submissionId")
	if err != nil {
		return err
	}
	sub, err := h.Service.GetByID(c.Request().Context(), id)
	if err != nil {
		return handleServiceError(err)
	}
	if sub == nil {
		return echo.NewHTTPError(http.StatusNotFound, "submission not found")
	}
	return c.JSON(http.StatusOK, sub)
}

func (h *SubmissionHandler) ListByHackathon(c echo.Context) error {
	hackathonID, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	limit, offset, err := parseLimitOffset(c)
	if err != nil {
		return err
	}
	subs, err := h.Service.ListByHackathon(c.Request().Context(), hackathonID, limit, offset)
	if err != nil {
		return handleServiceError(err)
	}
	return c.JSON(http.StatusOK, subs)
}

func (h *SubmissionHandler) Delete(c echo.Context) error {
	id, err := parseUUIDParam(c, "submissionId")
	if err != nil {
		return err
	}
	existing, err := h.Service.GetByID(c.Request().Context(), id)
	if err != nil {
		return handleServiceError(err)
	}
	if existing == nil {
		return echo.NewHTTPError(http.StatusNotFound, "submission not found to delete")
	}
	if !isAdminOrOrganizer(c) && existing.SubmittedBy != actorIDFromContext(c) {
		return echo.NewHTTPError(http.StatusForbidden, "forbidden")
	}
	if err := h.Service.Delete(c.Request().Context(), id); err != nil {
		return handleServiceError(err)
	}
	h.audit(c, existing.HackathonID, actorIDFromContext(c), "submission.deleted", map[string]string{"id": id})
	return c.JSON(http.StatusOK, map[string]string{"message": "deleted"})
}

func (h *SubmissionHandler) Lock(c echo.Context) error {
	id, err := parseUUIDParam(c, "submissionId")
	if err != nil {
		return err
	}
	sub, err := h.Service.Lock(c.Request().Context(), id)
	if err != nil {
		return handleServiceError(err)
	}
	h.emit(c, "submission.locked", map[string]any{
		"submission_id": sub.ID,
		"hackathon_id":  sub.HackathonID,
		"status":        sub.Status,
	})
	h.audit(c, sub.HackathonID, actorIDFromContext(c), "submission.locked", sub)
	return c.JSON(http.StatusOK, sub)
}

func (h *SubmissionHandler) MarkEvaluationRunning(c echo.Context) error {
	return h.updateEvaluationStatus(c, models.SubmissionStatusEvaluationRunning)
}

func (h *SubmissionHandler) MarkEvaluationFailed(c echo.Context) error {
	return h.updateEvaluationStatus(c, models.SubmissionStatusEvaluationFailed)
}

func (h *SubmissionHandler) MarkScored(c echo.Context) error {
	return h.updateEvaluationStatus(c, models.SubmissionStatusScored)
}

func (h *SubmissionHandler) Invalidate(c echo.Context) error {
	id, err := parseUUIDParam(c, "submissionId")
	if err != nil {
		return err
	}
	sub, err := h.Service.Invalidate(c.Request().Context(), id)
	if err != nil {
		return handleServiceError(err)
	}
	h.emit(c, "submission.invalidated", map[string]any{
		"submission_id": sub.ID,
		"hackathon_id":  sub.HackathonID,
		"status":        sub.Status,
	})
	h.audit(c, sub.HackathonID, actorIDFromContext(c), "submission.invalidated", sub)
	return c.JSON(http.StatusOK, sub)
}

func (h *SubmissionHandler) updateEvaluationStatus(c echo.Context, target string) error {
	id, err := parseUUIDParam(c, "submissionId")
	if err != nil {
		return err
	}
	var payload struct {
		Metadata json.RawMessage `json:"metadata,omitempty"`
	}
	if err := c.Bind(&payload); err != nil && err != io.EOF {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	var metadataPatch *json.RawMessage
	if len(payload.Metadata) > 0 {
		metadataPatch = &payload.Metadata
	}
	updated, err := h.Service.UpdateEvaluationStatus(c.Request().Context(), id, target, metadataPatch)
	if err != nil {
		return handleServiceError(err)
	}
	h.audit(c, updated.HackathonID, actorIDFromContext(c), "submission.evaluation."+target, updated)
	return c.JSON(http.StatusOK, updated)
}

func (h *SubmissionHandler) emit(c echo.Context, subject string, payload any) {
	if h.Publisher == nil {
		return
	}
	if err := h.Publisher.Publish(c.Request().Context(), subject, payload); err != nil {
		c.Logger().Error(err)
	}
}

func (h *SubmissionHandler) audit(c echo.Context, hackathonID, actorID, action string, payload any) {
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
