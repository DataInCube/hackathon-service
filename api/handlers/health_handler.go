package handlers

import (
    "net/http"

    "github.com/labstack/echo/v4"
)

// HealthCheck godoc
// @Summary Health check
// @Description Check if service is alive
// @Tags system
// @Produce json
// @Success 200 {object} map[string]string
// @Router /health [get]
func HealthCheck(c echo.Context) error {
    return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

// ReadinessCheck godoc
// @Summary Readiness check
// @Description Check if service is ready (DB connection etc.)
// @Tags system
// @Produce json
// @Success 200 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /ready [get]
func ReadinessCheck(c echo.Context) error {
    // Ici tu pourrais tester la connexion DB
    return c.JSON(http.StatusOK, map[string]string{"status": "ready"})
}