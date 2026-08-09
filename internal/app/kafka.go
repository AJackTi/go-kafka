package app

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/segmentio/kafka-go"

	"github.com/AJackTi/go-kafka/config"
	"github.com/AJackTi/go-kafka/pkg/constants"
	kafkaClient "github.com/AJackTi/go-kafka/pkg/kafka"
	"github.com/AJackTi/go-kafka/pkg/logger"
)

// KafkaEventsBusConfig kafka eventbus config.
type KafkaEventsBusConfig struct {
	Topic             string `mapstructure:"topic" validate:"required"`
	TopicPrefix       string `mapstructure:"topicPrefix" validate:"required"`
	Partitions        int    `mapstructure:"partitions" validate:"required,gte=0"`
	ReplicationFactor int    `mapstructure:"replicationFactor" validate:"required,gte=0"`
	Headers           []kafka.Header
}

func GetTopicName(eventStorePrefix, aggregateType string) string {
	return fmt.Sprintf("%s_%s", eventStorePrefix, aggregateType)
}

func connectKafkaBrokers(ctx context.Context, cfg *config.Config) (*kafka.Conn, error) {
	kafkaConn, err := kafkaClient.NewKafkaConn(ctx, &kafkaClient.Config{
		Brokers:    cfg.Kafka.Brokers,
		GroupID:    cfg.Kafka.GroupID,
		InitTopics: cfg.Kafka.InitTopics,
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
	log := logger.New(cfg.Log.Level)
	controller, err := kafkaConn.Controller()
	if err != nil {
		log.Errorf("kafkaConn.Controller err: %v", err)
		return
	}

	controllerURI := net.JoinHostPort(controller.Host, strconv.Itoa(controller.Port))
	log.Infof("(kafka controller uri) controllerURI: %s", controllerURI)

	conn, err := kafka.DialContext(ctx, constants.TCP, controllerURI)
	if err != nil {
		log.Errorf("initKafkaTopics.DialContext err: %v", err)
		return
	}
	defer func() {
		if closeErr := conn.Close(); closeErr != nil {
			log.Errorf("initKafkaTopics.Close: %v", closeErr)
		}
	}()

	log.Infof("(established new kafka controller connection) controllerURI: %s", controllerURI)

	taskAggregateTopic := GetKafkaAggregateTypeTopic(&KafkaEventsBusConfig{
		Topic:             cfg.Events.Topic,
		TopicPrefix:       cfg.Events.TopicPrefix,
		Partitions:        cfg.Events.Partitions,
		ReplicationFactor: cfg.Events.ReplicationFactor,
	}, "Task")

	if err := conn.CreateTopics(taskAggregateTopic); err != nil {
		log.Warnf("kafkaConn.CreateTopics: %v", err)
		return
	}

	log.Infof("(kafka topics created or already exists): %+v", []kafka.TopicConfig{taskAggregateTopic})
}

func GetKafkaAggregateTypeTopic(cfg *KafkaEventsBusConfig, aggregateType string) kafka.TopicConfig {
	return kafka.TopicConfig{
		Topic:             GetTopicName(cfg.TopicPrefix, aggregateType),
		NumPartitions:     cfg.Partitions,
		ReplicationFactor: cfg.ReplicationFactor,
	}
}
