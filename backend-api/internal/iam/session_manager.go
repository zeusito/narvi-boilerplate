package iam

import (
	"github.com/labstack/echo/v5"
)

type SessionManager interface {
	Introspect(ctx *echo.Context, token string) (*PrincipalClaims, error)
}

type defaultSessionManager struct {
	useCases authUseCases
}

func newSessionManager(useCases authUseCases) SessionManager {
	return &defaultSessionManager{useCases: useCases}
}

func (m *defaultSessionManager) Introspect(ctx *echo.Context, token string) (*PrincipalClaims, error) {
	return m.useCases.Introspect(ctx, token)
}
