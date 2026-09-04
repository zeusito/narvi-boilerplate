package iam

import (
	"backend-api/pkg/terrors"
	"net/http"

	"github.com/labstack/echo/v5"
)

type authnController struct {
	useCases authUseCases
}

func newAuthnController(mux *echo.Echo, useCases authUseCases, middleware *authnMiddleware) *authnController {
	c := &authnController{
		useCases: useCases,
	}

	authGroup := mux.Group("/v1/auth")
	authGroup.POST("/otp/send", c.handleSendOTP)
	authGroup.POST("/otp/verify", c.handleVerifyOTP)
	authGroup.GET("/introspect", c.handleIntrospect, middleware.RequireAuth)

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

	resp, err := c.useCases.SendOTP(ctx, req)
	if err != nil {
		if tErr, ok := err.(*terrors.Terror); ok {
			return tErr.ToEchoHttpError()
		}
		return echo.ErrInternalServerError.Wrap(err)
	}

	return ctx.JSON(http.StatusOK, resp)
}

func (c *authnController) handleVerifyOTP(ctx *echo.Context) error {
	req := new(VerifyOTPRequest)
	if err := ctx.Bind(req); err != nil {
		return echo.ErrBadRequest.Wrap(err)
	}

	if err := ctx.Validate(req); err != nil {
		return echo.ErrBadRequest.Wrap(err)
	}

	ipAddress := ctx.RealIP()
	userAgent := ctx.Request().UserAgent()

	resp, err := c.useCases.VerifyOTP(ctx, req, ipAddress, userAgent)
	if err != nil {
		if tErr, ok := err.(*terrors.Terror); ok {
			return tErr.ToEchoHttpError()
		}
		return echo.ErrInternalServerError.Wrap(err)
	}

	return ctx.JSON(http.StatusOK, resp)
}

func (c *authnController) handleIntrospect(ctx *echo.Context) error {
	claims := GetPrincipal(ctx)
	if claims == nil {
		return terrors.UnAuthorized("unauthenticated").ToEchoHttpError()
	}

	return ctx.JSON(http.StatusOK, claims)
}
