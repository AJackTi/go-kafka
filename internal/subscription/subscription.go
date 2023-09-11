package subscription

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/AJackTi/go-kafka/config"
	"github.com/AJackTi/go-kafka/internal/domain"
	"github.com/AJackTi/go-kafka/internal/entity"
	"github.com/AJackTi/go-kafka/internal/events"
	"github.com/AJackTi/go-kafka/internal/repo"
	"github.com/AJackTi/go-kafka/pkg/es"
	"github.com/AJackTi/go-kafka/pkg/logger"
	"github.com/AJackTi/go-kafka/pkg/postgres"
	"github.com/segmentio/kafka-go"
)

const (
	TaskAggregateType string = "Task"
)

var (
	ErrUnknownEventType = errors.New("unknown event type")
)

// AggregateType type of the Aggregate
type AggregateType string

// EventType is the type of any event, used as its unique identifier.
type EventType string

// Event is an internal representation of an event, returned when the Aggregate
// uses NewEvent to create a new event. The events loaded from the db is
// represented by each DBs internal event type, implementing Event.
type Event struct {
	EventID       string
	AggregateID   string
	EventType     EventType
	AggregateType AggregateType
	Version       uint64
	Data          []byte
	Metadata      []byte
	Timestamp     time.Time
}

type subscription struct {
	log             logger.Logger
	cfg             *config.Config
	eventSerializer *domain.EventSerializer
	pg              *postgres.Postgres
	taskRepo        *repo.TaskRepo
}

func NewSubscription(
	log logger.Logger,
	cfg *config.Config,
	eventSerializer *domain.EventSerializer,
	pg *postgres.Postgres,
) *subscription {
	// Repo
	taskRepo := repo.NewTask(pg)
	return &subscription{
		log:             log,
		cfg:             cfg,
		eventSerializer: eventSerializer,
		pg:              pg,
		taskRepo:        taskRepo,
	}
}

func GetTopicName(eventStorePrefix string, aggregateType string) string {
	return fmt.Sprintf("%s_%s", eventStorePrefix, aggregateType)
}

func (s *subscription) ProcessMessagesErrGroup(ctx context.Context, r *kafka.Reader, workerID int) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		m, err := r.FetchMessage(ctx)
		if err != nil {
			s.log.Warnf("(mongoSubscription) workerID: %d, err: %v", workerID, err)
			continue
		}

		s.logProcessMessage(m, workerID)

		switch m.Topic {
		case GetTopicName(s.cfg.KafkaPublisherConfig.TopicPrefix, TaskAggregateType):
			s.handleTaskEvents(ctx, r, m)
		}
	}
}

func Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func (s *subscription) handleTaskEvents(ctx context.Context, r *kafka.Reader, m kafka.Message) {
	var events []Event
	if err := Unmarshal(m.Value, &events); err != nil {
		s.commitErrMessage(ctx, r, m)
		return
	}

	for _, event := range events {
		if err := s.handle(ctx, r, m, event); err != nil {
			return
		}
	}
	s.commitMessage(ctx, r, m)
}

func (s *subscription) handle(ctx context.Context, r *kafka.Reader, m kafka.Message, event Event) error {
	if err := s.when(ctx, es.Event{
		EventID:       event.EventID,
		AggregateID:   event.AggregateID,
		EventType:     es.EventType(event.EventType),
		AggregateType: es.AggregateType(event.AggregateType),
		Version:       event.Version,
		Data:          event.Data,
		Metadata:      event.Metadata,
		Timestamp:     event.Timestamp,
	}); err != nil {
		return err
	}

	return nil
}

func (s *subscription) when(ctx context.Context, esEvent es.Event) error {
	var (
		err  error
		task *entity.Task
	)

	deserializedEvent, err := s.eventSerializer.DeserializeEvent(esEvent)
	if err != nil {
		return err
	}

	switch deserializedEvent.(type) {
	case *events.TaskCreatedEventV1:
		task, err = unmarshalToTask(string(esEvent.Data))
		if err == nil {
			return s.taskRepo.CreateTask(ctx, task)
		}
	case *events.TaskUpdatedEventV1:
		task, err = unmarshalToTask(string(esEvent.Data))
		if err == nil {
			return s.taskRepo.UpdateTask(ctx, task)
		}

	default:
		return ErrUnknownEventType
	}

	return err
}

func unmarshalToTask(jsonStr string) (*entity.Task, error) {
	var response entity.Task
	err := json.Unmarshal([]byte(jsonStr), &response)
	if err != nil {
		return nil, err
	}

	return &response, nil
}
