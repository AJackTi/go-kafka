package app

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/evrone/go-clean-template/config"
	"github.com/evrone/go-clean-template/pkg/constants"
	kafkaClient "github.com/evrone/go-clean-template/pkg/kafka"
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/segmentio/kafka-go"
)

const (
	TaskAggregateType string = "Task"
)

// KafkaEventsBusConfig kafka eventbus config.
type KafkaEventsBusConfig struct {
	Topic             string `mapstructure:"topic" validate:"required"`
	TopicPrefix       string `mapstructure:"topicPrefix" validate:"required"`
	Partitions        int    `mapstructure:"partitions" validate:"required,gte=0"`
	ReplicationFactor int    `mapstructure:"replicationFactor" validate:"required,gte=0"`
	Headers           []kafka.Header
}

func GetTopicName(eventStorePrefix string, aggregateType string) string {
	return fmt.Sprintf("%s_%s", eventStorePrefix, aggregateType)
}

func connectKafkaBrokers(ctx context.Context, cfg *config.Config) (*kafka.Conn, error) {
	kafkaConn, err := kafkaClient.NewKafkaConn(ctx, &kafkaClient.Config{
		Brokers:    cfg.Brokers,
		GroupID:    cfg.GroupID,
		InitTopics: cfg.InitTopics,
	})
	if err != nil {
		return nil, err
	}

	_, err = kafkaConn.Brokers()
	if err != nil {
		return nil, err
	}

	return kafkaConn, nil
}

func initKafkaTopics(ctx context.Context, cfg *config.Config, kafkaConn *kafka.Conn) {
	logger := logger.New(cfg.Log.Level)
	controller, err := kafkaConn.Controller()
	if err != nil {
		logger.Error("kafkaConn.Controller err: %v", err)
		return
	}

	controllerURI := net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port))
	logger.Infof("(kafka controller uri) controllerURI: %s", controllerURI)

	conn, err := kafka.DialContext(ctx, constants.TCP, controllerURI)
	if err != nil {
		logger.Errorf("initKafkaTopics.DialContext err: %v", err)
		return
	}
	defer conn.Close() // nolint: errcheck

	logger.Infof("(established new kafka controller connection) controllerURI: %s", controllerURI)

	bankAccountAggregateTopic := GetKafkaAggregateTypeTopic(&KafkaEventsBusConfig{
		Topic:             cfg.Topic,
		TopicPrefix:       cfg.TopicPrefix,
		Partitions:        cfg.Partitions,
		ReplicationFactor: cfg.ReplicationFactor,
	}, "Task")

	if err := conn.CreateTopics(bankAccountAggregateTopic); err != nil {
		logger.Warn("kafkaConn.CreateTopics", err)
		return
	}

	logger.Infof("(kafka topics created or already exists): %+v", []kafka.TopicConfig{bankAccountAggregateTopic})
}

func GetKafkaAggregateTypeTopic(cfg *KafkaEventsBusConfig, aggregateType string) kafka.TopicConfig {
	return kafka.TopicConfig{
		Topic:             GetTopicName(cfg.TopicPrefix, aggregateType),
		NumPartitions:     cfg.Partitions,
		ReplicationFactor: cfg.ReplicationFactor,
	}
}

func getConsumerGroupTopics(cfg *config.Config) []string {
	logger := logger.New(cfg.Log.Level)

	topics := []string{
		GetTopicName(cfg.KafkaPublisherConfig.TopicPrefix, string("Task")),
	}

	logger.Infof("(Consumer Topics) topics: %+v", topics)
	return topics
}
