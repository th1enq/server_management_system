//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/google/wire"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/th1enq/server_management_system/internal/config"
	"github.com/th1enq/server_management_system/internal/database"
	"github.com/th1enq/server_management_system/internal/handler"
	"github.com/th1enq/server_management_system/internal/repositories"
	"github.com/th1enq/server_management_system/internal/services"
	"github.com/th1enq/server_management_system/internal/worker"

	"gorm.io/gorm"
)

type App struct {
	Config           *config.Config
	DB               *gorm.DB
	Redis            *redis.Client
	PGP              *pgxpool.Pool
	ServerHandler    *handler.ServerHandler
	MonitoringWorker *worker.MonitoringWorker
}

func InitializeApp(config *config.Config) (*App, error) {
	wire.Build(
		// Database
		provideDB,
		providePgx,
		provideRedis,

		// Repositories
		repositories.NewServerRepository,

		// Services
		services.NewServerService,

		// Handlers
		handler.NewServerHandler,

		// Worker
		provideMonitoringWorker,

		// App
		wire.Struct(new(App), "*"),
	)
	return nil, nil
}

func provideDB(config *config.Config) (*gorm.DB, error) {
	err := database.LoadDB(config)
	if err != nil {
		return nil, err
	}
	return database.DB, nil
}

func providePgx(config *config.Config) (*pgxpool.Pool, error) {
	err := database.LoadPgPool(config)
	if err != nil {
		return nil, err
	}
	return database.PgPool, nil
}

func provideRedis(config *config.Config) (*redis.Client, error) {
	err := database.LoadRedis(config)
	if err != nil {
		return nil, err
	}
	return database.RedisClient, nil
}

func provideMonitoringWorker(config *config.Config, serverRepo repositories.ServerRepository) *worker.MonitoringWorker {
	stopChan := make(chan bool)
	return worker.NewMonitoringWorker(config, serverRepo, stopChan)
}
