package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/DataInCube/hackathon-service/api/middlewares"
	"github.com/DataInCube/hackathon-service/api/services"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

const (
	defaultLimit = 50
	maxLimit     = 200
)

func parseUUIDParam(c echo.Context, name string) (string, error) {
	raw := c.Param(name)
	if raw == "" {
		return "", echo.NewHTTPError(http.StatusBadRequest, "missing "+name)
	}
	if _, err := uuid.Parse(raw); err != nil {
		return "", echo.NewHTTPError(http.StatusBadRequest, "invalid "+name)
	}
	return raw, nil
}

func parseLimitOffset(c echo.Context) (int, int, error) {
	limit := defaultLimit
	offset := 0

	if raw := c.QueryParam("limit"); raw != "" {
		val, err := strconv.Atoi(raw)
		if err != nil || val <= 0 {
			return 0, 0, echo.NewHTTPError(http.StatusBadRequest, "invalid limit")
		}
		if val > maxLimit {
			val = maxLimit
		}
		limit = val
	}

	if raw := c.QueryParam("offset"); raw != "" {
		val, err := strconv.Atoi(raw)
		if err != nil || val < 0 {
			return 0, 0, echo.NewHTTPError(http.StatusBadRequest, "invalid offset")
		}
		offset = val
	}

	return limit, offset, nil
}

func parseQueryUUID(c echo.Context, name string) (string, error) {
	raw := c.QueryParam(name)
	if raw == "" {
		return "", echo.NewHTTPError(http.StatusBadRequest, "missing "+name)
	}
	if _, err := uuid.Parse(raw); err != nil {
		return "", echo.NewHTTPError(http.StatusBadRequest, "invalid "+name)
	}
	return raw, nil
}

func handleServiceError(err error) *echo.HTTPError {
	switch {
	case errors.Is(err, services.ErrNotFound):
		return echo.NewHTTPError(http.StatusNotFound, err.Error())
	case errors.Is(err, services.ErrConflict):
		return echo.NewHTTPError(http.StatusConflict, err.Error())
	case errors.Is(err, services.ErrInvalid):
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	case errors.Is(err, services.ErrForbidden):
		return echo.NewHTTPError(http.StatusForbidden, err.Error())
	default:
		return echo.NewHTTPError(http.StatusInternalServerError, "internal error")
	}
}

func hasAnyRole(roles []string, targets ...string) bool {
	if len(roles) == 0 || len(targets) == 0 {
		return false
	}
	seen := make(map[string]struct{}, len(roles))
	for _, r := range roles {
		seen[r] = struct{}{}
	}
	for _, target := range targets {
		if _, ok := seen[target]; ok {
			return true
		}
	}
	return false
}

func isAdminOrOrganizer(c echo.Context) bool {
	roles := middlewares.RolesFromContext(c)
	return hasAnyRole(roles, "hackathon_admin", "hackathon_organizer", "platform_admin")
}
