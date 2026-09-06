package iam

import (
	"context"

	"backend-api/pkg/authz"
	"backend-api/pkg/mailer"
	"backend-api/pkg/toolbox/hasher"

	"github.com/labstack/echo/v5"
	"github.com/rs/zerolog/log"
	"github.com/uptrace/bun"
)

type Module struct {
	SessionIntrospectionService SessionIntrospectionService
	AuthzEnforcer               authz.Enforcer
}

// RequirePermission provides convenient access to the authorization guard middleware configured with the module's OPA enforcer.
func (m *Module) RequirePermission(action string, opts ...authz.Option) echo.MiddlewareFunc {
	return authz.RequirePermission(m.AuthzEnforcer, action, opts...)
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
	sessionIntrospectionService := newSessionIntrospectionService(sessionRepo, hmacHasher)
	sessionService := newDefaultSessionService(sessionRepo)
	_ = newAuthnController(mux, sessionIntrospectionService, sessionService, authService)

	opaEnforcer, err := authz.NewEnforcer(context.Background(), "")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize OPA authorization enforcer")
	}

	return &Module{
		SessionIntrospectionService: sessionIntrospectionService,
		AuthzEnforcer:               opaEnforcer,
	}
}
