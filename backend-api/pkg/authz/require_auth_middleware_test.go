package authz

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubIntrospector struct {
	claims *PrincipalClaims
}

func (s *stubIntrospector) Introspect(_ context.Context, _ string) *PrincipalClaims {
	return s.claims
}

func TestRequireAuthMiddleware_ValidTokenInjectsClaims(t *testing.T) {
	introspector := &stubIntrospector{
		claims: &PrincipalClaims{
			IsAuthenticated:      true,
			SessionID:            "ses_123",
			IdentityID:           "id_123",
			Email:                "admin@example.com",
			ActiveOrganizationID: "org_123",
		},
	}

	var seenClaims PrincipalClaims
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenClaims = ExtractClaimsFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/auth/introspect", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()

	RequireAuthMiddleware(introspector)(next).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, seenClaims.IsAuthenticated)
	assert.Equal(t, "ses_123", seenClaims.SessionID)
	assert.Equal(t, "admin@example.com", seenClaims.Email)
	assert.Equal(t, "org_123", seenClaims.ActiveOrganizationID)
}

func TestRequireAuthMiddleware_MissingToken(t *testing.T) {
	introspector := &stubIntrospector{
		claims: &PrincipalClaims{IsAuthenticated: true},
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/auth/introspect", nil)
	rec := httptest.NewRecorder()

	RequireAuthMiddleware(introspector)(next).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestRequireAuthMiddleware_InvalidToken(t *testing.T) {
	introspector := &stubIntrospector{
		claims: &PrincipalClaims{IsAuthenticated: false},
	}

	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/auth/introspect", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()

	RequireAuthMiddleware(introspector)(next).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.False(t, nextCalled)
}

func TestRequireAuthMiddleware_MalformedHeader(t *testing.T) {
	introspector := &stubIntrospector{
		claims: &PrincipalClaims{IsAuthenticated: true},
	}

	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/v1/auth/introspect", nil)
	req.Header.Set("Authorization", "Token abc")
	rec := httptest.NewRecorder()

	RequireAuthMiddleware(introspector)(next).ServeHTTP(rec, req)

	require.Equal(t, http.StatusUnauthorized, rec.Code)
}
