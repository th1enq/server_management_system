package worker

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/th1enq/server_management_system/internal/config"
	"github.com/th1enq/server_management_system/internal/models"
	"github.com/th1enq/server_management_system/internal/repositories"
	"github.com/th1enq/server_management_system/pkg/logger"
	"go.uber.org/zap"
)

type MonitoringWorker struct {
	cfg        *config.Config
	serverRepo repositories.ServerRepository
	stopChan   chan bool
}

func NewMonitoringWorker(cfg *config.Config, serverRepo repositories.ServerRepository, stopChan chan bool) *MonitoringWorker {
	return &MonitoringWorker{
		cfg:        cfg,
		serverRepo: serverRepo,
		stopChan:   make(chan bool),
	}
}

func (w *MonitoringWorker) Start() {
	logger.Info("Starting monitoring worker",
		zap.Duration("interval", w.cfg.Monitoring.Interval),
	)

	ticker := time.NewTicker(w.cfg.Monitoring.Interval)
	defer ticker.Stop()

	// Run immediately on start
	w.checkServers()

	for {
		select {
		case <-ticker.C:
			w.checkServers()
		case <-w.stopChan:
			logger.Info("Stopping monitoring worker")
			return
		}
	}
}

func (w *MonitoringWorker) Stop() {
	close(w.stopChan)
}

func (w *MonitoringWorker) checkServers() {
	ctx := context.Background()
	logger.Info("Starting server health check")

	// Get all servers
	servers, err := w.serverRepo.GetAll(ctx)
	if err != nil {
		logger.Error("Failed to get servers", err)
		return
	}

	logger.Info("Checking servers", zap.Int("count", len(servers)))

	for _, server := range servers {
		go w.checkServer(ctx, server)
	}
}

func (w *MonitoringWorker) checkServer(ctx context.Context, server models.Server) {
	startTime := time.Now()
	status := models.ServerStatusOff
	responseTime := -1

	// Try to ping the server
	if server.IPv4 != "" {
		timeout := w.cfg.Monitoring.PingTimeout
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:80", server.IPv4), timeout)
		if err == nil {
			conn.Close()
			status = models.ServerStatusOn
			responseTime = int(time.Since(startTime).Milliseconds())
		}
	}

	if server.Status != status {
		err := w.serverRepo.UpdateStatus(ctx, server.ServerID, status)
		if err != nil {
			logger.Error("Failed to update server status", err,
				zap.String("server_id", server.ServerID),
				zap.String("status", string(status)),
			)
		}
	}

	logger.Info("Server checked",
		zap.String("server_id", server.ServerID),
		zap.String("status", string(status)),
		zap.Int("response_time", responseTime),
	)
}
