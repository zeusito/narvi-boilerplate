package authz

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/labstack/echo/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequirePermission(t *testing.T) {
	ctx := context.Background()
	enforcer, err := NewEnforcer(ctx, "")
	require.NoError(t, err)

	e := echo.New()

	successHandler := func(c *echo.Context) error {
		return c.String(http.StatusOK, "ok")
	}

	t.Run("allowed: principal has required permission in active organization", func(t *testing.T) {
		guard := RequirePermission(enforcer, PermOrgRead)
		handler := guard(successHandler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		SetClaimsInContext(c, &PrincipalClaims{
			IsAuthenticated:      true,
			IdentityID:           "usr_01",
			ActiveOrganizationID: "org_alpha",
			OrganizationRole:     RoleAdmin,
			Permissions:          ExpandRolePermissions(RoleAdmin),
		})

		err := handler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("denied: principal lacks required permission", func(t *testing.T) {
		guard := RequirePermission(enforcer, PermOrgDelete)
		handler := guard(successHandler)

		req := httptest.NewRequest(http.MethodDelete, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Admin lacks org:delete (only Owner has it)
		SetClaimsInContext(c, &PrincipalClaims{
			IsAuthenticated:      true,
			IdentityID:           "usr_02",
			ActiveOrganizationID: "org_alpha",
			OrganizationRole:     RoleAdmin,
			Permissions:          ExpandRolePermissions(RoleAdmin),
		})

		err := handler(c)
		require.Error(t, err)

		httpErr, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusForbidden, httpErr.Code)
	})

	t.Run("denied: cross-tenant access with path parameter extractor", func(t *testing.T) {
		guard := RequirePermission(enforcer, PermOrgRead, WithOrgParam("orgId"))
		handler := guard(successHandler)

		req := httptest.NewRequest(http.MethodGet, "/organizations/org_beta", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/organizations/:orgId")
		c.SetPathValues(echo.PathValues{
			{Name: "orgId", Value: "org_beta"},
		})

		// Principal belongs to org_alpha, attempting to access org_beta
		SetClaimsInContext(c, &PrincipalClaims{
			IsAuthenticated:      true,
			IdentityID:           "usr_03",
			ActiveOrganizationID: "org_alpha",
			OrganizationRole:     RoleOwner,
			Permissions:          ExpandRolePermissions(RoleOwner),
		})

		err := handler(c)
		require.Error(t, err)

		httpErr, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusForbidden, httpErr.Code)
	})

	t.Run("allowed: matching tenant with path parameter extractor", func(t *testing.T) {
		guard := RequirePermission(enforcer, PermOrgUpdate, WithOrgParam("orgId"))
		handler := guard(successHandler)

		req := httptest.NewRequest(http.MethodPatch, "/organizations/org_alpha", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/organizations/:orgId")
		c.SetPathValues(echo.PathValues{
			{Name: "orgId", Value: "org_alpha"},
		})

		SetClaimsInContext(c, &PrincipalClaims{
			IsAuthenticated:      true,
			IdentityID:           "usr_04",
			ActiveOrganizationID: "org_alpha",
			OrganizationRole:     RoleAdmin,
			Permissions:          ExpandRolePermissions(RoleAdmin),
		})

		err := handler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("denied: authenticated but missing active organization context", func(t *testing.T) {
		guard := RequirePermission(enforcer, PermOrgRead)
		handler := guard(successHandler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		SetClaimsInContext(c, &PrincipalClaims{
			IsAuthenticated:      true,
			IdentityID:           "usr_05",
			ActiveOrganizationID: "",
			OrganizationRole:     "",
			Permissions:          []string{},
		})

		err := handler(c)
		require.Error(t, err)

		httpErr, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusForbidden, httpErr.Code)
	})

	t.Run("denied: unauthenticated caller", func(t *testing.T) {
		guard := RequirePermission(enforcer, PermOrgRead)
		handler := guard(successHandler)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// No claims set in context
		err := handler(c)
		require.Error(t, err)

		httpErr, ok := err.(*echo.HTTPError)
		require.True(t, ok)
		assert.Equal(t, http.StatusUnauthorized, httpErr.Code)
	})

	t.Run("allowed: automatically expands permissions when not pre-populated", func(t *testing.T) {
		guard := RequirePermission(enforcer, PermMembersInvite)
		handler := guard(successHandler)

		req := httptest.NewRequest(http.MethodPost, "/test", nil)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		// Permissions slice is nil/empty, but OrganizationRole is admin
		SetClaimsInContext(c, &PrincipalClaims{
			IsAuthenticated:      true,
			IdentityID:           "usr_06",
			ActiveOrganizationID: "org_alpha",
			OrganizationRole:     RoleAdmin,
			Permissions:          nil,
		})

		err := handler(c)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, rec.Code)
	})
}
