package iam

import (
	"net/http"

	"github.com/labstack/echo/v5"
)

type authnController struct {
	useCases authUseCases
}

func newAuthnController(mux *echo.Echo, sm SessionManager, useCases authUseCases) *authnController {
	c := &authnController{
		useCases: useCases,
	}

	authGroup := mux.Group("/v1/auth")
	authGroup.POST("/otp/send", c.handleSendOTP)
	authGroup.POST("/otp/verify", c.handleVerifyOTP)
	authGroup.GET("/introspect", c.handleIntrospect, RequireAuth(sm))

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

	c.useCases.SendOTP(ctx, req)

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

	resp, err := c.useCases.VerifyOTP(ctx, req)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "code is not valid")
	}

	return ctx.JSON(http.StatusOK, resp)
}

func (c *authnController) handleIntrospect(ctx *echo.Context) error {
	claims := ExtractClaimsFromContext(ctx)

	return ctx.JSON(http.StatusOK, claims)
}
