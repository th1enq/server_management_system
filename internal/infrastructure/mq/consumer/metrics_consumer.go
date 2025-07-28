package consumer

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/IBM/sarama"
	"github.com/th1enq/server_management_system/internal/configs"
	"github.com/th1enq/server_management_system/internal/utils"
	"go.uber.org/zap"
)

type MetricsConsumer interface {
	RegisterHandler(queueName string, handlerFunc HandlerFunc)
	Start(ctx context.Context) error
}

type metricsConsumer struct {
	saramaConsumer            sarama.ConsumerGroup
	logger                    *zap.Logger
	queueNameToHandlerFuncMap map[string]HandlerFunc
}

func NewMetricsConsumer(
	mqConfig configs.MQ,
	logger *zap.Logger,
) (MetricsConsumer, error) {
	saramaConsumer, err := sarama.NewConsumerGroup(mqConfig.Addresses, "metrics_consumer", newSaramaConfig(mqConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to create sarama consumer: %w", err)
	}

	return &metricsConsumer{
		saramaConsumer:            saramaConsumer,
		logger:                    logger,
		queueNameToHandlerFuncMap: make(map[string]HandlerFunc),
	}, nil
}

func (c *metricsConsumer) RegisterHandler(queueName string, handlerFunc HandlerFunc) {
	c.queueNameToHandlerFuncMap[queueName] = handlerFunc
}

func (c *metricsConsumer) Start(ctx context.Context) error {
	logger := utils.LoggerWithContext(ctx, c.logger)

	exitSignalChannel := make(chan os.Signal, 1)
	signal.Notify(exitSignalChannel, os.Interrupt)

	for queueName, handlerFunc := range c.queueNameToHandlerFuncMap {
		go func(queueName string, handlerFunc HandlerFunc) {
			if err := c.saramaConsumer.Consume(
				context.Background(),
				[]string{queueName},
				newConsumerHandler(handlerFunc, exitSignalChannel),
			); err != nil {
				logger.
					With(zap.String("queue_name", queueName)).
					With(zap.Error(err)).
					Error("failed to consume message from queue")
			}
		}(queueName, handlerFunc)
	}

	<-exitSignalChannel
	return nil
}
