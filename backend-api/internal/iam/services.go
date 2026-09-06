package iam

import "context"

// SessionIntrospectionService this service is exposed to the outside world and is used to
// introspect a session token and retrieve the associated principal claims.
type SessionIntrospectionService interface {
	Introspect(ctx context.Context, token string) *PrincipalClaims
}

type authService interface {
	SendOTP(ctx context.Context, req *SendOTPRequest) error
	VerifyOTP(ctx context.Context, req *VerifyOTPRequest) (*SignInResponse, error)
}

type sessionService interface {
	Logout(ctx context.Context, sessionID string) error
}
