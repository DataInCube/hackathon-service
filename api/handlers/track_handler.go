package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/DataInCube/hackathon-service/api/services"
	"github.com/DataInCube/hackathon-service/internal/models"
	"github.com/labstack/echo/v4"
)

type TrackHandler struct {
	Service    *services.TrackService
	Governance *services.GovernanceService
}

func NewTrackHandler(service *services.TrackService, governance *services.GovernanceService) *TrackHandler {
	return &TrackHandler{Service: service, Governance: governance}
}

func (h *TrackHandler) Create(c echo.Context) error {
	hackathonID, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	var input models.Track
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	created, err := h.Service.Create(c.Request().Context(), hackathonID, input)
	if err != nil {
		return handleServiceError(err)
	}
	h.audit(c, hackathonID, actorIDFromContext(c), "track.created", created)
	return c.JSON(http.StatusCreated, created)
}

func (h *TrackHandler) GetByID(c echo.Context) error {
	hackathonID, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	trackID, err := parseUUIDParam(c, "trackId")
	if err != nil {
		return err
	}
	item, err := h.Service.GetByID(c.Request().Context(), hackathonID, trackID)
	if err != nil {
		return handleServiceError(err)
	}
	if item == nil {
		return echo.NewHTTPError(http.StatusNotFound, "track not found")
	}
	return c.JSON(http.StatusOK, item)
}

func (h *TrackHandler) List(c echo.Context) error {
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

func (h *TrackHandler) Update(c echo.Context) error {
	hackathonID, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	trackID, err := parseUUIDParam(c, "trackId")
	if err != nil {
		return err
	}
	var input services.TrackUpdateInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	updated, err := h.Service.Update(c.Request().Context(), hackathonID, trackID, input)
	if err != nil {
		return handleServiceError(err)
	}
	h.audit(c, hackathonID, actorIDFromContext(c), "track.updated", updated)
	return c.JSON(http.StatusOK, updated)
}

func (h *TrackHandler) Delete(c echo.Context) error {
	hackathonID, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	trackID, err := parseUUIDParam(c, "trackId")
	if err != nil {
		return err
	}
	item, err := h.Service.GetByID(c.Request().Context(), hackathonID, trackID)
	if err != nil {
		return handleServiceError(err)
	}
	if item == nil {
		return echo.NewHTTPError(http.StatusNotFound, "track not found to delete")
	}
	if err := h.Service.Delete(c.Request().Context(), hackathonID, trackID); err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "track not found to delete")
		}
		return handleServiceError(err)
	}
	h.audit(c, hackathonID, actorIDFromContext(c), "track.deleted", map[string]string{"id": trackID})
	return c.JSON(http.StatusOK, map[string]string{"message": "deleted"})
}

func (h *TrackHandler) audit(c echo.Context, hackathonID, actorID, action string, payload any) {
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
