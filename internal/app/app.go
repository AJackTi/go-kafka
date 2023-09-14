// Package app configures and runs application.
package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/AJackTi/go-kafka/config"
	http "github.com/AJackTi/go-kafka/internal/controller/http"
	"github.com/AJackTi/go-kafka/internal/domain"
	"github.com/AJackTi/go-kafka/pkg/es"
	"github.com/AJackTi/go-kafka/pkg/httpserver"
	kafkaClient "github.com/AJackTi/go-kafka/pkg/kafka"
	"github.com/AJackTi/go-kafka/pkg/logger"
	"github.com/AJackTi/go-kafka/pkg/postgres"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Run creates objects via constructors.
func Run(cfg *config.Config) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()
	logger := logger.New(cfg.Log.Level)

	// Repository
	pg, err := postgres.New(cfg.PG.URL, postgres.MaxPoolSize(cfg.PG.PoolMax))
	if err != nil {
		logger.Fatal(fmt.Errorf("app - Run - postgres.New: %w", err))
	}
	defer pg.Close()

	// Kafka producer
	kafkaProducer := kafkaClient.NewProducer(*logger, cfg.Kafka.Brokers)
	defer kafkaProducer.Close()

	// Kafka event serializer
	eventSerializer := domain.NewEventSerializer()

	eventBus := es.NewKafkaEventsBus(kafkaProducer, es.KafkaEventsBusConfig{
		Topic:             cfg.Topic,
		TopicPrefix:       cfg.TopicPrefix,
		Partitions:        cfg.Partitions,
		ReplicationFactor: cfg.ReplicationFactor,
	})

	// Connect kafka brokers
	kafkaConn, err := connectKafkaBrokers(ctx, cfg)
	if err != nil {
		logger.Fatal(fmt.Errorf("app - Run - connectKafkaBrokers: %w", err))
	}
	defer kafkaConn.Close() // nolint: errcheck

	// Init kafka topics
	if cfg.Kafka.InitTopics {
		initKafkaTopics(ctx, cfg, kafkaConn)
	}

	// HTTP Server
	handler := gin.New()

	// middleware for all
	// cors allow all origins
	if *cfg.HTTP.Cors {
		logger.Info("Set CORS for testing, please don't use it in production")
		handler.Use(cors.Default())
	}
	http.NewRouter(cfg, handler, logger, pg, eventSerializer, eventBus)
	httpServer := httpserver.New(handler, httpserver.Port(cfg.HTTP.Port))

	// Kafka consumer
	// subscription := subscription.NewSubscription(*logger, cfg, eventSerializer, pg)
	// consumerGroup := kafkaClient.NewConsumerGroup(cfg.Kafka.Brokers, cfg.GroupID, *logger)
	// go func() {
	// 	err := consumerGroup.ConsumeTopicWithErrGroup(
	// 		ctx,
	// 		getConsumerGroupTopics(cfg),
	// 		10,
	// 		subscription.ProcessMessagesErrGroup,
	// 	)
	// 	if err != nil {
	// 		logger.Errorf("(consumerGroup ConsumeTopicWithErrGroup) err: %v", err)
	// 		cancel()
	// 		return
	// 	}
	// }()

	// Waiting signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		logger.Info("app - Run - signal: " + s.String())
	case err = <-httpServer.Notify():
		logger.Error(fmt.Errorf("app - Run - httpServer.Notify: %w", err))
	}

	// Shutdown
	err = httpServer.Shutdown()
	if err != nil {
		logger.Error(fmt.Errorf("app - Run - httpServer.Shutdown: %w", err))
	}
}
