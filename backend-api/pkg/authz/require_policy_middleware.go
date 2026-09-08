package authz

import "net/http"

func RequirePolicyMiddleware(enforcer Enforcer, action string, resourceType string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims := ExtractClaimsFromContext(r.Context())
			if !claims.IsAuthenticated {
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			allowed, err := enforcer.IsAllowed(r.Context(), EvalInput{
				Principal: PrincipalInput{
					IsAuthenticated:        claims.IsAuthenticated,
					ActiveOrganizationID:   claims.ActiveOrganizationID,
					ActiveOrganizationKind: claims.OrganizationKind,
					Role:                   claims.OrganizationRole,
				},
				Action: action,
				Resource: ResourceInput{
					Type: resourceType,
				},
			})
			if err != nil || !allowed {
				http.Error(w, "Forbidden", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
