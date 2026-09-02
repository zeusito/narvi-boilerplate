package main

import (
	"backend-api/internal/healthcheck"
	"backend-api/pkg/configurer"
	"backend-api/pkg/logger"
	"backend-api/pkg/router"
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
)

func main() {
	// Parse flags
	cfgPath := flag.String("config", "resources/config.toml", "Path to the configuration file")
	flag.Parse()

	// Setup logger
	logger.MustConfigure()

	// Load configurations
	configStore, err := configurer.LoadConfigurations(*cfgPath)
	if err != nil {
		log.Fatal().Err(err).Msg("Error loading configurations")
	}

	// Init Http router
	myRouter := router.NewRouter()

	// Controllers and routes
	_ = healthcheck.NewModule(myRouter.Mux)

	// Create a context that is canceled on SIGINT or SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start the server
	myRouter.Start(ctx, configStore.Server.Port)
}
