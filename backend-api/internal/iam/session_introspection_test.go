package iam

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"backend-api/pkg/authz"
	"backend-api/pkg/toolbox"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testAdminIdentityID = "01a02086-04a2-75a7-ba24-1616b586c403"
	testAdminEmail      = "admin@example.com"
	testAdminOrgID      = "01a02086-04a2-75a7-ba24-12d5872b8c49"
)

func setupIntrospectionService() *DefaultSessionIntrospectionService {
	sessionRepo := newSessionRepository(testDB)
	return &DefaultSessionIntrospectionService{
		sessionRepo: sessionRepo,
		hmacHasher:  testHasher,
	}
}

func issueTestSessionToken(t *testing.T) string {
	t.Helper()

	authSvc, spy := setupAuthService(t)

	_, err := testDB.NewDelete().Table("verifications").Where("identity_id = ?", testAdminIdentityID).Exec(t.Context())
	require.NoError(t, err)
	_, err = testDB.NewDelete().Table("identity_sessions").Where("identity_id = ?", testAdminIdentityID).Exec(t.Context())
	require.NoError(t, err)

	err = authSvc.SendOTP(t.Context(), &SendOTPRequest{Email: testAdminEmail})
	require.NoError(t, err)

	code := spy.LastCode()
	require.NotEmpty(t, code)

	resp, err := authSvc.VerifyOTP(t.Context(), &VerifyOTPRequest{
		Email:     testAdminEmail,
		Code:      code,
		IPAddress: "127.0.0.1",
		UserAgent: "test-agent",
	})
	require.NoError(t, err)
	require.NotNil(t, resp)
	require.NotEmpty(t, resp.Token)

	return resp.Token
}

func TestIntrospect_ValidToken(t *testing.T) {
	token := issueTestSessionToken(t)
	introspectionSvc := setupIntrospectionService()

	claims := introspectionSvc.Introspect(t.Context(), token)

	require.NotNil(t, claims)
	assert.True(t, claims.IsAuthenticated)
	assert.Equal(t, testAdminIdentityID, claims.IdentityID)
	assert.Equal(t, testAdminEmail, claims.Email)
	assert.Equal(t, "Super Admin", claims.FullName)
	assert.Equal(t, testAdminOrgID, claims.ActiveOrganizationID)
	assert.Equal(t, "Acme Corp", claims.OrganizationName)
	assert.Equal(t, "acme", claims.OrganizationSlug)
	assert.Equal(t, "management", claims.OrganizationKind)
	assert.Equal(t, "owner", claims.OrganizationRole)
	assert.NotEmpty(t, claims.SessionID)
}

func TestIntrospect_InvalidToken(t *testing.T) {
	introspectionSvc := setupIntrospectionService()

	claims := introspectionSvc.Introspect(t.Context(), "ses_invalidtoken123")

	require.NotNil(t, claims)
	assert.False(t, claims.IsAuthenticated)
}

func TestIntrospect_ExpiredSession(t *testing.T) {
	token, hashedToken, err := toolbox.GenerateOpaqueToken(testHasher, "ses")
	require.NoError(t, err)

	expired := &IdentitySession{
		ID:             hashedToken,
		IdentityID:     testAdminIdentityID,
		OrganizationID: testAdminOrgID,
		IPAddress:      "127.0.0.1",
		UserAgent:      "test-agent",
		ExpiresAt:      time.Now().UTC().Add(-1 * time.Hour),
		CreatedAt:      time.Now().UTC().Add(-2 * time.Hour),
	}
	_, err = testDB.NewInsert().Model(expired).Exec(t.Context())
	require.NoError(t, err)

	t.Cleanup(func() {
		_, _ = testDB.NewDelete().Table("identity_sessions").Where("id = ?", hashedToken).Exec(context.Background())
	})

	introspectionSvc := setupIntrospectionService()
	claims := introspectionSvc.Introspect(t.Context(), token)

	require.NotNil(t, claims)
	assert.False(t, claims.IsAuthenticated)
}

func TestIntrospectEndpoint_ValidToken(t *testing.T) {
	token := issueTestSessionToken(t)
	introspectionSvc := setupIntrospectionService()

	mux := chi.NewMux()
	_ = newAuthnController(mux, introspectionSvc, newDefaultSessionService(newSessionRepository(testDB)), mustAuthService(t))

	req := httptest.NewRequest(http.MethodGet, "/v1/auth/introspect", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	var claims authz.PrincipalClaims
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &claims))
	assert.True(t, claims.IsAuthenticated)
	assert.Equal(t, testAdminIdentityID, claims.IdentityID)
	assert.Equal(t, testAdminEmail, claims.Email)
	assert.Equal(t, testAdminOrgID, claims.ActiveOrganizationID)
}

func TestIntrospectEndpoint_MissingToken(t *testing.T) {
	introspectionSvc := setupIntrospectionService()

	mux := chi.NewMux()
	_ = newAuthnController(mux, introspectionSvc, newDefaultSessionService(newSessionRepository(testDB)), mustAuthService(t))

	req := httptest.NewRequest(http.MethodGet, "/v1/auth/introspect", nil)
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestIntrospectEndpoint_InvalidToken(t *testing.T) {
	introspectionSvc := setupIntrospectionService()

	mux := chi.NewMux()
	_ = newAuthnController(mux, introspectionSvc, newDefaultSessionService(newSessionRepository(testDB)), mustAuthService(t))

	req := httptest.NewRequest(http.MethodGet, "/v1/auth/introspect", nil)
	req.Header.Set("Authorization", "Bearer ses_bogus")
	rec := httptest.NewRecorder()

	mux.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func mustAuthService(t *testing.T) authService {
	t.Helper()
	svc, _ := setupAuthService(t)
	return svc
}
