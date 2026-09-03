package iam

import (
	"backend-api/pkg/mailer"
	"backend-api/pkg/toolbox/hasher"

	"github.com/labstack/echo/v5"
	"github.com/uptrace/bun"
)

type Module struct {
	sessionManager SessionManager
	middleware     *authnMiddleware
	controller     *authnController
}

func NewModule(
	mux *echo.Echo,
	db *bun.DB,
	mail mailer.Mailer,
	hmacHasher hasher.Hasher,
) *Module {
	identityRepo := newIdentityRepository(db)
	verificationRepo := newVerificationRepository(db)
	sessionRepo := newSessionRepository(db)

	useCases := newAuthnUseCases(identityRepo, verificationRepo, sessionRepo, mail, hmacHasher)
	sessionManager := newSessionManager(useCases)
	middleware := newAuthnMiddleware(sessionManager)
	controller := newAuthnController(mux, useCases, middleware)

	return &Module{
		sessionManager: sessionManager,
		middleware:     middleware,
		controller:     controller,
	}
}

func (m *Module) SessionManager() SessionManager {
	return m.sessionManager
}

func (m *Module) RequireAuth() echo.MiddlewareFunc {
	return m.middleware.RequireAuth
}
