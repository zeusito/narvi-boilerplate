package router

import (
	"context"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"github.com/rs/zerolog/log"
)

type HttpRouter struct {
	Mux *echo.Echo
}

func NewRouter() *HttpRouter {
	e := echo.New()

	// Middlewares
	e.Use(middleware.BodyLimit(10_485_760)) // 10 MB
	e.Use(middleware.ContextTimeout(60 * time.Second))
	e.Use(middleware.RequestID())
	e.Use(middleware.Gzip())
	e.Use(middleware.Recover())

	// Other configurations
	e.IPExtractor = echo.ExtractIPFromXFFHeader()
	e.Validator = &CustomValidator{validator: validator.New()}

	return &HttpRouter{Mux: e}
}

func (s *HttpRouter) Start(ctx context.Context, addr string) {
	log.Info().Msgf("Server listening on port %s", addr)

	sc := echo.StartConfig{
		Address:         addr,
		GracefulTimeout: 10 * time.Second,
	}

	if err := sc.Start(ctx, s.Mux); err != nil {
		log.Fatal().Err(err).Msg("Failed to start server")
	}
}
