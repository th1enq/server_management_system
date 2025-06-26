package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/th1enq/server_management_system/internal/configs"
	"go.uber.org/zap"
)

type PgxPoolInterface interface {
	CopyFrom(ctx context.Context, tableName []string, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error)
}

type pgxPoolWrapper struct {
	Pool *pgxpool.Pool
}

func (w *pgxPoolWrapper) CopyFrom(ctx context.Context, tableName []string, columnNames []string, rowSrc pgx.CopyFromSource) (int64, error) {
	return w.Pool.CopyFrom(ctx, tableName, columnNames, rowSrc)
}

func LoadPgPool(cfg configs.Database, logger *zap.Logger) (PgxPoolInterface, func(), error) {
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.Name,
	)

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create pgxpool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, nil, fmt.Errorf("failed to connect pgxpool: %w", err)
	}

	PgPool := &pgxPoolWrapper{Pool: pool}

	cleanup := func() {
		PgPool.Pool.Close()
	}

	logger.Info("PgxPool connected successfully",
		zap.String("host", cfg.Host),
		zap.Int("port", cfg.Port),
		zap.String("database", cfg.Name),
	)

	return PgPool, cleanup, nil
}
