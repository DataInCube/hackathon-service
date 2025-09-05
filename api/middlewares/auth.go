package middlewares

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

type AuthResponse struct {
	Valid  bool     `json:"valid"`
	UserID string   `json:"user_id"`
	Email  string   `json:"email"`
	Roles  []string `json:"roles"`
}

func AuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth := c.Request().Header.Get("Authorization")
			if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing or invalid Authorization header")
			}

			token := strings.TrimPrefix(auth, "Bearer ")
			body, _ := json.Marshal(map[string]string{"token": token})

			resp, err := http.Post("http://keycloak-service:8080/verify", "application/json", bytes.NewBuffer(body))
			if err != nil || resp.StatusCode != 200 {
				return echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
			}
			defer resp.Body.Close()

			var authResp AuthResponse
			if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil || !authResp.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
			}

			// Inject user context
			c.Set("user_id", authResp.UserID)
			c.Set("email", authResp.Email)
			c.Set("roles", authResp.Roles)

			return next(c)
		}
	}
}
