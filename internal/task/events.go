// Package task contains the task aggregate and its versioned event payloads.
package task

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/AJackTi/go-kafka/internal/eventstream"
)

const (
	AggregateType             = "Task"
	EventSchemaVersion uint16 = 1
	CreatedEventType          = "task.created"
	UpdatedEventType          = "task.updated"
	DeletedEventType          = "task.deleted"

	maxTitleLength       = 255
	maxNameLength        = 255
	maxImageLength       = 2048
	maxDescriptionLength = 10000
)

var (
	ErrInvalidAttributes = errors.New("invalid task attributes")
	ErrInvalidHistory    = errors.New("invalid task event history")
	ErrUnknownEventType  = errors.New("unknown task event type")
	ErrAggregateDeleted  = errors.New("task aggregate is deleted")
	ErrEmptyHistory      = errors.New("task event history is empty")
)

// Attributes are the mutable task fields carried by create and update events.
// Identity and version belong to the envelope, never to event data.
type Attributes struct {
	Title       string `json:"title"`
	Name        string `json:"name"`
	Image       string `json:"image"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

func (a Attributes) Validate() error {
	switch {
	case strings.TrimSpace(a.Title) == "":
		return fmt.Errorf("%w: title is required", ErrInvalidAttributes)
	case len(a.Title) > maxTitleLength:
		return fmt.Errorf("%w: title exceeds %d characters", ErrInvalidAttributes, maxTitleLength)
	case strings.TrimSpace(a.Name) == "":
		return fmt.Errorf("%w: name is required", ErrInvalidAttributes)
	case len(a.Name) > maxNameLength:
		return fmt.Errorf("%w: name exceeds %d characters", ErrInvalidAttributes, maxNameLength)
	case len(a.Image) > maxImageLength:
		return fmt.Errorf("%w: image exceeds %d characters", ErrInvalidAttributes, maxImageLength)
	case len(a.Description) > maxDescriptionLength:
		return fmt.Errorf("%w: description exceeds %d characters", ErrInvalidAttributes, maxDescriptionLength)
	case a.Status != "Doing" && a.Status != "Done":
		return fmt.Errorf("%w: status must be Doing or Done", ErrInvalidAttributes)
	default:
		return nil
	}
}

// State is the rehydrated task aggregate. Deleted tasks remain addressable so
// their version can never be reset or accidentally reused.
type State struct {
	ID      string `json:"id"`
	Version uint64 `json:"version"`
	Attributes
	Deleted   bool      `json:"deleted"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Rehydrate applies a complete, contiguous stream to a new aggregate state.
func Rehydrate(stream eventstream.StreamID, history []eventstream.Envelope) (*State, error) {
	if stream.Type != AggregateType || strings.TrimSpace(stream.ID) == "" {
		return nil, fmt.Errorf("%w: stream type and id are required", ErrInvalidHistory)
	}
	if len(history) == 0 {
		return nil, ErrEmptyHistory
	}

	state := &State{ID: stream.ID}
	for _, event := range history {
		if err := state.Apply(event); err != nil {
			return nil, err
		}
	}
	if state.ID != stream.ID {
		return nil, fmt.Errorf("%w: stream id changed", ErrInvalidHistory)
	}
	return state, nil
}

// Apply validates stream sequencing and applies one event to the aggregate.
func (s *State) Apply(event eventstream.Envelope) error {
	if s == nil {
		return fmt.Errorf("%w: nil aggregate", ErrInvalidHistory)
	}
	if err := event.Validate(); err != nil {
		return err
	}
	if event.Aggregate.Type != AggregateType || event.Aggregate.ID != s.ID {
		return fmt.Errorf("%w: event belongs to %s/%s", ErrInvalidHistory, event.Aggregate.Type, event.Aggregate.ID)
	}
	if event.Aggregate.Version != s.Version+1 {
		return fmt.Errorf("%w: expected version %d, got %d", ErrInvalidHistory, s.Version+1, event.Aggregate.Version)
	}
	if event.SchemaVersion != EventSchemaVersion {
		return fmt.Errorf("%w: unsupported schema version %d", ErrInvalidHistory, event.SchemaVersion)
	}

	switch event.Type {
	case CreatedEventType:
		if s.Version != 0 {
			return fmt.Errorf("%w: task can only be created once", ErrInvalidHistory)
		}
		var attributes Attributes
		if err := decodePayload(event.Data, &attributes); err != nil {
			return err
		}
		if err := attributes.Validate(); err != nil {
			return err
		}
		s.Attributes = attributes
		s.CreatedAt = event.OccurredAt.UTC()
		s.UpdatedAt = event.OccurredAt.UTC()
	case UpdatedEventType:
		if s.Version == 0 {
			return fmt.Errorf("%w: update before create", ErrInvalidHistory)
		}
		if s.Deleted {
			return ErrAggregateDeleted
		}
		var attributes Attributes
		if err := decodePayload(event.Data, &attributes); err != nil {
			return err
		}
		if err := attributes.Validate(); err != nil {
			return err
		}
		s.Attributes = attributes
		s.UpdatedAt = event.OccurredAt.UTC()
	case DeletedEventType:
		if s.Version == 0 {
			return fmt.Errorf("%w: delete before create", ErrInvalidHistory)
		}
		if s.Deleted {
			return ErrAggregateDeleted
		}
		var payload struct{}
		if err := decodePayload(event.Data, &payload); err != nil {
			return err
		}
		s.Deleted = true
		s.UpdatedAt = event.OccurredAt.UTC()
	default:
		return fmt.Errorf("%w: %s", ErrUnknownEventType, event.Type)
	}

	s.Version = event.Aggregate.Version
	return nil
}

func decodePayload(data json.RawMessage, target any) error {
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		return fmt.Errorf("%w: payload must be an object", ErrInvalidHistory)
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return fmt.Errorf("%w: decode payload: %w", ErrInvalidHistory, err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return fmt.Errorf("%w: payload contains multiple JSON values", ErrInvalidHistory)
		}
		return fmt.Errorf("%w: trailing payload data: %w", ErrInvalidHistory, err)
	}
	return nil
}
