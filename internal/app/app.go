// Package app configures and runs application.
package app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"github.com/AJackTi/go-kafka/config"
	http "github.com/AJackTi/go-kafka/internal/controller/http"
	"github.com/AJackTi/go-kafka/internal/domain"
	"github.com/AJackTi/go-kafka/pkg/es"
	"github.com/AJackTi/go-kafka/pkg/httpserver"
	kafkaClient "github.com/AJackTi/go-kafka/pkg/kafka"
	"github.com/AJackTi/go-kafka/pkg/logger"
)

// Run creates objects via constructors and supervises their lifecycle.
func Run(cfg *config.Config) (runErr error) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	defer cancel()
	log := logger.New(cfg.Log.Level)

	// Repository
	// db, err := mysql.New(cfg.MySQL.URL)
	// if err != nil {
	// 	logger.Fatal(fmt.Errorf("app - Run - mysql.New: %w", err))
	// }
	// defer db.Close()

	// Kafka producer
	kafkaProducer := kafkaClient.NewProducer(*log, cfg.Kafka.Brokers)
	defer func() {
		if closeErr := kafkaProducer.Close(); closeErr != nil && runErr == nil {
			runErr = fmt.Errorf("app - Run - kafka producer close: %w", closeErr)
		}
	}()

	// Kafka event serializer
	eventSerializer := domain.NewEventSerializer()

	eventBus := es.NewKafkaEventsBus(kafkaProducer, es.KafkaEventsBusConfig{
		Topic:             cfg.Events.Topic,
		TopicPrefix:       cfg.Events.TopicPrefix,
		Partitions:        cfg.Events.Partitions,
		ReplicationFactor: cfg.Events.ReplicationFactor,
	})

	// Connect kafka brokers
	kafkaConn, err := connectKafkaBrokers(ctx, cfg)
	if err != nil {
		return fmt.Errorf("app - Run - connectKafkaBrokers: %w", err)
	}
	defer func() {
		if closeErr := kafkaConn.Close(); closeErr != nil {
			log.Errorf("app - Run - kafka close: %v", closeErr)
		}
	}()

	// Init kafka topics
	if cfg.Kafka.InitTopics {
		initKafkaTopics(ctx, cfg, kafkaConn)
	}

	// HTTP Server
	handler := gin.New()

	// middleware for all
	// cors allow all origins
	if cfg.HTTP.Cors {
		log.Info("Set CORS for testing, please don't use it in production")
		handler.Use(cors.Default())
	}
	http.NewRouter(cfg, handler, log, eventSerializer, eventBus)
	httpServer := httpserver.New(handler, httpserver.Port(cfg.HTTP.Port))

	// Kafka consumer
	// subscription := subscription.NewSubscription(*logger, cfg, eventSerializer, db)
	// consumerGroup := kafkaClient.NewConsumerGroup(cfg.Kafka.Brokers, cfg.GroupID, *logger)
	// go func() {
	// 	err := consumerGroup.ConsumeTopicWithErrGroup(
	// 		ctx,
	// 		[]string{GetTopicName(cfg.Events.TopicPrefix, "Task")},
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
		log.Infof("app - Run - signal: %s", s.String())
	case err = <-httpServer.Notify():
		if err != nil {
			runErr = fmt.Errorf("app - Run - httpServer.Notify: %w", err)
		}
	}

	// Shutdown
	err = httpServer.Shutdown()
	if err != nil {
		if runErr == nil {
			runErr = fmt.Errorf("app - Run - httpServer.Shutdown: %w", err)
		} else {
			log.Error(fmt.Errorf("app - Run - httpServer.Shutdown: %w", err))
		}
	}

	return runErr
}
