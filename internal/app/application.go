package app

import (
	"context"
	"syscall"

	"github.com/th1enq/server_management_system/internal/delivery/consumers"
	"github.com/th1enq/server_management_system/internal/delivery/http"
	"github.com/th1enq/server_management_system/internal/jobs/scheduler"
	"github.com/th1enq/server_management_system/internal/utils"
	"go.uber.org/zap"
)

type Application struct {
	httpServer   http.IServer
	jobManager   scheduler.JobManager
	rootConsumer consumers.Root
	logger       *zap.Logger
}

func NewApplication(
	httpServer http.IServer,
	jobManager scheduler.JobManager,
	rootConsumer consumers.Root,
	logger *zap.Logger,
) *Application {
	return &Application{
		httpServer:   httpServer,
		jobManager:   jobManager,
		rootConsumer: rootConsumer,
		logger:       logger,
	}
}

func (app *Application) Start(ctx context.Context) error {
	app.logger.Info("Starting application...")

	app.logger.Info("Starting background job manager...")
	if err := app.jobManager.Start(ctx); err != nil {
		app.logger.Error("Failed to start job manager", zap.Error(err))
		return err
	}
	app.logger.Info("Background job manager started successfully")

	app.logger.Info("Starting HTTP server...")
	go func() {
		app.logger.Info("Starting HTTP server...")
		if err := app.httpServer.Start(ctx); err != nil {
			app.logger.Error("HTTP server failed to start", zap.Error(err))
		}
	}()
	app.logger.Info("HTTP server started successfully")

	app.logger.Info("Starting Kafka consumer manager...")
	go func() {
		consumerStartErr := app.rootConsumer.Start(ctx)
		if consumerStartErr != nil {
			app.logger.Error("Failed to start Kafka consumer manager", zap.Error(consumerStartErr))
			return
		}
		app.logger.Info("Kafka consumer manager started successfully")
	}()
	app.logger.Info("Application started successfully")

	utils.BlockUntilSignal(syscall.SIGINT, syscall.SIGTERM)

	return app.shutdown()
}

func (app *Application) shutdown() error {
	app.logger.Info("Shutting down application...")

	app.logger.Info("Stopping background job manager...")
	if err := app.jobManager.Stop(); err != nil {
		app.logger.Error("Failed to stop job manager", zap.Error(err))
	} else {
		app.logger.Info("Background job manager stopped successfully")
	}

	app.logger.Info("Application shutdown completed")
	return nil
}
