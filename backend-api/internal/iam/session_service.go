package iam

import (
	"context"
	"fmt"

	"backend-api/pkg/authz"
	"backend-api/pkg/toolbox/hasher"
)

type SessionIntrospectionService interface {
	Introspect(ctx context.Context, token string) *PrincipalClaims
}

type defaultSessionIntrospectionService struct {
	sessionRepo sessionRepository
	hmacHasher  hasher.Hasher
}

func newSessionIntrospectionService(sessionRepo sessionRepository, hmacHasher hasher.Hasher) SessionIntrospectionService {
	return &defaultSessionIntrospectionService{
		sessionRepo: sessionRepo,
		hmacHasher:  hmacHasher,
	}
}

func (s *defaultSessionIntrospectionService) Introspect(ctx context.Context, token string) *PrincipalClaims {
	hashedToken, err := s.hmacHasher.Hash(token)
	if err != nil {
		return &PrincipalClaims{IsAuthenticated: false}
	}

	record, err := s.sessionRepo.FindActiveSessionIntrospection(ctx, hashedToken)
	if err != nil {
		return &PrincipalClaims{IsAuthenticated: false}
	}

	return &PrincipalClaims{
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
		Permissions:          authz.ExpandRolePermissions(record.OrganizationRole),
	}
}
