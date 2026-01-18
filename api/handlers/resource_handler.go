package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/DataInCube/hackathon-service/api/services"
	"github.com/DataInCube/hackathon-service/internal/models"
	"github.com/labstack/echo/v4"
)

type ResourceHandler struct {
	Service    *services.ResourceService
	Governance *services.GovernanceService
}

func NewResourceHandler(service *services.ResourceService, governance *services.GovernanceService) *ResourceHandler {
	return &ResourceHandler{Service: service, Governance: governance}
}

func (h *ResourceHandler) Create(c echo.Context) error {
	hackathonID, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	var input models.Resource
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	created, err := h.Service.Create(c.Request().Context(), hackathonID, input)
	if err != nil {
		return handleServiceError(err)
	}
	h.audit(c, hackathonID, actorIDFromContext(c), "resource.created", created)
	return c.JSON(http.StatusCreated, created)
}

func (h *ResourceHandler) List(c echo.Context) error {
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

func (h *ResourceHandler) GetByID(c echo.Context) error {
	hackathonID, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	resourceID, err := parseUUIDParam(c, "resourceId")
	if err != nil {
		return err
	}
	item, err := h.Service.GetByID(c.Request().Context(), hackathonID, resourceID)
	if err != nil {
		return handleServiceError(err)
	}
	if item == nil {
		return echo.NewHTTPError(http.StatusNotFound, "resource not found")
	}
	return c.JSON(http.StatusOK, item)
}

func (h *ResourceHandler) Update(c echo.Context) error {
	hackathonID, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	resourceID, err := parseUUIDParam(c, "resourceId")
	if err != nil {
		return err
	}
	var input services.ResourceUpdateInput
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	updated, err := h.Service.Update(c.Request().Context(), hackathonID, resourceID, input)
	if err != nil {
		return handleServiceError(err)
	}
	h.audit(c, hackathonID, actorIDFromContext(c), "resource.updated", updated)
	return c.JSON(http.StatusOK, updated)
}

func (h *ResourceHandler) Delete(c echo.Context) error {
	hackathonID, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	resourceID, err := parseUUIDParam(c, "resourceId")
	if err != nil {
		return err
	}
	item, err := h.Service.GetByID(c.Request().Context(), hackathonID, resourceID)
	if err != nil {
		return handleServiceError(err)
	}
	if item == nil {
		return echo.NewHTTPError(http.StatusNotFound, "resource not found to delete")
	}
	if err := h.Service.Delete(c.Request().Context(), hackathonID, resourceID); err != nil {
		if errors.Is(err, services.ErrNotFound) {
			return echo.NewHTTPError(http.StatusNotFound, "resource not found to delete")
		}
		return handleServiceError(err)
	}
	h.audit(c, hackathonID, actorIDFromContext(c), "resource.deleted", map[string]string{"id": resourceID})
	return c.JSON(http.StatusOK, map[string]string{"message": "deleted"})
}

func (h *ResourceHandler) audit(c echo.Context, hackathonID, actorID, action string, payload any) {
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
