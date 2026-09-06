package iam

import (
	"errors"
	"net/http"

	"backend-api/pkg/terrors"

	"github.com/labstack/echo/v5"
)

type authnController struct {
	authService authService
	sessionSvc  sessionService
}

func newAuthnController(mux *echo.Echo, siSvc SessionIntrospectionService, sSvc sessionService, authService authService) *authnController {
	c := &authnController{
		authService: authService,
		sessionSvc:  sSvc,
	}

	sendLimiter := NewAuthRateLimiter(1.0, 5)   // 1 req/sec sustained, burst of 5
	verifyLimiter := NewAuthRateLimiter(2.0, 5) // 2 req/sec sustained, burst of 5

	authGroup := mux.Group("/v1/auth")
	authGroup.POST("/otp/send", c.handleSendOTP, sendLimiter)
	authGroup.POST("/otp/verify", c.handleVerifyOTP, verifyLimiter)
	authGroup.GET("/introspect", c.handleIntrospect, RequireAuth(siSvc))
	authGroup.DELETE("/logout", c.handleLogout, RequireAuth(siSvc))

	return c
}

func (c *authnController) handleSendOTP(ctx *echo.Context) error {
	req := new(SendOTPRequest)
	if err := ctx.Bind(req); err != nil {
		return echo.ErrBadRequest.Wrap(err)
	}

	if err := ctx.Validate(req); err != nil {
		return echo.ErrBadRequest.Wrap(err)
	}

	if err := c.authService.SendOTP(ctx.Request().Context(), req); err != nil {
		var terr *terrors.Terror
		if errors.As(err, &terr) {
			return terr.ToEchoHttpError()
		}
		return echo.ErrInternalServerError.Wrap(err)
	}

	return ctx.JSON(http.StatusOK, nil)
}

func (c *authnController) handleVerifyOTP(ctx *echo.Context) error {
	req := new(VerifyOTPRequest)
	if err := ctx.Bind(req); err != nil {
		return echo.ErrBadRequest.Wrap(err)
	}

	if err := ctx.Validate(req); err != nil {
		return echo.ErrBadRequest.Wrap(err)
	}

	resp, err := c.authService.VerifyOTP(ctx.Request().Context(), req)
	if err != nil {
		var terr *terrors.Terror
		if errors.As(err, &terr) {
			return terr.ToEchoHttpError()
		}
		return echo.ErrBadRequest.Wrap(err)
	}

	return ctx.JSON(http.StatusOK, resp)
}

func (c *authnController) handleIntrospect(ctx *echo.Context) error {
	claims := ExtractClaimsFromContext(ctx)

	return ctx.JSON(http.StatusOK, claims)
}

func (c *authnController) handleLogout(ctx *echo.Context) error {
	claims := ExtractClaimsFromContext(ctx)

	_ = c.sessionSvc.Logout(ctx.Request().Context(), claims.SessionID)

	return ctx.JSON(http.StatusOK, nil)
}
