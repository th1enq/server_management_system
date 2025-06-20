package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/th1enq/server_management_system/internal/api"
	"github.com/th1enq/server_management_system/internal/config"
	"github.com/th1enq/server_management_system/internal/database"
	"github.com/th1enq/server_management_system/internal/wire"
	"github.com/th1enq/server_management_system/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	config, err := config.Load()
	if err != nil {
		panic(fmt.Sprintf("failed to load configuration: %v", err))
	}

	err = logger.Load(config)
	if err != nil {
		panic(fmt.Sprintf("failed to load logger: %v", err))
	}
	defer logger.Sync()

	logger.Info("Starting Viettel Server Management System",
		zap.String("version", "1.0.0"),
		zap.String("env", config.Server.Env))

	app, err := wire.InitializeApp(config)
	if err != nil {
		logger.Fatal("failed to initialize application", err)
	}

	if config.Server.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(gin.Recovery())
	api.SetupRoutes(router, app)

	go app.MonitoringWorker.Start()

	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", config.Server.Port),
		Handler: router,
	}

	go func() {
		logger.Info("HTTP server starting", zap.Int("port", config.Server.Port))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("failed to start HTTP Server", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Error("Server force to shut down", err)
	}

	database.Close()

	logger.Info("Server exited")
}
