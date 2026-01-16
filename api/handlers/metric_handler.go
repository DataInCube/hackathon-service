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

type MetricHandler struct {
	Service    *services.MetricService
	Governance *services.GovernanceService
	Publisher  events.Publisher
}

func NewMetricHandler(service *services.MetricService, governance *services.GovernanceService, publisher events.Publisher) *MetricHandler {
	return &MetricHandler{Service: service, Governance: governance, Publisher: publisher}
}

func (h *MetricHandler) Create(c echo.Context) error {
	hackathonID, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	var input models.EvaluationMetric
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	created, err := h.Service.Create(c.Request().Context(), hackathonID, input)
	if err != nil {
		return handleServiceError(err)
	}
	h.emit(c, "hackathon.metric.created", map[string]any{"hackathon_id": hackathonID, "metric_id": created.ID})
	h.audit(c, hackathonID, actorIDFromContext(c), "hackathon.metric.created", created)
	return c.JSON(http.StatusCreated, created)
}

func (h *MetricHandler) List(c echo.Context) error {
	hackathonID, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	limit, offset, err := parseLimitOffset(c)
	if err != nil {
		return err
	}
	items, err := h.Service.List(c.Request().Context(), hackathonID, limit, offset)
	if err != nil {
		return handleServiceError(err)
	}
	return c.JSON(http.StatusOK, items)
}

func (h *MetricHandler) GetByID(c echo.Context) error {
	hackathonID, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	metricID, err := parseUUIDParam(c, "metricId")
	if err != nil {
		return err
	}
	item, err := h.Service.GetByID(c.Request().Context(), hackathonID, metricID)
	if err != nil {
		return handleServiceError(err)
	}
	if item == nil {
		return echo.NewHTTPError(http.StatusNotFound, "metric not found")
	}
	return c.JSON(http.StatusOK, item)
}

func (h *MetricHandler) Update(c echo.Context) error {
	hackathonID, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	metricID, err := parseUUIDParam(c, "metricId")
	if err != nil {
		return err
	}
	var input services.MetricUpdateInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	updated, err := h.Service.Update(c.Request().Context(), hackathonID, metricID, input)
	if err != nil {
		return handleServiceError(err)
	}
	h.emit(c, "hackathon.metric.updated", map[string]any{"hackathon_id": hackathonID, "metric_id": updated.ID})
	h.audit(c, hackathonID, actorIDFromContext(c), "hackathon.metric.updated", updated)
	return c.JSON(http.StatusOK, updated)
}

func (h *MetricHandler) Delete(c echo.Context) error {
	hackathonID, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	metricID, err := parseUUIDParam(c, "metricId")
	if err != nil {
		return err
	}
	if err := h.Service.Delete(c.Request().Context(), hackathonID, metricID); err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "metric not found to delete")
		}
		return handleServiceError(err)
	}
	h.emit(c, "hackathon.metric.deleted", map[string]any{"hackathon_id": hackathonID, "metric_id": metricID})
	h.audit(c, hackathonID, actorIDFromContext(c), "hackathon.metric.deleted", map[string]string{"id": metricID})
	return c.JSON(http.StatusOK, map[string]string{"message": "deleted"})
}

func (h *MetricHandler) emit(c echo.Context, subject string, payload any) {
	if h.Publisher == nil {
		return
	}
	if err := h.Publisher.Publish(c.Request().Context(), subject, payload); err != nil {
		c.Logger().Error(err)
	}
}

func (h *MetricHandler) audit(c echo.Context, hackathonID, actorID, action string, payload any) {
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
