package middlewares

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"github.com/labstack/echo/v4"
)

func TestAuthMiddleware_Disabled(t *testing.T) {
	mw, err := AuthMiddleware(JWTConfig{Required: false}, nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if mw != nil {
		t.Fatal("expected nil middleware when auth is disabled")
	}
}

func TestAuthMiddleware_RequiredMissingJWKSURL(t *testing.T) {
	_, err := AuthMiddleware(JWTConfig{Required: true}, nil)
	if err == nil {
		t.Fatal("expected error when required auth has empty JWKS URL")
	}
}

func TestExtractRoles_RealmOnly(t *testing.T) {
	claims := jwt.MapClaims{
		"realm_access": map[string]any{
			"roles": []any{"user", "hackathon_admin"},
		},
	}
	roles := extractRoles(claims, "")
	if len(roles) != 2 {
		t.Fatalf("expected 2 roles, got %v", roles)
	}
}

func TestExtractRoles_WithClientAndDedupe(t *testing.T) {
	claims := jwt.MapClaims{
		"realm_access": map[string]any{
			"roles": []any{"user", "hackathon_admin"},
		},
		"resource_access": map[string]any{
			"hackathon-service-api-client": map[string]any{
				"roles": []any{"hackathon_admin", "rule_editor"},
			},
		},
	}
	roles := extractRoles(claims, "hackathon-service-api-client")
	seen := map[string]bool{}
	for _, r := range roles {
		if seen[r] {
			t.Fatalf("role %q was duplicated in %v", r, roles)
		}
		seen[r] = true
	}
	if !seen["rule_editor"] {
		t.Fatalf("expected client role rule_editor in %v", roles)
	}
}

func TestHasRole(t *testing.T) {
	roles := []string{"user", "hackathon_admin"}
	if !hasRole(roles, "hackathon_admin") {
		t.Fatal("expected hasRole to return true")
	}
	if hasRole(roles, "platform_admin") {
		t.Fatal("expected hasRole to return false for missing role")
	}
}

func TestClaimString(t *testing.T) {
	claims := jwt.MapClaims{
		"sub": "user_001",
		"num": 42,
	}
	if got := claimString(claims, "sub"); got != "user_001" {
		t.Fatalf("expected user_001, got %q", got)
	}
	if got := claimString(claims, "num"); got != "" {
		t.Fatalf("expected empty for non-string claim, got %q", got)
	}
	if got := claimString(claims, "missing"); got != "" {
		t.Fatalf("expected empty for missing claim, got %q", got)
	}
}

func TestRolesFromContext(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	if got := RolesFromContext(c); got != nil {
		t.Fatalf("expected nil roles when unset, got %v", got)
	}

	c.Set("roles", "not-a-slice")
	if got := RolesFromContext(c); got != nil {
		t.Fatalf("expected nil roles for invalid type, got %v", got)
	}

	c.Set("roles", []string{"user"})
	got := RolesFromContext(c)
	if len(got) != 1 || got[0] != "user" {
		t.Fatalf("unexpected roles %v", got)
	}
}

func TestRequireAnyRole(t *testing.T) {
	mw := RequireAnyRole("hackathon_admin", "platform_admin")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("roles", []string{"hackathon_admin"})

	called := false
	err := mw(func(ec echo.Context) error {
		called = true
		return ec.NoContent(http.StatusNoContent)
	})(c)
	if err != nil {
		t.Fatalf("expected middleware success, got %v", err)
	}
	if !called {
		t.Fatal("expected next handler to be called")
	}
}

func TestRequireAnyRole_Forbidden(t *testing.T) {
	mw := RequireAnyRole("hackathon_admin")

	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("roles", []string{"user"})

	err := mw(func(ec echo.Context) error { return ec.NoContent(http.StatusNoContent) })(c)
	httpErr, ok := err.(*echo.HTTPError)
	if !ok {
		t.Fatalf("expected *echo.HTTPError, got %T", err)
	}
	if httpErr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", httpErr.Code)
	}
}

func generateTestKey(t *testing.T) *rsa.PrivateKey {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("rsa.GenerateKey: %v", err)
	}
	return key
}

func testJWKSBody(key *rsa.PrivateKey, kid string) []byte {
	n := base64.RawURLEncoding.EncodeToString(key.PublicKey.N.Bytes())
	exp := make([]byte, 4)
	binary.BigEndian.PutUint32(exp, uint32(key.PublicKey.E))
	for len(exp) > 1 && exp[0] == 0 {
		exp = exp[1:]
	}
	body, _ := json.Marshal(map[string]any{
		"keys": []map[string]any{{
			"kty": "RSA",
			"use": "sig",
			"kid": kid,
			"alg": "RS256",
			"n":   n,
			"e":   base64.RawURLEncoding.EncodeToString(exp),
		}},
	})
	return body
}

func mintJWT(t *testing.T, key *rsa.PrivateKey, kid string, claims jwt.MapClaims) string {
	t.Helper()
	tok := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
	tok.Header["kid"] = kid
	signed, err := tok.SignedString(key)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}

