package router

import (
	"backend-api/pkg/configurer"
	"backend-api/pkg/logger"
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

type ChiRouter struct {
	Mux *chi.Mux
	srv *http.Server
}

func NewChiRouter(cfgs configurer.ServerConfigurations) *ChiRouter {
	router := chi.NewRouter()

	// A good base middleware stack
	router.Use(middleware.AllowContentType("application/json"))
	// TODO: Assuming a reverse proxy (Cloudflare) is set up in front of the application. Change if needed.
	router.Use(middleware.ClientIPFromHeader("CF-Connecting-IP"))
	router.Use(middleware.RequestID)
	router.Use(logger.Middleware)
	router.Use(middleware.Recoverer)

	// Set a timeout value on the request models (ctx), that will signal
	// through ctx.Done() that the request has timed out and further
	// processing should be stopped.
	router.Use(middleware.Timeout(20 * time.Second))

	// Customizing the server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%s", cfgs.Port),
		Handler:      router,
		ReadTimeout:  20 * time.Second,
		WriteTimeout: 20 * time.Second,
	}

	return &ChiRouter{
		Mux: router,
		srv: srv,
	}
}

func (s *ChiRouter) Start() {
	log.Info().Msgf("Server listening on port %s", s.srv.Addr)
	_ = s.srv.ListenAndServe()
}

func (s *ChiRouter) Shutdown(ctx context.Context) {
	log.Info().Msg("Server shutting down...")
	_ = s.srv.Shutdown(ctx)
}
