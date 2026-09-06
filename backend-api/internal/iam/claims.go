package iam

import (
	"backend-api/pkg/authz"

	"github.com/labstack/echo/v5"
)

const PrincipalClaimsKey = authz.PrincipalClaimsKey

// PrincipalClaims aliases authz.PrincipalClaims for backward compatibility within iam.
type PrincipalClaims = authz.PrincipalClaims

// ExtractClaimsFromContext extracts PrincipalClaims from Echo context.
func ExtractClaimsFromContext(ctx *echo.Context) *PrincipalClaims {
	return authz.ExtractClaimsFromContext(ctx)
}

// SetClaimsInContext sets PrincipalClaims in Echo context.
func SetClaimsInContext(ctx *echo.Context, claims *PrincipalClaims) {
	authz.SetClaimsInContext(ctx, claims)
}
