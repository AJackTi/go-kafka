// Package app configures and runs application.
package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"

	"github.com/evrone/go-clean-template/config"
	v1 "github.com/evrone/go-clean-template/internal/controller/http/v1"
	"github.com/evrone/go-clean-template/internal/repo"
	"github.com/evrone/go-clean-template/internal/usecase"
	"github.com/evrone/go-clean-template/pkg/httpserver"
	kafkaClient "github.com/evrone/go-clean-template/pkg/kafka"
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/evrone/go-clean-template/pkg/postgres"
)

// Run creates objects via constructors.
func Run(cfg *config.Config) {
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

	// connect kafka brokers
	kafkaConn, err := connectKafkaBrokers(context.Background(), cfg)
	if err != nil {
		logger.Fatal(fmt.Errorf("app - Run - connectKafkaBrokers: %w", err))
	}
	defer kafkaConn.Close() // nolint: errcheck

	// init kafka topics
	if cfg.Kafka.InitTopics {
		initKafkaTopics(context.Background(), cfg, kafkaConn)
	}

	// Use case
	taskUseCase := usecase.NewTask(
		repo.NewTask(pg),
	)

	// HTTP Server
	handler := gin.New()
	v1.NewRouter(handler, logger, taskUseCase)
	httpServer := httpserver.New(handler, httpserver.Port(cfg.HTTP.Port))

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

	// consumerGroup := kafkaClient.NewConsumerGroup(cfg.Kafka.Brokers, cfg.GroupID, *logger)
	go func() {
		// err := consumerGroup.ConsumeTopicWithErrGroup(
		// 	context.Background(),
		// 	getConsumerGroupTopics(cfg),
		// 	10,
		// 	mongoSubscription.ProcessMessagesErrGroup,
		// )
		// if err != nil {
		// 	a.log.Errorf("(mongoConsumerGroup ConsumeTopicWithErrGroup) err: %v", err)
		// 	cancel()
		// 	return
		// }
	}()
}
