package healthcheck

import (
	"backend-api/pkg/router"
	"net/http"

	"github.com/go-chi/chi/v5"
)

type healthCheckController struct {
	// Add any fields or dependencies required for the health check
}

func newHealthCheckController(mux *chi.Mux) *healthCheckController {
	c := &healthCheckController{}

	mux.Get("/health/liveness", c.handleLiveness)

	return c
}

func (c *healthCheckController) handleLiveness(w http.ResponseWriter, req *http.Request) {
	router.RenderJSON(req.Context(), w, http.StatusOK, router.DefaultSuccessResponseBody())
}
