package authz

import (
	"context"
)

type ctxKeyAuthClaims int

const (
	PrincipalClaimsKey ctxKeyAuthClaims = 1
)

// PrincipalClaims encapsulates the authenticated identity, tenant context, and resolved permissions.
type PrincipalClaims struct {
	IsAuthenticated      bool   `json:"isAuthenticated"`
	SessionID            string `json:"sessionId"`
	IdentityID           string `json:"identityId"`
	Email                string `json:"email"`
	FullName             string `json:"fullName"`
	ActiveOrganizationID string `json:"activeOrganizationId,omitempty"`
	OrganizationName     string `json:"organizationName,omitempty"`
	OrganizationSlug     string `json:"organizationSlug,omitempty"`
	OrganizationLogo     string `json:"organizationLogo,omitempty"`
	OrganizationRole     string `json:"organizationRole,omitempty"`
}

// ExtractClaimsFromContext retrieves the PrincipalClaims stored in the context.
func ExtractClaimsFromContext(ctx context.Context) PrincipalClaims {
	claims, ok := ctx.Value(PrincipalClaimsKey).(PrincipalClaims)
	if !ok {
		return PrincipalClaims{IsAuthenticated: false}
	}

	return claims
}

// AddClaimsToContext stores the PrincipalClaims in the provided context.
func AddClaimsToContext(ctx context.Context, claims *PrincipalClaims) context.Context {
	return context.WithValue(ctx, PrincipalClaimsKey, claims)
}
