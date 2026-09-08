package iam

import (
	"backend-api/pkg/authz"
	"context"
)

// SessionIntrospectionService this service is exposed to the outside world and is used to
// introspect a session token and retrieve the associated principal claims.
type SessionIntrospectionService interface {
	Introspect(ctx context.Context, token string) *authz.PrincipalClaims
}

type authService interface {
	SendOTP(ctx context.Context, req *SendOTPRequest) error
	VerifyOTP(ctx context.Context, req *VerifyOTPRequest) (*SignInResponse, error)
}

type sessionService interface {
	Logout(ctx context.Context, sessionID string) error
}

type orgService interface {
	Create(ctx context.Context, req *CreateOrganizationRequest) (*OrganizationResponse, error)
	Update(ctx context.Context, id string, req *UpdateOrganizationRequest) (*OrganizationResponse, error)
	GetById(ctx context.Context, id string) (*OrganizationResponse, error)
	GetAll(ctx context.Context) *OrganizationListResponse
}
