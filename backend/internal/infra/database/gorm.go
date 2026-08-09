package database

import (
	"fmt"

	gormdb "github.com/raykavin/gobox/database/gorm"

	"gorm.io/gorm"
)

// Open opens a single GORM connection from a DatabaseProvider.
func Open(dsn, dialector, logLevel string) (*gorm.DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("dsn is required")
	}
	if dialector == "" {
		return nil, fmt.Errorf("dialector is required")
	}

	level := "info"
	if logLevel != "" {
		level = logLevel
	}

	cfg := gormdb.DefaultGormConfig()
	cfg.DSN = dsn
	cfg.Dialector = dialector
	cfg.LogLevel = level

	return gormdb.New(cfg)
}

// Migrate runs auto-migration for the given domain models. It is a
// no-op when no models are provided, so callers decide what to migrate.
func Migrate(db *gorm.DB, models ...any) error {
	if len(models) == 0 {
		return nil
	}
	if err := db.AutoMigrate(models...); err != nil {
		return fmt.Errorf("auto-migrate models: %w", err)
	}
	return nil
}
