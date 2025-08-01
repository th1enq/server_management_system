package database

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/gob"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/th1enq/server_management_system/internal/configs"
	"github.com/th1enq/server_management_system/internal/infrastructure/models"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
	Scan(dest interface{}) error
	Count(count *int64) error
	Create(value interface{}) error
	Update(column string, value interface{}) error
	Updates(values interface{}) error
	Exec(query string, args ...interface{}) error
	Pluck(column string, dest interface{}) error
	BatchCreateOnConflict(value interface{}, dest interface{}) error
	Select(query string, args ...interface{}) DatabaseClient
	DB() (*sql.DB, error)

	Transaction(ctx context.Context, fn func(tx DatabaseClient) error) error

	GetRecordsByLockID(lockID string) ([]models.Record, error)
	UpdateRecordLockByState(lockID string, lockedOn time.Time, state models.RecordState) error
	UpdateRecordByID(message models.Record) error
	ClearLocksWithDurationBeforeDate(time time.Time) error
	ClearLocksByLockID(lockID string) error
	RemoveRecordsBeforeDatetime(expiryTime time.Time) error
}

type gormDatabase struct {
	client *gorm.DB
}

func (p *gormDatabase) Transaction(ctx context.Context, fn func(tx DatabaseClient) error) error {
	err := p.client.Transaction(func(tx *gorm.DB) error {
		txClient := &gormDatabase{client: tx}
		if err := fn(txClient); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return fmt.Errorf("transaction failed: %w", err)
	}
	return nil
}

func (p *gormDatabase) Updates(values interface{}) error {
	if err := p.client.Updates(values).Error; err != nil {
		return fmt.Errorf("failed to update records: %w", err)
	}
	return nil
}

func (p *gormDatabase) ClearLocksByLockID(lockID string) error {
	err := p.client.Model(&models.Record{}).
		Where("lock_id = ?", lockID).
		Updates(map[string]interface{}{
			"locked_at": nil,
			"lock_id":   nil,
		}).Error
	if err != nil {
		return fmt.Errorf("failed to clear locks by lock ID: %w", err)
	}
	return nil
}

// ClearLocksWithDurationBeforeDate implements DatabaseClient.
func (p *gormDatabase) ClearLocksWithDurationBeforeDate(time time.Time) error {
	err := p.client.Model(&models.Record{}).
		Where("locked_at < ?", time).
		Updates(map[string]interface{}{
			"locked_at": nil,
			"lock_id":   nil,
		}).Error
	if err != nil {
		return fmt.Errorf("failed to clear locks before date: %w", err)
	}
	return nil
}

// GetRecordsByLockID implements DatabaseClient.
func (p *gormDatabase) GetRecordsByLockID(lockID string) ([]models.Record, error) {
	rows, err := p.client.Model(&models.Record{}).
		Where("lock_id = ?", lockID).
		Find(&[]models.Record{}).Rows()
	if err != nil {
		return nil, fmt.Errorf("failed to get records by lock ID: %w", err)
	}
	defer rows.Close()

	var messages []models.Record
	for rows.Next() {
		var rec models.Record
		var data []byte
		if err := rows.Scan(&rec.ID, &data, &rec.State, &rec.CreatedAt, &rec.LockID, &rec.LockedAt, &rec.ProcessedAt, &rec.NumberOfAttempts, &rec.LastAttemptAt, &rec.Error); err != nil {
			if err == sql.ErrNoRows {
				return messages, nil
			}
			return messages, fmt.Errorf("failed to scan record: %w", err)
		}
		decErr := gob.NewDecoder(bytes.NewBuffer(data)).Decode(&rec.Message)
		if decErr != nil {
			return nil, fmt.Errorf("failed to decode message: %w", decErr)
		}
		messages = append(messages, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating over rows: %w", err)
	}
	return messages, nil
}

// RemoveRecordsBeforeDatetime implements DatabaseClient.
func (p *gormDatabase) RemoveRecordsBeforeDatetime(expiryTime time.Time) error {
	err := p.client.Model(&models.Record{}).
		Where("created_at < ?", expiryTime).
		Delete(&models.Record{}).Error
	if err != nil {
		return fmt.Errorf("failed to remove records before datetime: %w", err)
	}
	return nil
}

// UpdateRecordByID implements DatabaseClient.
func (p *gormDatabase) UpdateRecordByID(message models.Record) error {
	msgData := new(bytes.Buffer)
	enc := gob.NewEncoder(msgData)
	if err := enc.Encode(message); err != nil {
		return fmt.Errorf("failed to encode message: %w", err)
	}
	err := p.client.Model(&models.Record{}).
		Where("id = ?", message.ID).
		Updates(map[string]interface{}{
			"message":            msgData.Bytes(),
			"state":              message.State,
			"locked_at":          message.LockedAt,
			"lock_id":            message.LockID,
			"processed_at":       message.ProcessedAt,
			"number_of_attempts": message.NumberOfAttempts,
			"last_attempt_at":    message.LastAttemptAt,
			"error":              message.Error,
		}).Error
	if err != nil {
		return fmt.Errorf("failed to update record by ID: %w", err)
	}
	return nil
}

func (p *gormDatabase) UpdateRecordLockByState(lockID string, lockedOn time.Time, state models.RecordState) error {
	err := p.client.Model(&models.Record{}).
		Where("state = ?", state).
		Updates(map[string]interface{}{
			"locked_at": lockedOn,
			"lock_id":   lockID,
		}).Error
	if err != nil {
		return fmt.Errorf("failed to update record lock by state: %w", err)
	}
	return nil
}

func (p *gormDatabase) Count(count *int64) error {
	if err := p.client.Count(count).Error; err != nil {
		return fmt.Errorf("failed to count records: %w", err)
	}
	return nil
}

func (p *gormDatabase) Scan(dest interface{}) error {
	if err := p.client.Scan(dest).Error; err != nil {
		return fmt.Errorf("failed to scan records: %w", err)
	}
	return nil
}

func (p *gormDatabase) BatchCreateOnConflict(value interface{}, dest interface{}) error {
	val := reflect.ValueOf(value)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Slice || val.Len() == 0 {
		return nil
	}

	servers := make([]models.Server, val.Len())
	for i := 0; i < val.Len(); i++ {
		if item := val.Index(i); item.CanInterface() {
			if server, ok := item.Interface().(models.Server); ok {
				servers[i] = server
			} else if item.Kind() == reflect.Ptr {
				if server, ok := item.Elem().Interface().(models.Server); ok {
					servers[i] = server
				}
			}
		}
	}

	// Build raw query
	placeholders := make([]string, len(servers))
	values := make([]interface{}, 0, len(servers)*9)

	for i, server := range servers {
		placeholders[i] = fmt.Sprintf("($%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d,$%d)",
			i*9+1, i*9+2, i*9+3, i*9+4, i*9+5, i*9+6, i*9+7, i*9+8, i*9+9)
		values = append(values,
			server.ServerID, server.ServerName, server.Status,
			server.IPv4, server.Description, server.Location,
			server.OS, server.IntervalTime, server.CreatedAt)
	}

	query := fmt.Sprintf(`
		INSERT INTO servers (server_id, server_name, status, ipv4, description, location, os, interval_time, created_at)
		VALUES %s
		ON CONFLICT (server_id) DO NOTHING
		RETURNING *`, strings.Join(placeholders, ","))

	return p.client.Raw(query, values...).Scan(dest).Error
}

func (p *gormDatabase) Create(value interface{}) error {
	if err := p.client.Create(value).Error; err != nil {
		return fmt.Errorf("failed to create record: %w", err)
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

func (p *gormDatabase) Pluck(column string, dest interface{}) error {
	if err := p.client.Pluck(column, dest).Error; err != nil {
		return fmt.Errorf("failed to pluck column %s: %w", column, err)
	}
	return nil
}

func (p *gormDatabase) Select(query string, args ...interface{}) DatabaseClient {
	return &gormDatabase{
		client: p.client.Select(query, args...),
	}
}

func (p *gormDatabase) Clauses(conds ...interface{}) DatabaseClient {
	exprs := make([]clause.Expression, len(conds))
	for i, cond := range conds {
		exprs[i], _ = cond.(clause.Expression)
	}
	return &gormDatabase{
		client: p.client.Clauses(exprs...),
	}
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
