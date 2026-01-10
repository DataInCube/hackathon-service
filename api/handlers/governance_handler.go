package handlers

import (
	"net/http"

	"github.com/DataInCube/hackathon-service/api/services"
	"github.com/DataInCube/hackathon-service/internal/models"
	"github.com/labstack/echo/v4"
)

type GovernanceHandler struct {
	Service *services.GovernanceService
}

func NewGovernanceHandler(service *services.GovernanceService) *GovernanceHandler {
	return &GovernanceHandler{Service: service}
}

func (h *GovernanceHandler) CreateReport(c echo.Context) error {
	hackathonID, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	var input models.Report
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	report, err := h.Service.CreateReport(c.Request().Context(), hackathonID, input, actorIDFromContext(c))
	if err != nil {
		return handleServiceError(err)
	}
	return c.JSON(http.StatusCreated, report)
}

func (h *GovernanceHandler) CreateAppeal(c echo.Context) error {
	var input models.Appeal
	if err := c.Bind(&input); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	appeal, err := h.Service.CreateAppeal(c.Request().Context(), input, actorIDFromContext(c))
	if err != nil {
		return handleServiceError(err)
	}
	return c.JSON(http.StatusCreated, appeal)
}

func (h *GovernanceHandler) AuditHackathon(c echo.Context) error {
	hackathonID, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		return err
	}
	limit, offset, err := parseLimitOffset(c)
	if err != nil {
		return err
	}
	items, err := h.Service.AuditLogs(c.Request().Context(), hackathonID, limit, offset)
	if err != nil {
		return handleServiceError(err)
	}
	return c.JSON(http.StatusOK, items)
}
