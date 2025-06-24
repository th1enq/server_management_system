// @title VCS Server Management System API
// @version 1.0
// @description A comprehensive server management system for monitoring and reporting server status
// @termsOfService http://swagger.io/terms/

// @contact.name API Support
// @contact.url http://www.swagger.io/support
// @contact.email support@swagger.io

// @license.name Apache 2.0
// @license.url http://www.apache.org/licenses/LICENSE-2.0.html

// @host localhost:8080

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Type "Bearer" followed by a space and JWT token.

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

	_ "github.com/th1enq/server_management_system/docs"
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
	go app.ReportService.StartDailyReportScheduler()

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
	app.Redis.Close()
	app.PGP.Close()
	app.ReportService.StopDailyReportScheduler()
	app.MonitoringWorker.Stop()

	logger.Info("Server exited")
}
