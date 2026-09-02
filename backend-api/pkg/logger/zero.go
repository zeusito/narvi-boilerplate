package logger

import (
	"time"

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
