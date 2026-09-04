package iam

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v5"
	"github.com/rs/zerolog/log"
)

func RequireAuth(sessionManager SessionManager) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			token := c.Request().Header.Get("Authorization")

			if token == "" || !strings.HasPrefix(token, "Bearer ") {
				log.Warn().Msg("no token provided")
				return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
			}

			// Remove the "Bearer " prefix
			token = token[7:]

			// Introspect token
			claims := sessionManager.Introspect(c, token)
			if !claims.IsAuthenticated {
				log.Warn().Msg("session not authenticated")
				return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
			}

			// Set claims in context
			SetClaimsInContext(c, claims)

			return next(c)
		}
	}
}
