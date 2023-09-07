// Package app configures and runs application.
package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/evrone/go-clean-template/config"
	v1 "github.com/evrone/go-clean-template/internal/controller/http/v1"
	"github.com/evrone/go-clean-template/internal/domain"
	"github.com/evrone/go-clean-template/internal/repo"
	"github.com/evrone/go-clean-template/internal/subscription"
	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/evrone/go-clean-template/pkg/es"
	"github.com/evrone/go-clean-template/pkg/httpserver"
	kafkaClient "github.com/evrone/go-clean-template/pkg/kafka"
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/evrone/go-clean-template/pkg/postgres"
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

	// kafka producer
	kafkaProducer := kafkaClient.NewProducer(*logger, cfg.Kafka.Brokers)
	defer kafkaProducer.Close()

	eventSerializer := domain.NewEventSerializer()
	eventBus := es.NewKafkaEventsBus(kafkaProducer, es.KafkaEventsBusConfig{
		Topic:             cfg.Topic,
		TopicPrefix:       cfg.TopicPrefix,
		Partitions:        cfg.Partitions,
		ReplicationFactor: cfg.ReplicationFactor,
	})

	// connect kafka brokers
	kafkaConn, err := connectKafkaBrokers(ctx, cfg)
	if err != nil {
		logger.Fatal(fmt.Errorf("app - Run - connectKafkaBrokers: %w", err))
	}
	defer kafkaConn.Close() // nolint: errcheck

	// init kafka topics
	if cfg.Kafka.InitTopics {
		initKafkaTopics(ctx, cfg, kafkaConn)
	}

	taskRepo := repo.NewTask(pg)
	// Use case
	taskUseCase := usecase.NewTask(eventSerializer, eventBus)

	// HTTP Server
	handler := gin.New()
	v1.NewRouter(handler, logger, taskUseCase)
	httpServer := httpserver.New(handler, httpserver.Port(cfg.HTTP.Port))

	subscription := subscription.NewSubscription(*logger, cfg, eventSerializer, taskRepo)
	consumerGroup := kafkaClient.NewConsumerGroup(cfg.Kafka.Brokers, cfg.GroupID, *logger)
	go func() {
		err := consumerGroup.ConsumeTopicWithErrGroup(
			ctx,
			getConsumerGroupTopics(cfg),
			10,
			subscription.ProcessMessagesErrGroup,
		)
		if err != nil {
			logger.Errorf("(consumerGroup ConsumeTopicWithErrGroup) err: %v", err)
			cancel()
			return
		}
	}()

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
