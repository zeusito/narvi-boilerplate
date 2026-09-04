package iam

import (
	"backend-api/pkg/mailer"
	"backend-api/pkg/toolbox/hasher"

	"github.com/labstack/echo/v5"
	"github.com/uptrace/bun"
)

type Module struct {
	SessionManager SessionManager
}

func NewModule(
	mux *echo.Echo,
	db *bun.DB,
	mail mailer.Mailer,
	hmacHasher hasher.Hasher,
) *Module {
	identityRepo := newIdentityRepository(db)
	orgRepo := newOrganizationRepository(db)
	verificationRepo := newVerificationRepository(db)
	sessionRepo := newSessionRepository(db)

	useCases := newAuthnUseCases(orgRepo, identityRepo, verificationRepo, sessionRepo, mail, hmacHasher)
	sessionManager := newSessionManager(useCases)
	_ = newAuthnController(mux, sessionManager, useCases)

	return &Module{
		SessionManager: sessionManager,
	}
}
