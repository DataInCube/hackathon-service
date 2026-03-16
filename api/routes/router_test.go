package routes

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

func hasRoute(routes []*echo.Route, method, path string) bool {
	for _, r := range routes {
		if r.Method == method && r.Path == path {
			return true
		}
	}
	return false
}

func TestRegisterRoutes_RegistersCoreEndpoints(t *testing.T) {
	e := echo.New()
	RegisterRoutes(e, nil, logrus.New(), nil, "hackathon-service", "1.2.3", nil)

	routes := e.Routes()
	checks := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/health"},
		{http.MethodGet, "/ready"},
		{http.MethodGet, "/version"},
		{http.MethodGet, "/metrics"},
		{http.MethodPost, "/api/v1/hackathons"},
		{http.MethodGet, "/api/v1/hackathons"},
		{http.MethodPost, "/api/v1/hackathons/:hackathonId/tracks"},
		{http.MethodPost, "/api/v1/hackathons/:hackathonId/rules"},
		{http.MethodPost, "/api/v1/hackathons/:hackathonId/submissions"},
	}

	for _, tc := range checks {
		if !hasRoute(routes, tc.method, tc.path) {
			t.Fatalf("missing route %s %s", tc.method, tc.path)
		}
	}
}

func TestRegisterRoutes_HealthAndVersionHandlers(t *testing.T) {
	e := echo.New()
	RegisterRoutes(e, nil, logrus.New(), nil, "hackathon-service", "1.2.3", nil)

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("health status: want 200 got %d", rec.Code)
	}
	var health map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &health); err != nil {
		t.Fatalf("decode health: %v", err)
	}
	if health["status"] != "ok" {
		t.Fatalf("unexpected health payload: %v", health)
	}

	req = httptest.NewRequest(http.MethodGet, "/version", nil)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("version status: want 200 got %d", rec.Code)
	}
	var version map[string]string
	if err := json.Unmarshal(rec.Body.Bytes(), &version); err != nil {
		t.Fatalf("decode version: %v", err)
	}
	if version["service"] != "hackathon-service" || version["version"] != "1.2.3" {
		t.Fatalf("unexpected version payload: %v", version)
	}
}
