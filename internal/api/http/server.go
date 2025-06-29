package http

import (
	"context"
	"fmt"
	"net/http"

	"github.com/th1enq/server_management_system/internal/configs"
	"go.uber.org/zap"
)

type Server interface {
	Start(ctx context.Context) error
}

type server struct {
	handler    *Handler
	httpConfig configs.Server
	logger     *zap.Logger
}

func NewServer(httpConfig configs.Server, logger *zap.Logger, handler *Handler) Server {
	return &server{
		handler:    handler,
		httpConfig: httpConfig,
		logger:     logger,
	}
}

func (s *server) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", s.httpConfig.Port),
		Handler: s.handler.RegisterRoutes(),
	}
	s.logger.Info("HTTP server starting", zap.Int("port", s.httpConfig.Port))
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		s.logger.Fatal("failed to start HTTP Server", zap.Error(err))
	}
	return nil
}
