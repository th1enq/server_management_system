package mq

import (
	"github.com/IBM/sarama"
	"github.com/th1enq/server_management_system/internal/configs"
	"github.com/th1enq/server_management_system/internal/infrastructure/models"
)

type MessageBroker interface {
	Send(message models.Message) error
}

type messageBroker struct {
	producer sarama.SyncProducer
}

func NewBroker(cfg configs.Broker) (MessageBroker, error) {
	saramaConfig := sarama.NewConfig()
	saramaConfig.Producer.Return.Successes = true
	producer, err := sarama.NewSyncProducer(cfg.Addresses, saramaConfig)
	if err != nil {
		return nil, err
	}
	return &messageBroker{
		producer: producer,
	}, nil
}

func (d messageBroker) Send(event models.Message) error {
	var headers []sarama.RecordHeader

	for k, v := range event.Headers {
		headers = append(headers, sarama.RecordHeader{
			Key:   sarama.ByteEncoder(k),
			Value: sarama.ByteEncoder(v),
		})
	}

	msg := &sarama.ProducerMessage{
		Topic:   event.Topic,
		Key:     sarama.StringEncoder(event.Key),
		Value:   sarama.ByteEncoder(event.Body),
		Headers: headers,
	}

	_, _, err := d.producer.SendMessage(msg)

	return err
}
