package iam

import (
	"backend-api/pkg/terrors"
	"strings"

	"github.com/labstack/echo/v5"
)

const PrincipalContextKey = "principal"

type authnMiddleware struct {
	sessionManager SessionManager
}

func newAuthnMiddleware(sessionManager SessionManager) *authnMiddleware {
	return &authnMiddleware{sessionManager: sessionManager}
}

func (m *authnMiddleware) RequireAuth(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c *echo.Context) error {
		authHeader := c.Request().Header.Get("Authorization")
		if authHeader == "" {
			return terrors.UnAuthorized("missing authorization header").ToEchoHttpError()
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			return terrors.UnAuthorized("invalid authorization header format, expected 'Bearer <token>'").ToEchoHttpError()
		}

		token := strings.TrimSpace(parts[1])
		if token == "" {
			return terrors.UnAuthorized("empty token provided").ToEchoHttpError()
		}

		claims, err := m.sessionManager.Introspect(c.Request().Context(), token)
		if err != nil {
			if tErr, ok := err.(*terrors.Terror); ok {
				return tErr.ToEchoHttpError()
			}
			return terrors.UnAuthorized("invalid or expired session").ToEchoHttpError()
		}

		c.Set(PrincipalContextKey, claims)

		return next(c)
	}
}

func GetPrincipal(c *echo.Context) *PrincipalClaims {
	val := c.Get(PrincipalContextKey)
	if val == nil {
		return nil
	}
	claims, ok := val.(*PrincipalClaims)
	if !ok {
		return nil
	}
	return claims
}
