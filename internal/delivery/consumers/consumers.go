package consumers

import (
	"context"
	"encoding/json"

	"github.com/th1enq/server_management_system/internal/infrastructure/mq/consumer"
	"github.com/th1enq/server_management_system/internal/infrastructure/mq/producer"
	"go.uber.org/zap"
)

type Root interface {
	Start(ctx context.Context) error
}

type root struct {
	metricsUpdate    MetricsUpdate
	statusUpdate     UpdateStatus
	uptimeCalculator uptimeCalculatorConsumer
	uptimeConsumer   consumer.UptimeConsumer
	metricsConsumer  consumer.MetricsConsumer
	serverConsumer   consumer.ServerConsumer
	logger           *zap.Logger
}

func NewRoot(
	metricsUpdate MetricsUpdate,
	statusUpdate UpdateStatus,
	uptimeCalculator uptimeCalculatorConsumer,
	uptimeConsumer consumer.UptimeConsumer,
	serverConsumer consumer.ServerConsumer,
	metricsConsumer consumer.MetricsConsumer,
	logger *zap.Logger,
) Root {
	return &root{
		metricsUpdate:    metricsUpdate,
		statusUpdate:     statusUpdate,
		uptimeCalculator: uptimeCalculator,
		uptimeConsumer:   uptimeConsumer,
		metricsConsumer:  metricsConsumer,
		serverConsumer:   serverConsumer,
		logger:           logger,
	}
}

func (r root) Start(ctx context.Context) error {
	r.logger.Info("Starting consumers")

	r.serverConsumer.RegisterHandler(
		producer.MessageMonitoring,
		func(ctx context.Context, queueName string, payload []byte) error {
			var event producer.Message
			if err := json.Unmarshal(payload, &event); err != nil {
				return err
			}
			return r.statusUpdate.Handle(ctx, event)
		},
	)

	r.uptimeConsumer.RegisterHandler(
		producer.MessageStatusChange,
		func(ctx context.Context, queueName string, payload []byte) error {
			var event producer.StatusChangeMessage
			if err := json.Unmarshal(payload, &event); err != nil {
				return err
			}
			return r.uptimeCalculator.Handle(ctx, event)
		},
	)

	r.metricsConsumer.RegisterHandler(
		producer.MessageMonitoring,
		func(ctx context.Context, queueName string, payload []byte) error {
			var event producer.Message
			if err := json.Unmarshal(payload, &event); err != nil {
				return err
			}
			return r.metricsUpdate.Handle(ctx, event)
		},
	)

	r.logger.Info("Consumers registered successfully")

	go func() {
		if err := r.uptimeConsumer.Start(ctx); err != nil {
			r.logger.Error("Failed to start uptime consumer", zap.Error(err))
		}
	}()

	go func() {
		if err := r.serverConsumer.Start(ctx); err != nil {
			r.logger.Error("Failed to start server consumer", zap.Error(err))
		}
	}()

	go func() {
		if err := r.metricsConsumer.Start(ctx); err != nil {
			r.logger.Error("Failed to start metrics consumer", zap.Error(err))
		}
	}()

	return nil
}
