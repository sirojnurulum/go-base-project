package main

import (
	"context"
	"go-base-project/docs"
	"go-base-project/internal/app"
	"go-base-project/internal/bootstrap"
	"go-base-project/internal/config"
	"go-base-project/platform/logger"
	"strings"

	"github.com/rs/zerolog/log"
)

// @title Go Base Project API
// @version 1.0
// @description This is the API documentation for the Go Base Project backend.
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

	// 2. Set Swagger host dynamically from BACKEND_URL
	// Remove protocol (http:// or https://) from BACKEND_URL
	backendHost := strings.TrimPrefix(cfg.BackendURL, "https://")
	backendHost = strings.TrimPrefix(backendHost, "http://")
	docs.SwaggerInfo.Host = backendHost

	// 3. Initialize Logger
	logger.Init("production") // Always use production-ready logging

	// Log the loaded configuration for debugging. Only display safe values.
	log.Info().
		Str("port", cfg.Port).
		Str("frontend_url", cfg.FrontendURL).
		Str("backend_url", cfg.BackendURL).
		Str("swagger_host", docs.SwaggerInfo.Host).
		Msg("Loaded application configuration")

	// 4. Initialize Tracer Provider
	tp, err := bootstrap.InitTracerProvider("go-base-project", cfg.EnableDetailedTracing)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize tracer provider")
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			log.Error().Err(err).Msg("Failed to shutdown tracer provider")
		}
	}()

	// 5. Create and run the application
	application, err := app.New(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to initialize application")
	}

	application.Start()
}
