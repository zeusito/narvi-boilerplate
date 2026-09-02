package healthcheck

import "github.com/labstack/echo/v5"

type healthCheckController struct {
	// Add any fields or dependencies required for the health check
}

func newHealthCheckController(mux *echo.Echo) *healthCheckController {
	c := &healthCheckController{}

	mux.GET("/health/liveness", c.handleLiveness)

	return c
}

func (c *healthCheckController) handleLiveness(ctx *echo.Context) error {
	return ctx.JSON(200, map[string]string{"status": "ok"})
}
