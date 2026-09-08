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
	"time"

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
	myRouter := router.NewChiRouter(configStore.Server)

	// Database connection
	dbPool := database.MustCreatePooledConnection(configStore.Database)

	// Mailer
	mailService := mailer.NewResendMailer(configStore.Email)

	// Hasher
	hmacHasher, err := hasher.NewHmacSHA256(configStore.Iam.HmacSecret)
	if err != nil {
		log.Fatal().Err(err).Msg("Error initializing HMAC hasher")
	}

	// Modules
	_ = healthcheck.NewModule(myRouter.Mux)
	_ = iam.NewModule(myRouter.Mux, dbPool.Conn, mailService, hmacHasher)

	// Start server in background
	go myRouter.Start()

	// Graceful shutdown
	gracefulShutdown(myRouter)
}

func gracefulShutdown(myRouter *router.ChiRouter) {
	// Wait for the interrupt signal to gracefully shut down the server with a timeout of 10 seconds.
	// Use a buffered channel to avoid missing signals as recommended for signal.Notify
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	// Signal acquired, starting to shut down all systems
	log.Warn().Msg("Shutting down gracefully...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	myRouter.Shutdown(ctx)
}
