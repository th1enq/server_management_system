package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/th1enq/server_management_system/internal/configs"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormLogger "gorm.io/gorm/logger"
)

type DatabaseClient interface {
	WithContext(ctx context.Context) DatabaseClient
	Where(query interface{}, args ...interface{}) DatabaseClient
	Limit(limit int) DatabaseClient
	Offset(offset int) DatabaseClient
	Model(value interface{}) DatabaseClient
	Order(value interface{}) DatabaseClient
	First(dest interface{}, conds ...interface{}) error
	Delete(value interface{}, conds ...interface{}) error
	Find(dest interface{}, conds ...interface{}) error
	Save(value interface{}) error
	Count(count *int64) error
	CreateInBatches(value interface{}, batchSize int) error
	Create(value interface{}) error
	Update(column string, value interface{}) error
	Exec(query string, args ...interface{}) error
	DB() (*sql.DB, error)
}

type gormDatabase struct {
	client *gorm.DB
}

func (p *gormDatabase) Count(count *int64) error {
	if err := p.client.Count(count).Error; err != nil {
		return fmt.Errorf("failed to count records: %w", err)
	}
	return nil
}

func (p *gormDatabase) Create(value interface{}) error {
	if err := p.client.Create(value).Error; err != nil {
		return fmt.Errorf("failed to create record: %w", err)
	}
	return nil
}

func (p *gormDatabase) CreateInBatches(value interface{}, batchSize int) error {
	if err := p.client.CreateInBatches(value, batchSize).Error; err != nil {
		return fmt.Errorf("failed to create records in batches: %w", err)
	}
	return nil
}

func (p *gormDatabase) Delete(value interface{}, conds ...interface{}) error {
	if err := p.client.Delete(value, conds...).Error; err != nil {
		return fmt.Errorf("failed to delete record: %w", err)
	}
	return nil
}

func (p *gormDatabase) Find(dest interface{}, conds ...interface{}) error {
	if err := p.client.Find(dest, conds...).Error; err != nil {
		return fmt.Errorf("failed to find records: %w", err)
	}
	return nil
}

func (p *gormDatabase) First(dest interface{}, conds ...interface{}) error {
	if err := p.client.First(dest, conds...).Error; err != nil {
		return fmt.Errorf("failed to find first record: %w", err)
	}
	return nil
}

func (p *gormDatabase) Limit(limit int) DatabaseClient {
	return &gormDatabase{
		client: p.client.Limit(limit),
	}
}

func (p *gormDatabase) Model(value interface{}) DatabaseClient {
	return &gormDatabase{
		client: p.client.Model(value),
	}
}

func (p *gormDatabase) Offset(offset int) DatabaseClient {
	return &gormDatabase{
		client: p.client.Offset(offset),
	}
}

func (p *gormDatabase) Order(value interface{}) DatabaseClient {
	return &gormDatabase{
		client: p.client.Order(value),
	}
}

func (p *gormDatabase) Save(value interface{}) error {
	if err := p.client.Save(value).Error; err != nil {
		return fmt.Errorf("failed to save record: %w", err)
	}
	return nil
}

func (p *gormDatabase) Where(query interface{}, args ...interface{}) DatabaseClient {
	return &gormDatabase{
		client: p.client.Where(query, args...),
	}
}

func (p *gormDatabase) Exec(query string, args ...interface{}) error {
	if err := p.client.Exec(query, args...).Error; err != nil {
		return fmt.Errorf("failed to execute query: %w", err)
	}
	return nil
}

func (p *gormDatabase) WithContext(ctx context.Context) DatabaseClient {
	return &gormDatabase{
		client: p.client.WithContext(ctx),
	}
}

func (p *gormDatabase) DB() (*sql.DB, error) {
	sqlDB, err := p.client.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}
	return sqlDB, nil
}

func (p *gormDatabase) Update(column string, value interface{}) error {
	if err := p.client.Update(column, value).Error; err != nil {
		return fmt.Errorf("failed to update record: %w", err)
	}
	return nil
}

func NewDatabase(cfg configs.Database, logger *zap.Logger) (DatabaseClient, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=disable",
		cfg.Host,
		cfg.User,
		cfg.Password,
		cfg.Name,
		cfg.Port)

	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger.Default.LogMode(gormLogger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database instance: %w", err)
	}
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)

	logger.Info("Database connected successfully",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.Name),
	)
	return &gormDatabase{client: gormDB}, nil
}

func NewDatabaseWithGorm(gormDB *gorm.DB) DatabaseClient {
	if gormDB == nil {
		return nil
	}
	return &gormDatabase{client: gormDB}
}
