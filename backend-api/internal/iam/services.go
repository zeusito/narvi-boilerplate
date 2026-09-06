package iam

import "context"

type authService interface {
	SendOTP(ctx context.Context, req *SendOTPRequest) error
	VerifyOTP(ctx context.Context, req *VerifyOTPRequest) (*SignInResponse, error)
}
