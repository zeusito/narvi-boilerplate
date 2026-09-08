package logger

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// MustConfigure initializes global zerolog settings and sets the default context logger fallback.
func MustConfigure() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	zerolog.TimeFieldFormat = time.RFC3339
	log.Logger = log.With().Caller().Logger()
	zerolog.DefaultContextLogger = &log.Logger
}

// Middleware injects a sub-logger enriched with the request traceId into the request context.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		reqID := middleware.GetReqID(ctx)

		subLogger := log.Logger.With().Logger()
		if reqID != "" {
			subLogger = log.With().Str("traceId", reqID).Logger()
		}

		ctx = subLogger.WithContext(ctx)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
