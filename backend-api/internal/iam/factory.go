package iam

import (
	"backend-api/pkg/mailer"
	"backend-api/pkg/toolbox/hasher"

	"github.com/go-chi/chi/v5"
	"github.com/uptrace/bun"
)

type Module struct {
	SessionIntrospector SessionIntrospectionService
}

func NewModule(
	mux *chi.Mux,
	db *bun.DB,
	mail mailer.Mailer,
	hmacHasher hasher.Hasher,
) *Module {
	identityRepo := newIdentityRepository(db)
	orgRepo := newOrganizationRepository(db)
	verificationRepo := newVerificationRepository(db)
	sessionRepo := newSessionRepository(db)

	authService := newAuthnService(orgRepo, identityRepo, verificationRepo, sessionRepo, mail, hmacHasher)
	orgService := newOrgService(orgRepo)
	sessionIntrospectionService := newSessionIntrospectionService(sessionRepo, hmacHasher)
	sessionService := newDefaultSessionService(sessionRepo)
	_ = newAuthnController(mux, sessionIntrospectionService, sessionService, authService)
	_ = newOrgController(mux, sessionIntrospectionService, orgService)

	return &Module{
		SessionIntrospector: sessionIntrospectionService,
	}
}
