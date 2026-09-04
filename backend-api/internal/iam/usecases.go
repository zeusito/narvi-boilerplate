package iam

import (
	"github.com/labstack/echo/v5"
)

type authUseCases interface {
	SendOTP(ctx *echo.Context, req *SendOTPRequest)
	VerifyOTP(ctx *echo.Context, req *VerifyOTPRequest, ipAddress, userAgent string) (*VerifyOTPResponse, error)
	Introspect(ctx *echo.Context, token string) (*PrincipalClaims, error)
}
