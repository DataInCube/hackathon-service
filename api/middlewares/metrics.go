package middlewares

import (
	"net/http"
	"time"

	"github.com/DataInCube/hackathon-service/internal/metrics"

	"github.com/labstack/echo/v4"
)

func MetricsMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)
			status := c.Response().Status
			path := c.Path()
			method := c.Request().Method

			metrics.RequestCounter.WithLabelValues(method, path, httpStatus(status)).Inc()
			metrics.RequestDuration.WithLabelValues(method, path).Observe(time.Since(start).Seconds())

			return err
		}
	}
}

func httpStatus(code int) string {
	return http.StatusText(code)
}
