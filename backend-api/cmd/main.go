package main

import (
	"backend-api/internal/healthcheck"
	"backend-api/internal/iam"
	"backend-api/pkg/configurer"
	"backend-api/pkg/database"
	"backend-api/pkg/logger"
	"backend-api/pkg/mailer"
	"backend-api/pkg/router"
	"backend-api/pkg/toolbox/hasher"
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

	// Database connection
	dbPool := database.MustCreatePooledConnection(configStore.Database)

	// Mailer
	mailService := mailer.NewFakeMailer()

	// Hasher
	hmacSecret := configStore.Iam.HmacSecret
	if hmacSecret == "" {
		hmacSecret = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
	}
	hmacHasher, err := hasher.NewHmacSHA256(hmacSecret)
	if err != nil {
		log.Fatal().Err(err).Msg("Error initializing HMAC hasher")
	}

	// Controllers and routes
	_ = healthcheck.NewModule(myRouter.Mux)
	_ = iam.NewModule(myRouter.Mux, dbPool.Conn, mailService, hmacHasher)

	// Create a context that is canceled on SIGINT or SIGTERM
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// Start the server
	myRouter.Start(ctx, configStore.Server.Port)
}
