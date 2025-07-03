package database

import (
	"fmt"
	"server_management_system/pkg/config"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func LoadPostgres(cfg config.Postgres, log *zap.Logger) (*gorm.DB, func(), error) {
	dsn := cfg.GetPostgresDSN()

	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Error("Failed to connect to database", zap.Error(err))
		return nil, nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Error("Failed to get database instance", zap.Error(err))
		return nil, nil, fmt.Errorf("failed to get database instance: %w", err)
	}
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)

	cleanup := func() {
		log.Info("Closing database connection")
		sqlDB.Close()
	}

	log.Info("Database connected successfully",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.Name),
	)
	return gormDB, cleanup, nil
}
