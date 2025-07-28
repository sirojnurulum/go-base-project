package main

import (
	"beresin-backend/internal/app"
	"beresin-backend/internal/config"
	"beresin-backend/platform/logger"

	"github.com/rs/zerolog/log"

	_ "beresin-backend/docs" // swagger docs
)

// @title Beresin App API
// @version 1.0
// @description This is the API documentation for the Beresin App backend.
// @host localhost:8080
// @BasePath /api
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// 1. Load Configuration
	cfg, err := config.Load()
	if err != nil {
		// Use standard logger as zerolog is not yet initialized
		log.Fatal().Err(err).Msg("Failed to load configuration")
	}

	// 2. Initialize Logger
	logger.Init(cfg.Env)

	// Log the loaded configuration for debugging. Only display safe values.
	log.Info().
		Str("environment", cfg.Env).
		Str("port", cfg.Port).
		Str("frontend_url", cfg.FrontendURL).
		Msg("Loaded application configuration")

	// 3. Create and run the application
	application, err := app.New(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize application")
	}

	application.Start()
}
