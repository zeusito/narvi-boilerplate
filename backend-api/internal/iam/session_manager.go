package iam

import (
	"github.com/labstack/echo/v5"
)

type SessionManager interface {
	Introspect(ctx *echo.Context, token string) *PrincipalClaims
}

type DefaultSessionManager struct {
	authUseCase authUseCases
}

func newSessionManager(useCases authUseCases) SessionManager {
	return &DefaultSessionManager{authUseCase: useCases}
}

func (m *DefaultSessionManager) Introspect(ctx *echo.Context, token string) *PrincipalClaims {
	return m.authUseCase.Introspect(ctx, token)
}
