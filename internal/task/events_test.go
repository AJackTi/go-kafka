package task_test

import (
	"errors"
	"testing"
	"time"

	"github.com/AJackTi/go-kafka/internal/eventstream"
	"github.com/AJackTi/go-kafka/internal/task"
)

func TestRehydrateRequiresContiguousEventsFromTheTaskStream(t *testing.T) {
	t.Parallel()

	created := mustEnvelope(t, "event-1", task.CreatedEventType, 1, task.Attributes{
		Title: "One", Name: "one", Status: "Doing",
	})
	gap := mustEnvelopeWithVersion(t, "event-3", task.UpdatedEventType, 3, task.Attributes{
		Title: "Three", Name: "three", Status: "Done",
	})

	_, err := task.Rehydrate(eventstream.StreamID{Type: task.AggregateType, ID: "task-1"}, []eventstream.Envelope{created, gap})
	if !errors.Is(err, task.ErrInvalidHistory) {
		t.Fatalf("Rehydrate() error = %v, want ErrInvalidHistory", err)
	}

	_, err = task.Rehydrate(eventstream.StreamID{Type: "Other", ID: "task-1"}, []eventstream.Envelope{created})
	if !errors.Is(err, task.ErrInvalidHistory) {
		t.Fatalf("wrong stream type error = %v, want ErrInvalidHistory", err)
	}
}

func TestDeletedEventIsAnImmutableTombstone(t *testing.T) {
	t.Parallel()

	created := mustEnvelope(t, "event-1", task.CreatedEventType, 1, task.Attributes{
		Title: "One", Name: "one", Status: "Doing",
	})
	deleted := mustEnvelopeWithVersion(t, "event-2", task.DeletedEventType, 2, struct{}{})
	state, err := task.Rehydrate(eventstream.StreamID{Type: task.AggregateType, ID: "task-1"}, []eventstream.Envelope{created, deleted})
	if err != nil {
		t.Fatalf("Rehydrate() error = %v", err)
	}
	if !state.Deleted || state.Version != 2 {
		t.Fatalf("state = %#v, want deleted version 2", state)
	}

	updated := mustEnvelopeWithVersion(t, "event-3", task.UpdatedEventType, 3, task.Attributes{
		Title: "Three", Name: "three", Status: "Done",
	})
	if err := state.Apply(updated); !errors.Is(err, task.ErrAggregateDeleted) {
		t.Fatalf("Apply(after delete) error = %v, want ErrAggregateDeleted", err)
	}
}

func TestDeletedPayloadMustBeAnEmptyObject(t *testing.T) {
	t.Parallel()

	created := mustEnvelope(t, "event-1", task.CreatedEventType, 1, task.Attributes{
		Title: "One", Name: "one", Status: "Doing",
	})
	badDelete, err := eventstream.NewEnvelope(
		"event-2",
		task.DeletedEventType,
		task.EventSchemaVersion,
		eventstream.AggregateRef{Type: task.AggregateType, ID: "task-1", Version: 2},
		time.Date(2026, time.August, 9, 1, 2, 3, 0, time.UTC),
		map[string]string{"id": "task-1"},
	)
	if err != nil {
		t.Fatalf("NewEnvelope() error = %v", err)
	}
	if _, err := task.Rehydrate(eventstream.StreamID{Type: task.AggregateType, ID: "task-1"}, []eventstream.Envelope{created, badDelete}); !errors.Is(err, task.ErrInvalidHistory) {
		t.Fatalf("Rehydrate() error = %v, want ErrInvalidHistory", err)
	}
}

func mustEnvelope(t *testing.T, id, eventType string, version uint64, payload any) eventstream.Envelope {
	t.Helper()
	return mustEnvelopeWithVersion(t, id, eventType, version, payload)
}

func mustEnvelopeWithVersion(t *testing.T, id, eventType string, version uint64, payload any) eventstream.Envelope {
	t.Helper()
	event, err := eventstream.NewEnvelope(
		id,
		eventType,
		task.EventSchemaVersion,
		eventstream.AggregateRef{Type: task.AggregateType, ID: "task-1", Version: version},
		time.Date(2026, time.August, 9, 1, 2, 3, 0, time.UTC),
		payload,
	)
	if err != nil {
		t.Fatalf("NewEnvelope() error = %v", err)
	}
	return event
}
