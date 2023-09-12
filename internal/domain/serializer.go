package domain

import (
	"errors"

	"github.com/AJackTi/go-kafka/internal/events"
	"github.com/AJackTi/go-kafka/pkg/es"
	"github.com/AJackTi/go-kafka/pkg/es/serializer"
)

var (
	ErrInvalidEvent = errors.New("invalid event")
)

type EventSerializer struct {
}

func NewEventSerializer() *EventSerializer {
	return &EventSerializer{}
}

func (s *EventSerializer) SerializeEvent(aggregate es.Aggregate, event any) (es.Event, error) {
	eventBytes, err := serializer.Marshal(event)
	if err != nil {
		return es.Event{}, err
	}

	switch evt := event.(type) {
	case *events.TaskCreatedEventV1:
		return es.NewEvent(aggregate, events.TaskCreatedEventType, eventBytes, evt.Metadata), nil
	case *events.TaskUpdatedEventV1:
		return es.NewEvent(aggregate, events.TaskUpdatedEventType, eventBytes, evt.Metadata), nil
	case *events.TaskDeletedEventV1:
		return es.NewEvent(aggregate, events.TaskDeletedEventType, eventBytes, evt.Metadata), nil

	default:
		return es.Event{}, err
	}

}

func (s *EventSerializer) DeserializeEvent(event es.Event) (any, error) {
	switch event.GetEventType() {
	case events.TaskCreatedEventType:
		return deserializeEvent(event, new(events.TaskCreatedEventV1))
	case events.TaskUpdatedEventType:
		return deserializeEvent(event, new(events.TaskUpdatedEventV1))
	case events.TaskDeletedEventType:
		return deserializeEvent(event, new(events.TaskDeletedEventV1))

	default:
		return nil, ErrInvalidEvent
	}
}

func deserializeEvent(event es.Event, targetEvent any) (any, error) {
	if err := event.GetJsonData(&targetEvent); err != nil {
		return nil, err
	}
	return targetEvent, nil
}
