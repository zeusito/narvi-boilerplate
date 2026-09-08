package authz

import (
	"context"
	"net/http"
	"strings"

	"github.com/rs/zerolog/log"
)

// SessionIntrospector defines the interface for session introspection, Your IAM module should implement this interface to provide session introspection functionality.
type SessionIntrospector interface {
	Introspect(ctx context.Context, token string) *PrincipalClaims
}

// RequireAuthMiddleware is a middleware that checks if the request has a valid token
func RequireAuthMiddleware(sessionIntrospector SessionIntrospector) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := r.Header.Get("Authorization")

			if token == "" || !strings.HasPrefix(token, "Bearer ") {
				log.Ctx(r.Context()).Warn().Msg("no token provided")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Remove the "Bearer " prefix
			token = token[7:]

			// Introspect token
			claims := sessionIntrospector.Introspect(r.Context(), token)
			if !claims.IsAuthenticated {
				log.Ctx(r.Context()).Warn().Msg("session not authenticated")
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}

			// Add claims to context
			ctx := AddClaimsToContext(r.Context(), claims)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
