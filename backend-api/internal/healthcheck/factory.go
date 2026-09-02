package healthcheck

import "github.com/labstack/echo/v5"

type Module struct {
	// Add any fields or dependencies required for the health check module
}

// NewModule creates a new instance of the health check module
func NewModule(mux *echo.Echo) *Module {
	_ = newHealthCheckController(mux)

	return &Module{}
}
