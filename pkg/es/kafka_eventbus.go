package es

import (
	"context"
	"fmt"
	"time"

	"github.com/evrone/go-clean-template/pkg/es/serializer"
	kafkaClient "github.com/evrone/go-clean-template/pkg/kafka"
	"github.com/segmentio/kafka-go"
)

// KafkaEventsBusConfig kafka eventbus config.
type KafkaEventsBusConfig struct {
	Topic             string `mapstructure:"topic" validate:"required"`
	TopicPrefix       string `mapstructure:"topicPrefix" validate:"required"`
	Partitions        int    `mapstructure:"partitions" validate:"required,gte=0"`
	ReplicationFactor int    `mapstructure:"replicationFactor" validate:"required,gte=0"`
	Headers           []kafka.Header
}

type KafkaEventsBus struct {
	producer kafkaClient.Producer
	cfg      KafkaEventsBusConfig
}

// NewKafkaEventsBus kafkaEventsBus constructor.
func NewKafkaEventsBus(producer kafkaClient.Producer, cfg KafkaEventsBusConfig) *KafkaEventsBus {
	return &KafkaEventsBus{producer: producer, cfg: cfg}
}

// ProcessEvents serialize to json and publish es.Event's to the kafka topic.
func (e *KafkaEventsBus) ProcessEvents(ctx context.Context, events []Event) error {
	eventsBytes, err := serializer.Marshal(events)
	if err != nil {
		return err
	}

	topicName := GetTopicName(e.cfg.TopicPrefix, string(events[0].GetAggregateType()))

	return e.producer.PublishMessage(ctx, kafka.Message{
		Topic:   topicName,
		Value:   eventsBytes,
		Headers: nil,
		Time:    time.Now().UTC(),
	})
}

func GetTopicName(eventStorePrefix string, aggregateType string) string {
	return fmt.Sprintf("%s_%s", eventStorePrefix, aggregateType)
}

func GetKafkaAggregateTypeTopic(cfg KafkaEventsBusConfig, aggregateType string) kafka.TopicConfig {
	return kafka.TopicConfig{
		Topic:             GetTopicName(cfg.TopicPrefix, aggregateType),
		NumPartitions:     cfg.Partitions,
		ReplicationFactor: cfg.ReplicationFactor,
	}
}
