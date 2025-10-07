package app

import (
	"go-base-project/internal/bootstrap"
	"go-base-project/internal/config"
	"go-base-project/internal/handler"
	customMiddleware "go-base-project/internal/middleware"
	"go-base-project/internal/router"
	"go-base-project/internal/seeder"
	"go-base-project/internal/validator"
	"go-base-project/platform/database"
	"go-base-project/platform/redis"
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// App represents the main application.
type App struct {
	echo *echo.Echo
	cfg  config.Config
	db   *gorm.DB
}

// New creates a new application instance.
func New(cfg config.Config) (*App, error) {
	// Initialize Database Connection
	db, err := database.NewConnection(cfg)
	if err != nil {
		return nil, err
	}
	log.Info().Msg("Database connected successfully")

	// Run Migrations
	if err := runMigrations(db); err != nil {
		return nil, fmt.Errorf("migration failed: %w", err)
	}

	// Run Seeders
	if err := runSeeders(db, cfg); err != nil {
		return nil, fmt.Errorf("seeder failed: %w", err)
	}

	// Initialize Redis Connection
	redisClient, err := redis.NewClient(cfg.RedisURL)
	if err != nil {
		return nil, err
	}
	log.Info().Msg("Redis connected successfully")

	// Dependency Injection
	repositories := bootstrap.InitRepositories(db)
	services := bootstrap.InitServices(repositories, redisClient, cfg)
	handlers := bootstrap.InitHandlers(services, cfg)
	middlewares := customMiddleware.NewMiddleware(services.Authorization, cfg.JWTSecret)

	// Initialize Echo
	e := echo.New()
	e.Validator = validator.NewCustomValidator()
	e.HTTPErrorHandler = handler.CustomHTTPErrorHandler

	// Setup Routes
	router.SetupRoutes(e, cfg, handlers, middlewares, redisClient)

	return &App{echo: e, cfg: cfg, db: db}, nil
}

// Start runs the HTTP server and handles graceful shutdown.
func (a *App) Start() {
	go func() {
		log.Info().Msgf("Server starting on port %s", a.cfg.Port)
		if err := a.echo.Start(":" + a.cfg.Port); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("Server failed to start")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Info().Msg("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := a.echo.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("Server shutdown failed")
	}
	log.Info().Msg("Server gracefully stopped")
}

// runMigrations executes database migrations using goose.
func runMigrations(gormDB *gorm.DB) error {
	log.Info().Msg("Running database migrations...")
	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Error().Err(err).Msg("Failed to get underlying sql.DB for migrations")
		return err
	}
	if err := goose.SetDialect("postgres"); err != nil {
		log.Error().Err(err).Msg("Failed to set goose dialect")
		return err
	}
	if err := goose.Up(sqlDB, "migrations"); err != nil {
		log.Error().Err(err).Msg("Failed to run database migrations")
		return err
	}
	log.Info().Msg("Database migrations completed successfully")
	return nil
}

// runSeeders executes all necessary seeders for the application.
func runSeeders(db *gorm.DB, cfg config.Config) error {
	log.Info().Msg("Running seeders...")
	if err := seeder.SeedRolesAndPermissions(db); err != nil {
		return fmt.Errorf("failed to run RBAC seeder: %w", err)
	}
	if err := seeder.CreateAdminUser(db, cfg.AdminDefaultPassword); err != nil {
		return fmt.Errorf("failed to run admin seeder: %w", err)
	}
	log.Info().Msg("Seeders completed successfully")
	return nil
}
