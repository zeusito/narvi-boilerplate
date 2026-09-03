package iam

import (
	"context"
)

type SessionManager interface {
	Introspect(ctx context.Context, token string) (*PrincipalClaims, error)
}

type defaultSessionManager struct {
	useCases authUseCases
}

func newSessionManager(useCases authUseCases) SessionManager {
	return &defaultSessionManager{useCases: useCases}
}

func (m *defaultSessionManager) Introspect(ctx context.Context, token string) (*PrincipalClaims, error) {
	return m.useCases.Introspect(ctx, token)
}
