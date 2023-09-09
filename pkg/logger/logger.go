package logger

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/evrone/go-clean-template/pkg/constants"
	"github.com/rs/zerolog"
)

// Interface -.
type Interface interface {
	Debug(message interface{}, args ...interface{})
	Info(message string, args ...interface{})
	Infof(template string, args ...interface{})
	Warn(message string, args ...interface{})
	Warnf(template string, args ...interface{})
	Error(message interface{}, args ...interface{})
	Fatal(message interface{}, args ...interface{})
	Errorf(template string, args ...interface{})
	KafkaProcessMessage(topic string, partition int, message []byte, workerID int, offset int64, time time.Time)
	KafkaLogCommittedMessage(topic string, partition int, offset int64)
}

// Logger -.
type Logger struct {
	logger *zerolog.Logger
}

var _ Interface = (*Logger)(nil)

// New -.
func New(level string) *Logger {
	var log zerolog.Level

	switch strings.ToLower(level) {
	case "error":
		log = zerolog.ErrorLevel
	case "warn":
		log = zerolog.WarnLevel
	case "info":
		log = zerolog.InfoLevel
	case "debug":
		log = zerolog.DebugLevel
	default:
		log = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(log)

	skipFrameCount := 3
	logger := zerolog.New(os.Stdout).With().Timestamp().CallerWithSkipFrameCount(zerolog.CallerSkipFrameCount + skipFrameCount).Logger()

	return &Logger{
		logger: &logger,
	}
}

// Debug -.
func (logger *Logger) Debug(message interface{}, args ...interface{}) {
	logger.msg("debug", message, args...)
}

// Info -.
func (logger *Logger) Info(message string, args ...interface{}) {
	logger.log(message, args...)
}

// Infof uses fmt.Sprintf to log a templated message.
func (logger *Logger) Infof(template string, args ...interface{}) {
	logger.log(template, args...)
}

// Warn -.
func (logger *Logger) Warn(message string, args ...interface{}) {
	logger.log(message, args...)
}

// Warnf uses fmt.Sprintf to log a templated message.
func (logger *Logger) Warnf(template string, args ...interface{}) {
	logger.log(template, args...)
}

// Error -.
func (logger *Logger) Error(message interface{}, args ...interface{}) {
	if logger.logger.GetLevel() == zerolog.DebugLevel {
		logger.Debug(message, args...)
	}

	logger.msg("error", message, args...)
}

// Errorf -.
func (logger *Logger) Errorf(template string, args ...interface{}) {
	if logger.logger.GetLevel() == zerolog.DebugLevel {
		logger.Debug(template, args...)
	}

	logger.msg("error", template, args...)
}

// Fatal -.
func (logger *Logger) Fatal(message interface{}, args ...interface{}) {
	logger.msg("fatal", message, args...)

	os.Exit(1)
}

func (logger *Logger) log(message string, args ...interface{}) {
	if len(args) == 0 {
		logger.logger.Info().Msg(message)
	} else {
		logger.logger.Info().Msgf(message, args...)
	}
}
func (logger *Logger) KafkaProcessMessage(topic string, partition int, message []byte, workerID int, offset int64, time time.Time) {
	logger.logger.Info().
		Str(constants.Topic, topic).
		Int(constants.Partition, partition).
		Int(constants.MessageSize, len(message)).
		Int(constants.WorkerID, workerID).
		Int64(constants.Offset, offset).
		Time(constants.Time, time).
		Msg("(Processing Kafka message)")
}

func (logger *Logger) KafkaLogCommittedMessage(topic string, partition int, offset int64) {
	logger.logger.Debug().Str(constants.Topic, topic).
		Int(constants.Partition, partition).
		Int64(constants.Offset, offset).
		Msg("(Committed Kafka message)")
}

func (logger *Logger) msg(level string, message interface{}, args ...interface{}) {
	switch msg := message.(type) {
	case error:
		logger.log(msg.Error(), args...)
	case string:
		logger.log(msg, args...)
	default:
		logger.log(fmt.Sprintf("%s message %v has unknown type %v", level, message, msg), args...)
	}
}
