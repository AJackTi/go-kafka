package subscription

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/evrone/go-clean-template/config"
	"github.com/evrone/go-clean-template/pkg/logger"
	"github.com/segmentio/kafka-go"
)

const (
	TaskAggregateType string = "Task"
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
	log logger.Logger
	cfg *config.Config
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
			s.handleBankAccountEvents(ctx, r, m)
		}
	}
}

func Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func (s *subscription) handleBankAccountEvents(ctx context.Context, r *kafka.Reader, m kafka.Message) {
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

// TODO: Handle subscription
func (s *subscription) handle(ctx context.Context, r *kafka.Reader, m kafka.Message, event Event) error {
	// err := s.projection.When(ctx, event)
	// if err != nil {
	// 	s.log.Errorf("MongoSubscription When err: %v", err)

	// 	recreateErr := s.recreateProjection(ctx, event)
	// 	if recreateErr != nil {
	// 		return tracing.TraceWithErr(span, errors.Wrapf(recreateErr, "recreateProjection err: %v", err))
	// 	}

	// 	s.commitErrMessage(ctx, r, m)
	// 	return tracing.TraceWithErr(span, errors.Wrapf(err, "When type: %s, aggregateID: %s", event.GetEventType(), event.GetAggregateID()))
	// }

	// s.log.Infof("MongoSubscription <<<commit>>> event: %s", event.String())
	return nil
}
