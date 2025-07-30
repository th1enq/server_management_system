package producer

import (
	"context"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/th1enq/server_management_system/internal/configs"
	"github.com/th1enq/server_management_system/internal/domain/entity"
	"github.com/th1enq/server_management_system/internal/utils"
	"go.uber.org/zap"
)

type Client interface {
	Produce(ctx context.Context, queueName string, payload []byte) error
}

type client struct {
	saramaSyncProducer sarama.SyncProducer
	logger             *zap.Logger
}

func newSaramaConfig(mqConfig configs.MQ) *sarama.Config {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.Retry.Max = 3
	saramaConfig.Producer.RequiredAcks = sarama.WaitForAll
	saramaConfig.Producer.Return.Successes = true
	saramaConfig.ClientID = mqConfig.ClientID
	saramaConfig.Metadata.Full = true
	return saramaConfig
}

type Message struct {
	ServerID  string              `json:"server_id" binding:"required"`
	OldStatus entity.ServerStatus `json:"old_status" binding:"required"`
	NewStatus entity.ServerStatus `json:"new_status" binding:"required"`
	Timestamp time.Time           `json:"timestamp" binding:"required"`
	CPU       int                 `json:"cpu" binding:"required,gte=0"`
	RAM       int                 `json:"ram" binding:"required,gte=0"`
	Disk      int                 `json:"disk" binding:"required,gte=0"`
}

type StatusChangeMessage struct {
	ServerID  string              `json:"server_id" binding:"required"`
	OldStatus entity.ServerStatus `json:"old_status" binding:"required"`
	NewStatus entity.ServerStatus `json:"new_status" binding:"required"`
	Timestamp time.Time           `json:"timestamp" binding:"required"`
	Interval  int64               `json:"interval" binding:"required,gte=0"`
}

func NewClient(
	mqConfig configs.MQ,
	logger *zap.Logger,
) (Client, error) {
	saramaSyncProducer, err := sarama.NewSyncProducer(mqConfig.Addresses, newSaramaConfig(mqConfig))
	if err != nil {
		return nil, fmt.Errorf("failed to create sarama sync producer: %w", err)
	}

	return &client{
		saramaSyncProducer: saramaSyncProducer,
		logger:             logger,
	}, nil
}

func (c client) Produce(ctx context.Context, queueName string, payload []byte) error {
	logger := utils.LoggerWithContext(ctx, c.logger).
		With(zap.String("queue_name", queueName)).
		With(zap.ByteString("payload", payload))

	if _, _, err := c.saramaSyncProducer.SendMessage(&sarama.ProducerMessage{
		Topic: queueName,
		Value: sarama.ByteEncoder(payload),
	}); err != nil {
		logger.With(zap.Error(err)).Error("failed to produce message")
		return fmt.Errorf("failed to produce message: %w", err)
	}

	return nil
}
