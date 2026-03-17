package handlers

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

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
	if target == models.SubmissionStatusScored {
		metadataMap := metadataToMap(updated.Metadata)
		secondary := extractSecondaryMetricsFromMetadata(updated.Metadata)
		updatesLeaderboard := extractBoolFromMetadata(updated.Metadata, true, "updates_leaderboard", "updatesLeaderboard")
		payload := map[string]any{
			"submission_id":       updated.ID,
			"hackathon_id":        updated.HackathonID,
			"user_id":             updated.SubmittedBy,
			"source":              "git",
			"is_official":         true,
			"is_practice":         false,
			"updates_leaderboard": updatesLeaderboard,
			"evaluated_at":        time.Now().UTC().Format(time.RFC3339),
			"metadata":            metadataMap,
		}
		if updated.TeamID != nil && *updated.TeamID != "" {
			payload["team_id"] = *updated.TeamID
		}
		if score, ok := extractScoreFromMetadata(updated.Metadata); ok {
			payload["score"] = score
			payload["scores"] = map[string]any{
				"primary":   score,
				"secondary": secondary,
			}
		}
		if metric := extractStringFromMetadata(updated.Metadata, "primary_metric", "metric"); metric != "" {
			payload["primary_metric"] = metric
		}
		if commit := extractStringFromMetadata(updated.Metadata, "commit_sha", "commit", "git_commit"); commit != "" {
			payload["commit_sha"] = commit
		}
		if boardType := extractStringFromMetadata(updated.Metadata, "board_type", "boardType"); boardType != "" {
			payload["board_type"] = boardType
		}
		if submittedAt := extractStringFromMetadata(updated.Metadata, "submitted_at", "submission_time", "created_at"); submittedAt != "" {
			payload["submission_time"] = submittedAt
		}
		if evaluationJobID := extractStringFromMetadata(updated.Metadata, "evaluation_job_id", "job_id"); evaluationJobID != "" {
			payload["evaluation_job_id"] = evaluationJobID
		}
		if practiceJobID := extractStringFromMetadata(updated.Metadata, "practice_job_id", "last_practice_job_id"); practiceJobID != "" {
			payload["practice_job_id"] = practiceJobID
		}
		h.emit(c, "evaluation.completed", payload)
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

func extractScoreFromMetadata(raw json.RawMessage) (float64, bool) {
	if len(raw) == 0 {
		return 0, false
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return 0, false
	}
	if score, ok := numericFrom(payload["score"]); ok {
		return score, true
	}
	if metrics, ok := payload["metrics"].(map[string]any); ok {
		if score, ok := numericFrom(metrics["score"]); ok {
			return score, true
		}
		if score, ok := numericFrom(metrics["primary"]); ok {
			return score, true
		}
		if score, ok := numericFrom(metrics["value"]); ok {
			return score, true
		}
	}
	if evaluation, ok := payload["evaluation"].(map[string]any); ok {
		if score, ok := numericFrom(evaluation["score"]); ok {
			return score, true
		}
	}
	return 0, false
}

func extractStringFromMetadata(raw json.RawMessage, keys ...string) string {
	if len(raw) == 0 {
		return ""
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return ""
	}
	for _, key := range keys {
		if v, ok := payload[key]; ok {
			if s, ok := v.(string); ok {
				return strings.TrimSpace(s)
			}
		}
	}
	if evaluation, ok := payload["evaluation"].(map[string]any); ok {
		for _, key := range keys {
			if v, ok := evaluation[key]; ok {
				if s, ok := v.(string); ok {
					return strings.TrimSpace(s)
				}
			}
		}
	}
	return ""
}

func extractSecondaryMetricsFromMetadata(raw json.RawMessage) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return map[string]any{}
	}
	if scores, ok := payload["scores"].(map[string]any); ok {
		if secondary, ok := scores["secondary"].(map[string]any); ok {
			return secondary
		}
	}
	for _, key := range []string{"efficiency_metrics", "secondary_metrics", "metrics_secondary"} {
		if secondary, ok := payload[key].(map[string]any); ok {
			return secondary
		}
	}
	if evaluation, ok := payload["evaluation"].(map[string]any); ok {
		if secondary, ok := evaluation["secondary"].(map[string]any); ok {
			return secondary
		}
		if secondary, ok := evaluation["efficiency_metrics"].(map[string]any); ok {
			return secondary
		}
	}
	return map[string]any{}
}

func extractBoolFromMetadata(raw json.RawMessage, fallback bool, keys ...string) bool {
	if len(raw) == 0 {
		return fallback
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return fallback
	}
	for _, key := range keys {
		value, ok := payload[key]
		if !ok {
			continue
		}
		switch typed := value.(type) {
		case bool:
			return typed
		case string:
			lower := strings.ToLower(strings.TrimSpace(typed))
			if lower == "true" || lower == "1" || lower == "yes" {
				return true
			}
			if lower == "false" || lower == "0" || lower == "no" {
				return false
			}
		}
	}
	return fallback
}

func metadataToMap(raw json.RawMessage) map[string]any {
	if len(raw) == 0 {
		return map[string]any{}
	}
	var payload map[string]any
	if err := json.Unmarshal(raw, &payload); err != nil {
		return map[string]any{}
	}
	return payload
}

func numericFrom(v any) (float64, bool) {
	switch n := v.(type) {
	case float64:
		return n, true
	case float32:
		return float64(n), true
	case int:
		return float64(n), true
	case int64:
		return float64(n), true
	case json.Number:
		f, err := n.Float64()
		if err == nil {
			return f, true
		}
	}
	return 0, false
}
