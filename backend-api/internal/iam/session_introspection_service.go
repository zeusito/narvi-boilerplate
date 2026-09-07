package iam

import (
	"backend-api/pkg/authz"
	"backend-api/pkg/toolbox/hasher"
	"context"
	"fmt"

	"github.com/rs/zerolog/log"
)

type DefaultSessionIntrospectionService struct {
	sessionRepo sessionRepository
	hmacHasher  hasher.Hasher
}

func newSessionIntrospectionService(sessionRepo sessionRepository, hmacHasher hasher.Hasher) SessionIntrospectionService {
	return &DefaultSessionIntrospectionService{
		sessionRepo: sessionRepo,
		hmacHasher:  hmacHasher,
	}
}

func (s *DefaultSessionIntrospectionService) Introspect(ctx context.Context, token string) *authz.PrincipalClaims {
	hashedToken, err := s.hmacHasher.Hash(token)
	if err != nil {
		return &authz.PrincipalClaims{IsAuthenticated: false}
	}

	record, err := s.sessionRepo.FindActiveSessionIntrospection(ctx, hashedToken)
	if err != nil {
		return &authz.PrincipalClaims{IsAuthenticated: false}
	}

	if record.IdentityState != "active" {
		log.Ctx(ctx).Warn().Msgf("identity '%s' is not active", record.IdentityID)
		return &authz.PrincipalClaims{IsAuthenticated: false}
	}

	return &authz.PrincipalClaims{
		IsAuthenticated:      true,
		SessionID:            record.SessionID,
		IdentityID:           record.IdentityID,
		Email:                record.IdentityEmail,
		FullName:             fmt.Sprintf("%s %s", record.IdentityFirstName, record.IdentityLastName),
		ActiveOrganizationID: record.OrganizationID,
		OrganizationName:     record.OrganizationName,
		OrganizationSlug:     record.OrganizationSlug,
		OrganizationLogo:     record.OrganizationLogo,
		OrganizationRole:     record.OrganizationRole,
	}
}
