package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/th1enq/server_management_system/internal/configs"
	"go.uber.org/zap"
)

type IServer interface {
	Start(ctx context.Context) error
}

type server struct {
	controller *Controller
	httpConfig configs.Server
	logger     *zap.Logger
}

func NewServer(httpConfig configs.Server, logger *zap.Logger, controller *Controller) IServer {
	return &server{
		controller: controller,
		httpConfig: httpConfig,
		logger:     logger,
	}
}

func (s *server) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.httpConfig.Port),
		Handler: s.controller.RegisterRoutes(),
	}
	s.logger.Info("HTTP server starting", zap.Int("port", s.httpConfig.Port))
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.Fatal("failed to start HTTP Server", zap.Error(err))
	}
	return nil
}
