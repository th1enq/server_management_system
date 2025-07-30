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

type UptimeConsumer interface {
	RegisterHandler(queueName string, handlerFunc HandlerFunc)
	Start(ctx context.Context) error
}

type uptimeConsumer struct {
	saramaConsumer            sarama.ConsumerGroup
	logger                    *zap.Logger
	queueNameToHandlerFuncMap map[string]HandlerFunc
}

func NewUptimeConsumer(
	mqConfig configs.MQ,
	logger *zap.Logger,
) (UptimeConsumer, error) {
	saramaConsumer, err := sarama.NewConsumerGroup(mqConfig.Addresses, "uptime_consumer", newSaramaConfig(mqConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to create sarama consumer: %w", err)
	}

	return &uptimeConsumer{
		saramaConsumer:            saramaConsumer,
		logger:                    logger,
		queueNameToHandlerFuncMap: make(map[string]HandlerFunc),
	}, nil
}

func (c *uptimeConsumer) RegisterHandler(queueName string, handlerFunc HandlerFunc) {
	c.queueNameToHandlerFuncMap[queueName] = handlerFunc
}

func (c *uptimeConsumer) Start(ctx context.Context) error {
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
