package iam

import (
	"backend-api/pkg/authz"
	"net/http"
	"strings"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/rs/zerolog/log"
)

func RequireAuth(sessionService SessionIntrospectionService) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			token := c.Request().Header.Get("Authorization")

			if token == "" || !strings.HasPrefix(token, "Bearer ") {
				log.Warn().Msg("no token provided")
				return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
			}

			// Remove the "Bearer " prefix
			token = strings.TrimSpace(token[7:])

			// Introspect token passing standard context
			claims := sessionService.Introspect(c.Request().Context(), token)
			if !claims.IsAuthenticated {
				log.Warn().Msg("session not authenticated")
				return echo.NewHTTPError(http.StatusUnauthorized, "authentication required")
			}

			// Set claims in context
			authz.SetClaimsInContext(c, claims)

			return next(c)
		}
	}
}

// NewAuthRateLimiter creates an IP-based rate limiter middleware for authentication endpoints.
func NewAuthRateLimiter(rate float64, burst int) echo.MiddlewareFunc {
	return middleware.RateLimiterWithConfig(middleware.RateLimiterConfig{
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(middleware.RateLimiterMemoryStoreConfig{
			Rate:      rate,
			Burst:     burst,
			ExpiresIn: 5 * time.Minute,
		}),
		IdentifierExtractor: func(ctx *echo.Context) (string, error) {
			return ctx.RealIP(), nil
		},
		ErrorHandler: func(ctx *echo.Context, err error) error {
			return echo.NewHTTPError(http.StatusInternalServerError, "rate limiter error")
		},
		DenyHandler: func(ctx *echo.Context, identifier string, err error) error {
			return echo.NewHTTPError(http.StatusTooManyRequests, "too many requests, please try again later")
		},
	})
}