func TestAuthMiddleware_Integration(t *testing.T) {
	key := generateTestKey(t)
	kid := "test-key-1"
	issuer := "https://auth.example.com"
	audience := "hackathon-api"
	clientID := "hackathon-service-api-client"
	now := time.Now()

	jwksBody := testJWKSBody(key, kid)
	jwksSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(jwksBody)
	}))
	defer jwksSrv.Close()

	cfg := JWTConfig{
		JWKSURL:  jwksSrv.URL,
		Issuer:   issuer,
		Audience: audience,
		ClientID: clientID,
		Required: true,
	}
	mw, err := AuthMiddleware(cfg, nil)
	if err != nil {
		t.Fatalf("AuthMiddleware: %v", err)
	}

	validClaims := jwt.MapClaims{
		"sub":   "user-123",
		"email": "user-123@example.com",
		"iss":   issuer,
		"aud":   audience,
		"iat":   float64(now.Unix()),
		"exp":   float64(now.Add(1 * time.Hour).Unix()),
		"realm_access": map[string]any{
			"roles": []any{"user"},
		},
		"resource_access": map[string]any{
			clientID: map[string]any{
				"roles": []any{"hackathon_admin"},
			},
		},
	}
	bannedClaims := jwt.MapClaims{
		"sub": "user-999",
		"iss": issuer,
		"aud": audience,
		"iat": float64(now.Unix()),
		"exp": float64(now.Add(1 * time.Hour).Unix()),
		"realm_access": map[string]any{
			"roles": []any{"banned_user"},
		},
	}
	wrongIssuerClaims := jwt.MapClaims{
		"sub": "user-123",
		"iss": "https://other.example.com",
		"aud": audience,
		"iat": float64(now.Unix()),
		"exp": float64(now.Add(1 * time.Hour).Unix()),
	}
	wrongAudienceClaims := jwt.MapClaims{
		"sub": "user-123",
		"iss": issuer,
		"aud": "other-audience",
		"iat": float64(now.Unix()),
		"exp": float64(now.Add(1 * time.Hour).Unix()),
	}

	tests := []struct {
		name          string
		authHeader    string
		wantCode      int
		wantCalled    bool
		wantUserID    string
		wantRole      string
		expectedError int
	}{
		{
			name:          "missing authorization header",
			authHeader:    "",
			expectedError: http.StatusUnauthorized,
		},
		{
			name:          "malformed authorization header",
			authHeader:    "Token abc",
			expectedError: http.StatusUnauthorized,
		},
		{
			name:          "garbage token",
			authHeader:    "Bearer not.a.jwt",
			expectedError: http.StatusUnauthorized,
		},
		{
			name:          "wrong issuer",
			authHeader:    "Bearer " + mintJWT(t, key, kid, wrongIssuerClaims),
			expectedError: http.StatusUnauthorized,
		},
		{
			name:          "wrong audience",
			authHeader:    "Bearer " + mintJWT(t, key, kid, wrongAudienceClaims),
			expectedError: http.StatusUnauthorized,
		},
		{
			name:          "banned user",
			authHeader:    "Bearer " + mintJWT(t, key, kid, bannedClaims),
			expectedError: http.StatusForbidden,
		},
		{
			name:       "valid token",
			authHeader: "Bearer " + mintJWT(t, key, kid, validClaims),
			wantCode:   http.StatusNoContent,
			wantCalled: true,
			wantUserID: "user-123",
			wantRole:   "hackathon_admin",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			called := false
			err := mw(func(ec echo.Context) error {
				called = true
				if tc.wantUserID != "" {
					if got, _ := ec.Get("user_id").(string); got != tc.wantUserID {
						t.Fatalf("user_id mismatch: got=%q want=%q", got, tc.wantUserID)
					}
				}
				if tc.wantRole != "" {
					if !hasRole(RolesFromContext(ec), tc.wantRole) {
						t.Fatalf("expected role %q in %v", tc.wantRole, RolesFromContext(ec))
					}
				}
				return ec.NoContent(tc.wantCode)
			})(c)

			if tc.expectedError > 0 {
				httpErr, ok := err.(*echo.HTTPError)
				if !ok {
					t.Fatalf("expected *echo.HTTPError, got %T (%v)", err, err)
				}
				if httpErr.Code != tc.expectedError {
					t.Fatalf("expected status %d, got %d", tc.expectedError, httpErr.Code)
				}
				if called {
					t.Fatal("next handler should not be called on auth failure")
				}
				return
			}

			if err != nil {
				t.Fatalf("expected success, got %v", err)
			}
			if rec.Code != tc.wantCode {
				t.Fatalf("expected response code %d, got %d", tc.wantCode, rec.Code)
			}
			if called != tc.wantCalled {
				t.Fatalf("handler called mismatch: got=%v want=%v", called, tc.wantCalled)
			}
		})
	}
}
