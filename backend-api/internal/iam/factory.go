package iam

import (
	"backend-api/pkg/mailer"
	"backend-api/pkg/toolbox/hasher"

	"github.com/labstack/echo/v5"
	"github.com/uptrace/bun"
)

type Module struct {
	SessionIntrospectionService SessionIntrospectionService
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

	authService := newAuthnService(orgRepo, identityRepo, verificationRepo, sessionRepo, mail, hmacHasher)
	sessionService := newSessionIntrospectionService(sessionRepo, hmacHasher)
	_ = newAuthnController(mux, sessionService, authService)

	return &Module{
		SessionIntrospectionService: sessionService,
	}
}
