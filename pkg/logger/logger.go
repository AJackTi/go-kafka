package logger

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"

	"github.com/AJackTi/go-kafka/pkg/constants"
)

type Interface interface {
	Debug(message interface{})
	Info(message string)
	Infof(template string, args ...interface{})
	Warn(message string)
	Warnf(template string, args ...interface{})
	Error(message interface{})
	Fatal(message interface{})
	Errorf(template string, args ...interface{})
	KafkaProcessMessage(topic string, partition int, message []byte, workerID int, offset int64, time time.Time)
	KafkaLogCommittedMessage(topic string, partition int, offset int64)
}

type Logger struct {
	logger *zerolog.Logger
}

var _ Interface = (*Logger)(nil)

func New(level string) *Logger {
	return NewWithWriter(level, os.Stdout)
}

func NewWithWriter(level string, output io.Writer) *Logger {
	logLevel := parseLevel(level)
	configured := zerolog.New(output).
		Level(logLevel).
		With().
		Timestamp().
		CallerWithSkipFrameCount(zerolog.CallerSkipFrameCount + 3).
		Logger()

	return &Logger{logger: &configured}
}

func parseLevel(level string) zerolog.Level {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return zerolog.DebugLevel
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	default:
		return zerolog.InfoLevel
	}
}

func (logger *Logger) Debug(message interface{}) {
	logger.write(zerolog.DebugLevel, message)
}

func (logger *Logger) Info(message string) {
	logger.write(zerolog.InfoLevel, message)
}

func (logger *Logger) Infof(template string, args ...interface{}) {
	logger.writef(zerolog.InfoLevel, template, args...)
}

func (logger *Logger) Warn(message string) {
	logger.write(zerolog.WarnLevel, message)
}

func (logger *Logger) Warnf(template string, args ...interface{}) {
	logger.writef(zerolog.WarnLevel, template, args...)
}

func (logger *Logger) Error(message interface{}) {
	logger.write(zerolog.ErrorLevel, message)
}

func (logger *Logger) Errorf(template string, args ...interface{}) {
	logger.writef(zerolog.ErrorLevel, template, args...)
}

func (logger *Logger) Fatal(message interface{}) {
	logger.write(zerolog.FatalLevel, message)
	os.Exit(1)
}

func (logger *Logger) KafkaProcessMessage(
	topic string,
	partition int,
	message []byte,
	workerID int,
	offset int64,
	messageTime time.Time,
) {
	logger.logger.Info().
		Str(constants.Topic, topic).
		Int(constants.Partition, partition).
		Int(constants.MessageSize, len(message)).
		Int(constants.WorkerID, workerID).
		Int64(constants.Offset, offset).
		Time(constants.Time, messageTime).
		Msg("processing Kafka message")
}

func (logger *Logger) KafkaLogCommittedMessage(topic string, partition int, offset int64) {
	logger.logger.Debug().
		Str(constants.Topic, topic).
		Int(constants.Partition, partition).
		Int64(constants.Offset, offset).
		Msg("committed Kafka message")
}

func (logger *Logger) write(level zerolog.Level, message interface{}) {
	logger.logger.WithLevel(level).Msg(messageText(message))
}

func (logger *Logger) writef(level zerolog.Level, template string, args ...interface{}) {
	logger.logger.WithLevel(level).Msgf(template, args...)
}

func messageText(message interface{}) string {
	switch value := message.(type) {
	case error:
		return value.Error()
	case string:
		return value
	default:
		return fmt.Sprintf("%v", value)
	}
}
