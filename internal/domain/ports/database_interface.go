package ports

import (
	"context"
	"database/sql"
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
	DB() (*sql.DB, error)
}
