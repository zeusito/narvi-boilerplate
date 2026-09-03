package iam

import (
	"context"
)

type authUseCases interface {
	SendOTP(ctx context.Context, req *SendOTPRequest) (*SendOTPResponse, error)
	VerifyOTP(ctx context.Context, req *VerifyOTPRequest, ipAddress, userAgent string) (*VerifyOTPResponse, error)
	Introspect(ctx context.Context, token string) (*PrincipalClaims, error)
}
