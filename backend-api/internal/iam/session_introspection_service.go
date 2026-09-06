package iam

import (
	"backend-api/pkg/authz"
	"backend-api/pkg/toolbox/hasher"
	"context"
	"fmt"
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
		Permissions:          []string{}, // Permissions are not included in the session introspection response
	}
}
