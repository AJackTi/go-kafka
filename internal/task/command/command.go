// Package command contains task command handlers and the event-store port
// they depend on. It deliberately has no Kafka or SQL dependency.
package command

import (
	"context"

	"github.com/AJackTi/go-kafka/internal/eventstream"
	"github.com/AJackTi/go-kafka/internal/task"
)

// EventStore is the atomic command boundary. Implementations must append the
// event and its outbox record in one transaction and enforce expectedVersion
// against the current stream version.
type EventStore interface {
	Load(context.Context, eventstream.StreamID) ([]eventstream.Envelope, error)
	Append(context.Context, eventstream.StreamID, uint64, ...eventstream.Envelope) error
}

type CreateTask struct {
	ID          string
	Title       string
	Name        string
	Image       string
	Description string
	Status      string
}

type UpdateTask struct {
	ID              string
	ExpectedVersion uint64
	Title           string
	Name            string
	Image           string
	Description     string
	Status          string
}

type DeleteTask struct {
	ID              string
	ExpectedVersion uint64
}

// Receipt identifies the event accepted by the command boundary.
type Receipt struct {
	AggregateID string               `json:"aggregate_id"`
	EventID     string               `json:"event_id"`
	Version     uint64               `json:"version"`
	Type        string               `json:"type"`
	Event       eventstream.Envelope `json:"-"`
}

func (c CreateTask) attributes() task.Attributes {
	return task.Attributes{
		Title:       c.Title,
		Name:        c.Name,
		Image:       c.Image,
		Description: c.Description,
		Status:      c.Status,
	}
}

func (c UpdateTask) attributes() task.Attributes {
	return task.Attributes{
		Title:       c.Title,
		Name:        c.Name,
		Image:       c.Image,
		Description: c.Description,
		Status:      c.Status,
	}
}
