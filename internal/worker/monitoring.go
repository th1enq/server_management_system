package worker

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/th1enq/server_management_system/internal/configs"
	"github.com/th1enq/server_management_system/internal/models"
	"github.com/th1enq/server_management_system/internal/repositories"
	"go.uber.org/zap"
)

type MonitoringWorker struct {
	cfg        *configs.Config
	serverRepo repositories.ServerRepository
	logger     *zap.Logger
	stopChan   chan bool
}

func NewMonitoringWorker(cfg *configs.Config, serverRepo repositories.ServerRepository, logger *zap.Logger) *MonitoringWorker {
	return &MonitoringWorker{
		cfg:        cfg,
		serverRepo: serverRepo,
		logger:     logger,
		stopChan:   make(chan bool),
	}
}

func (w *MonitoringWorker) Start() {
	w.logger.Info("Starting monitoring worker")

	// For now, use a default interval since monitoring config might not exist
	interval := 30 * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	// Run immediately on start
	w.checkServers()

	for {
		select {
		case <-ticker.C:
			w.checkServers()
		case <-w.stopChan:
			w.logger.Info("Stopping monitoring worker")
			return
		}
	}
}

func (w *MonitoringWorker) Stop() {
	close(w.stopChan)
}

func (w *MonitoringWorker) checkServers() {
	ctx := context.Background()
	w.logger.Info("Starting server health check")

	// Get all servers
	servers, err := w.serverRepo.GetAll(ctx)
	if err != nil {
		w.logger.Error("Failed to get servers", zap.Error(err))
		return
	}

	w.logger.Info("Checking servers", zap.Int("count", len(servers)))

	for _, server := range servers {
		go w.checkServer(ctx, server)
	}
}

func (w *MonitoringWorker) checkServer(ctx context.Context, server models.Server) {
	startTime := time.Now()
	status := models.ServerStatusOff

	// Try to ping the server
	if server.IPv4 != "" {
		timeout := 5 * time.Second // Default timeout since config might not have Monitoring
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:80", server.IPv4), timeout)
		if err == nil {
			conn.Close()
			status = models.ServerStatusOn
		}
	}

	if server.Status != status {
		err := w.serverRepo.UpdateStatus(ctx, server.ServerID, status)
		if err != nil {
			w.logger.Error("Failed to update server status",
				zap.Error(err),
				zap.String("server_id", server.ServerID),
				zap.String("status", string(status)),
			)
		}
	}

	responseTime := time.Since(startTime).Milliseconds()
	w.logger.Info("Server checked",
		zap.String("server_id", server.ServerID),
		zap.String("status", string(status)),
		zap.Int64("response_time", responseTime),
	)
}
