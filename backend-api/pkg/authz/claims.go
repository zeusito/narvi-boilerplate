package authz

import "github.com/labstack/echo/v5"

const PrincipalClaimsKey = "claims"

// PrincipalClaims encapsulates the authenticated identity, tenant context, and resolved permissions.
type PrincipalClaims struct {
	IsAuthenticated      bool     `json:"isAuthenticated"`
	SessionID            string   `json:"sessionId"`
	IdentityID           string   `json:"identityId"`
	Email                string   `json:"email"`
	FullName             string   `json:"fullName"`
	ActiveOrganizationID string   `json:"activeOrganizationId,omitempty"`
	OrganizationName     string   `json:"organizationName,omitempty"`
	OrganizationSlug     string   `json:"organizationSlug,omitempty"`
	OrganizationLogo     string   `json:"organizationLogo,omitempty"`
	OrganizationRole     string   `json:"organizationRole,omitempty"`
	Permissions          []string `json:"permissions"`
}

// ExtractClaimsFromContext retrieves the PrincipalClaims stored in the Echo context.
func ExtractClaimsFromContext(ctx *echo.Context) *PrincipalClaims {
	claims, err := echo.ContextGetOr[*PrincipalClaims](ctx, PrincipalClaimsKey, &PrincipalClaims{IsAuthenticated: false, Permissions: []string{}})
	if err != nil {
		return &PrincipalClaims{IsAuthenticated: false, Permissions: []string{}}
	}
	return claims
}

// SetClaimsInContext stores the PrincipalClaims in the Echo context.
func SetClaimsInContext(ctx *echo.Context, claims *PrincipalClaims) {
	ctx.Set(PrincipalClaimsKey, claims)
}
