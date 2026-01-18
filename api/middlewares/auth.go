package middlewares

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/MicahParks/keyfunc"
	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

type JWTConfig struct {
	JWKSURL  string
	Issuer   string
	Audience string
	ClientID string
	Required bool
}

func AuthMiddleware(cfg JWTConfig, logger *logrus.Logger) (echo.MiddlewareFunc, error) {
	if !cfg.Required {
		return nil, nil
	}
	if cfg.JWKSURL == "" {
		return nil, errors.New("AUTH_JWKS_URL is required when AUTH_REQUIRED is true")
	}

	jwks, err := keyfunc.Get(cfg.JWKSURL, keyfunc.Options{
		RefreshInterval:  1 * time.Hour,
		RefreshRateLimit: 1 * time.Minute,
		RefreshTimeout:   10 * time.Second,
		RefreshErrorHandler: func(err error) {
			if logger != nil {
				logger.WithError(err).Warn("jwks refresh error")
			}
		},
	})
	if err != nil {
		return nil, err
	}

	parserOpts := []jwt.ParserOption{jwt.WithValidMethods([]string{"RS256"})}
	parser := jwt.NewParser(parserOpts...)

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			auth := c.Request().Header.Get("Authorization")
			if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
				return echo.NewHTTPError(http.StatusUnauthorized, "Missing or invalid Authorization header")
			}

			tokenStr := strings.TrimPrefix(auth, "Bearer ")
			claims := jwt.MapClaims{}
			token, err := parser.ParseWithClaims(tokenStr, claims, jwks.Keyfunc)
			if err != nil || !token.Valid {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token")
			}
			if cfg.Issuer != "" && !claims.VerifyIssuer(cfg.Issuer, true) {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token issuer")
			}
			if cfg.Audience != "" && !claims.VerifyAudience(cfg.Audience, true) {
				return echo.NewHTTPError(http.StatusUnauthorized, "Invalid token audience")
			}

			roles := extractRoles(claims, cfg.ClientID)
			if hasRole(roles, "banned_user") {
				return echo.NewHTTPError(http.StatusForbidden, "User is banned")
			}

			c.Set("user_id", claimString(claims, "sub"))
			c.Set("email", claimString(claims, "email"))
			c.Set("roles", roles)

			return next(c)
		}
	}, nil
}

func RequireAnyRole(roles ...string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			if len(roles) == 0 {
				return next(c)
			}
			userRoles := RolesFromContext(c)
			for _, role := range roles {
				if hasRole(userRoles, role) {
					return next(c)
				}
			}
			return echo.NewHTTPError(http.StatusForbidden, "Forbidden")
		}
	}
}

func RolesFromContext(c echo.Context) []string {
	raw := c.Get("roles")
	if raw == nil {
		return nil
	}
	roles, ok := raw.([]string)
	if !ok {
		return nil
	}
	return roles
}

func extractRoles(claims jwt.MapClaims, clientID string) []string {
	var roles []string

	if realmAccess, ok := claims["realm_access"].(map[string]any); ok {
		if realmRoles, ok := realmAccess["roles"].([]any); ok {
			for _, r := range realmRoles {
				if role, ok := r.(string); ok {
					roles = append(roles, role)
				}
			}
		}
	}

	if clientID == "" {
		return dedupeRoles(roles)
	}

	if resourceAccess, ok := claims["resource_access"].(map[string]any); ok {
		if client, ok := resourceAccess[clientID].(map[string]any); ok {
			if clientRoles, ok := client["roles"].([]any); ok {
				for _, r := range clientRoles {
					if role, ok := r.(string); ok {
						roles = append(roles, role)
					}
				}
			}
		}
	}

	return dedupeRoles(roles)
}

func dedupeRoles(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, r := range in {
		if _, ok := seen[r]; ok {
			continue
		}
		seen[r] = struct{}{}
		out = append(out, r)
	}
	return out
}

func hasRole(roles []string, target string) bool {
	for _, r := range roles {
		if r == target {
			return true
		}
	}
	return false
}

func claimString(claims jwt.MapClaims, key string) string {
	val, ok := claims[key]
	if !ok {
		return ""
	}
	if s, ok := val.(string); ok {
		return s
	}
	return ""
}
