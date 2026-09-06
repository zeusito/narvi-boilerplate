package authz

import (
	"backend-api/pkg/terrors"

	"github.com/labstack/echo/v5"
	"github.com/rs/zerolog/log"
)

// ResourceOrgExtractor extracts the target organization ID for a given request.
type ResourceOrgExtractor func(c *echo.Context, claims *PrincipalClaims) string

// GuardOptions holds configuration options for the authorization guard.
type GuardOptions struct {
	ResourceOrgExtractor ResourceOrgExtractor
	ResourceType         string
}

// Option configures GuardOptions.
type Option func(*GuardOptions)

// WithOrgParam configures the guard to extract the target organization ID from a path parameter.
func WithOrgParam(paramName string) Option {
	return func(o *GuardOptions) {
		o.ResourceOrgExtractor = func(c *echo.Context, claims *PrincipalClaims) string {
			return c.Param(paramName)
		}
	}
}

// WithCustomOrgExtractor configures a custom organization ID extractor function.
func WithCustomOrgExtractor(extractor ResourceOrgExtractor) Option {
	return func(o *GuardOptions) {
		o.ResourceOrgExtractor = extractor
	}
}

// WithResourceType configures the resource type passed to OPA (defaults to "organization").
func WithResourceType(resType string) Option {
	return func(o *GuardOptions) {
		o.ResourceType = resType
	}
}

// RequirePermission creates an Echo middleware that enforces tenant isolation and role permissions via OPA.
// It requires an authenticated principal in the Echo context (e.g. populated by RequireAuth).
func RequirePermission(enforcer Enforcer, action string, opts ...Option) echo.MiddlewareFunc {
	config := GuardOptions{
		ResourceType: "organization",
		ResourceOrgExtractor: func(c *echo.Context, claims *PrincipalClaims) string {
			return claims.ActiveOrganizationID
		},
	}

	for _, opt := range opts {
		opt(&config)
	}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			claims := ExtractClaimsFromContext(c)
			if !claims.IsAuthenticated {
				log.Warn().Msg("unauthenticated principal attempted to access permission-guarded endpoint")
				return terrors.UnAuthorized("authentication required").ToEchoHttpError()
			}

			if claims.ActiveOrganizationID == "" {
				log.Warn().Str("identity_id", claims.IdentityID).Msg("principal has no active organization context")
				return terrors.Forbidden("active organization context required").ToEchoHttpError()
			}

			targetOrgID := config.ResourceOrgExtractor(c, claims)
			if targetOrgID == "" {
				log.Warn().Msg("unable to resolve target resource organization ID")
				return terrors.Forbidden("insufficient permissions").ToEchoHttpError()
			}

			// Ensure permissions are populated (expand from OrganizationRole if needed)
			permissions := claims.Permissions
			if len(permissions) == 0 && claims.OrganizationRole != "" {
				permissions = ExpandRolePermissions(claims.OrganizationRole)
			}

			input := EvalInput{
				Principal: PrincipalInput{
					IdentityID:           claims.IdentityID,
					Email:                claims.Email,
					ActiveOrganizationID: claims.ActiveOrganizationID,
					Role:                 claims.OrganizationRole,
					Permissions:          permissions,
				},
				Action: action,
				Resource: ResourceInput{
					Type:           config.ResourceType,
					OrganizationID: targetOrgID,
				},
			}

			allowed, err := enforcer.IsAllowed(c.Request().Context(), input)
			if err != nil {
				log.Error().Err(err).Str("action", action).Msg("opa evaluation error")
				return terrors.Forbidden("insufficient permissions").ToEchoHttpError()
			}

			if !allowed {
				log.Warn().
					Str("identity_id", claims.IdentityID).
					Str("role", claims.OrganizationRole).
					Str("action", action).
					Str("active_org", claims.ActiveOrganizationID).
					Str("target_org", targetOrgID).
					Msg("authorization denied by OPA policy")
				return terrors.Forbidden("insufficient permissions").ToEchoHttpError()
			}

			return next(c)
		}
	}
}
