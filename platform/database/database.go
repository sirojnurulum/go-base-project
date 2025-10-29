package database

import (
	"go-base-project/internal/config"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// NewConnection creates and returns a new GORM database connection.
func NewConnection(cfg config.Config) (*gorm.DB, error) {
	// Use Info level by default, or Warn if detailed tracing is disabled
	logLevel := logger.Info
	if !cfg.EnableDetailedTracing {
		logLevel = logger.Warn // Reduce log verbosity
	}

	db, err := gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	// Configure Connection Pool
	sqlDB.SetMaxIdleConns(cfg.DBMaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.DBMaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.DBConnMaxLifetime)
	sqlDB.SetConnMaxIdleTime(cfg.DBConnMaxIdleTime)

	return db, nil
}
