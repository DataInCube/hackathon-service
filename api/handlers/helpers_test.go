package handlers

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/DataInCube/hackathon-service/api/services"
	"github.com/labstack/echo/v4"
)

func newHandlerContext(method, target string) echo.Context {
	e := echo.New()
	req := httptest.NewRequest(method, target, nil)
	rec := httptest.NewRecorder()
	return e.NewContext(req, rec)
}

func TestParseUUIDParam(t *testing.T) {
	validID := "6c6bee1e-990e-4f13-a966-4b77d90f9a89"

	c := newHandlerContext(http.MethodGet, "/")
	c.SetPath("/api/v1/hackathons/:hackathonId")
	c.SetParamNames("hackathonId")
	c.SetParamValues(validID)

	got, err := parseUUIDParam(c, "hackathonId")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got != validID {
		t.Fatalf("expected %q, got %q", validID, got)
	}
}

func TestParseUUIDParam_Missing(t *testing.T) {
	c := newHandlerContext(http.MethodGet, "/")
	c.SetPath("/api/v1/hackathons/:hackathonId")
	c.SetParamNames("hackathonId")
	c.SetParamValues("")

	_, err := parseUUIDParam(c, "hackathonId")
	httpErr, ok := err.(*echo.HTTPError)
	if !ok || httpErr.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %v", err)
	}
}

func TestParseUUIDParam_Invalid(t *testing.T) {
	c := newHandlerContext(http.MethodGet, "/")
	c.SetPath("/api/v1/hackathons/:hackathonId")
	c.SetParamNames("hackathonId")
	c.SetParamValues("bad-id")

	_, err := parseUUIDParam(c, "hackathonId")
	httpErr, ok := err.(*echo.HTTPError)
	if !ok || httpErr.Code != http.StatusBadRequest {
		t.Fatalf("expected bad request, got %v", err)
	}
}

func TestParseQueryUUID(t *testing.T) {
	validID := "5f3de833-5f0b-4675-b0b6-7e9ef32ed500"
	c := newHandlerContext(http.MethodGet, "/?teamId="+validID)

	got, err := parseQueryUUID(c, "teamId")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if got != validID {
		t.Fatalf("expected %q, got %q", validID, got)
	}
}

func TestParseLimitOffset(t *testing.T) {
	c := newHandlerContext(http.MethodGet, "/")
	limit, offset, err := parseLimitOffset(c)
	if err != nil {
		t.Fatalf("expected defaults, got error %v", err)
	}
	if limit != defaultLimit || offset != 0 {
		t.Fatalf("unexpected defaults: limit=%d offset=%d", limit, offset)
	}

	c = newHandlerContext(http.MethodGet, "/?limit=999&offset=10")
	limit, offset, err = parseLimitOffset(c)
	if err != nil {
		t.Fatalf("expected parsed values, got error %v", err)
	}
	if limit != maxLimit || offset != 10 {
		t.Fatalf("unexpected values: limit=%d offset=%d", limit, offset)
	}
}

func TestParseLimitOffset_Invalid(t *testing.T) {
	c := newHandlerContext(http.MethodGet, "/?limit=0")
	_, _, err := parseLimitOffset(c)
	if err == nil {
		t.Fatal("expected error for invalid limit")
	}

	c = newHandlerContext(http.MethodGet, "/?offset=-1")
	_, _, err = parseLimitOffset(c)
	if err == nil {
		t.Fatal("expected error for invalid offset")
	}
}

func TestHandleServiceError(t *testing.T) {
	cases := []struct {
		name     string
		err      error
		wantCode int
	}{
		{"not found", fmt.Errorf("wrap: %w", services.ErrNotFound), http.StatusNotFound},
		{"conflict", fmt.Errorf("wrap: %w", services.ErrConflict), http.StatusConflict},
		{"invalid", fmt.Errorf("wrap: %w", services.ErrInvalid), http.StatusBadRequest},
		{"forbidden", fmt.Errorf("wrap: %w", services.ErrForbidden), http.StatusForbidden},
		{"internal", errors.New("boom"), http.StatusInternalServerError},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			httpErr := handleServiceError(tc.err)
			if httpErr.Code != tc.wantCode {
				t.Fatalf("expected status %d, got %d", tc.wantCode, httpErr.Code)
			}
		})
	}
}

func TestHasAnyRole(t *testing.T) {
	if !hasAnyRole([]string{"user", "hackathon_admin"}, "hackathon_admin") {
		t.Fatal("expected role match")
	}
	if hasAnyRole([]string{"user"}, "platform_admin") {
		t.Fatal("did not expect role match")
	}
}

func TestIsAdminOrOrganizer(t *testing.T) {
	c := newHandlerContext(http.MethodGet, "/")
	c.Set("roles", []string{"user"})
	if isAdminOrOrganizer(c) {
		t.Fatal("user role should not be admin/organizer")
	}

	c.Set("roles", []string{"hackathon_organizer"})
	if !isAdminOrOrganizer(c) {
		t.Fatal("organizer role should be accepted")
	}
}
