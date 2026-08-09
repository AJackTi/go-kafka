package command_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/AJackTi/go-kafka/internal/eventstream"
	"github.com/AJackTi/go-kafka/internal/task"
	"github.com/AJackTi/go-kafka/internal/task/command"
)

func TestHandlerCreateAppendsOneVersionedEvent(t *testing.T) {
	t.Parallel()

	store := newMemoryStore()
	handler := newTestHandler(t, store, "task-123", "event-123")

	receipt, err := handler.Create(context.Background(), command.CreateTask{
		Title:  "Ship it",
		Name:   "release",
		Status: "Doing",
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	events := store.events[eventstream.StreamID{Type: task.AggregateType, ID: "task-123"}]
	if len(events) != 1 {
		t.Fatalf("appended %d events, want 1", len(events))
	}
	event := events[0]
	if event.ID != "event-123" || event.Type != task.CreatedEventType || event.Aggregate.Version != 1 {
		t.Fatalf("unexpected event: %#v", event)
	}
	if receipt.AggregateID != "task-123" || receipt.EventID != "event-123" || receipt.Version != 1 {
		t.Fatalf("unexpected receipt: %#v", receipt)
	}

	var payload task.Attributes
	if err := json.Unmarshal(event.Data, &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if payload.Title != "Ship it" || payload.Name != "release" || payload.Status != "Doing" {
		t.Fatalf("unexpected payload: %#v", payload)
	}
	var raw map[string]any
	if err := json.Unmarshal(event.Data, &raw); err != nil {
		t.Fatalf("decode raw payload: %v", err)
	}
	if _, exists := raw["id"]; exists {
		t.Fatal("payload must not duplicate aggregate id")
	}
}

func TestHandlerUpdateAndDeleteAdvanceTheSameStream(t *testing.T) {
	t.Parallel()

	store := newMemoryStore()
	handler := newTestHandler(t, store, "task-123", "event-1", "event-2", "event-3")
	created, err := handler.Create(context.Background(), command.CreateTask{Title: "One", Name: "one", Status: "Doing"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	updated, err := handler.Update(context.Background(), command.UpdateTask{
		ID:              created.AggregateID,
		ExpectedVersion: 1,
		Title:           "Two",
		Name:            "two",
		Status:          "Done",
	})
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}
	deleted, err := handler.Delete(context.Background(), command.DeleteTask{ID: created.AggregateID, ExpectedVersion: 2})
	if err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	events := store.events[eventstream.StreamID{Type: task.AggregateType, ID: created.AggregateID}]
	if len(events) != 3 {
		t.Fatalf("appended %d events, want 3", len(events))
	}
	if events[1].Aggregate.ID != created.AggregateID || events[1].Aggregate.Version != 2 || updated.Version != 2 {
		t.Fatalf("update moved to a different stream: %#v / %#v", events[1], updated)
	}
	if events[2].Aggregate.ID != created.AggregateID || events[2].Aggregate.Version != 3 || deleted.Version != 3 {
		t.Fatalf("delete moved to a different stream: %#v / %#v", events[2], deleted)
	}
}

func TestHandlerRejectsStaleAndDeletedCommandsWithoutAppending(t *testing.T) {
	t.Parallel()

	store := newMemoryStore()
	handler := newTestHandler(t, store, "task-123", "event-1", "event-2", "event-3")
	created, err := handler.Create(context.Background(), command.CreateTask{Title: "One", Name: "one", Status: "Doing"})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}
	if _, updateErr := handler.Update(context.Background(), command.UpdateTask{
		ID:              created.AggregateID,
		ExpectedVersion: 1,
		Title:           "Two",
		Name:            "two",
		Status:          "Doing",
	}); updateErr != nil {
		t.Fatalf("first Update() error = %v", updateErr)
	}
	beforeStale := store.appendCalls
	_, err = handler.Update(context.Background(), command.UpdateTask{
		ID:              created.AggregateID,
		ExpectedVersion: 1,
		Title:           "Stale",
		Name:            "stale",
		Status:          "Doing",
	})
	var conflict *command.VersionConflictError
	if !errors.As(err, &conflict) || conflict.Expected != 1 || conflict.Actual != 2 {
		t.Fatalf("stale Update() error = %v, want version conflict 1/2", err)
	}
	if store.appendCalls != beforeStale {
		t.Fatal("stale command must not append")
	}

	if _, deleteErr := handler.Delete(context.Background(), command.DeleteTask{ID: created.AggregateID, ExpectedVersion: 2}); deleteErr != nil {
		t.Fatalf("Delete() error = %v", deleteErr)
	}
	beforeDeleted := store.appendCalls
	_, err = handler.Update(context.Background(), command.UpdateTask{
		ID:              created.AggregateID,
		ExpectedVersion: 3,
		Title:           "After delete",
		Name:            "deleted",
		Status:          "Doing",
	})
	if !errors.Is(err, command.ErrAggregateDeleted) {
		t.Fatalf("update after delete error = %v, want ErrAggregateDeleted", err)
	}
	if store.appendCalls != beforeDeleted {
		t.Fatal("command after delete must not append")
	}
}

func TestHandlerPropagatesAppendFailure(t *testing.T) {
	t.Parallel()

	store := newMemoryStore()
	store.appendErr = errors.New("database unavailable")
	handler := newTestHandler(t, store, "task-123", "event-123")

	_, err := handler.Create(context.Background(), command.CreateTask{Title: "One", Name: "one", Status: "Doing"})
	if !errors.Is(err, store.appendErr) {
		t.Fatalf("Create() error = %v, want %v", err, store.appendErr)
	}
	if store.appendCalls != 1 {
		t.Fatalf("append calls = %d, want 1", store.appendCalls)
	}
}

func TestHandlerValidatesCommandsBeforeUsingStore(t *testing.T) {
	t.Parallel()

	store := newMemoryStore()
	handler := newTestHandler(t, store, "task-123", "event-123")

	_, err := handler.Create(context.Background(), command.CreateTask{Name: "missing title", Status: "Doing"})
	if !errors.Is(err, command.ErrInvalidCommand) {
		t.Fatalf("Create() error = %v, want ErrInvalidCommand", err)
	}
	if store.appendCalls != 0 {
		t.Fatal("invalid command must not use the store")
	}
}

func newTestHandler(t *testing.T, store *memoryStore, ids ...string) *command.Handler {
	t.Helper()
	index := 0
	handler, err := command.NewHandler(
		store,
		command.WithIDGenerator(func() (string, error) {
			if index >= len(ids) {
				return "", fmt.Errorf("id generator exhausted")
			}
			id := ids[index]
			index++
			return id, nil
		}),
		command.WithClock(func() time.Time {
			return time.Date(2026, time.August, 9, 1, 2, 3, 0, time.UTC)
		}),
	)
	if err != nil {
		t.Fatalf("NewHandler() error = %v", err)
	}
	return handler
}

type memoryStore struct {
	events      map[eventstream.StreamID][]eventstream.Envelope
	appendErr   error
	appendCalls int
}

func newMemoryStore() *memoryStore {
	return &memoryStore{events: make(map[eventstream.StreamID][]eventstream.Envelope)}
}

func (s *memoryStore) Load(ctx context.Context, stream eventstream.StreamID) ([]eventstream.Envelope, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	events := s.events[stream]
	if len(events) == 0 {
		return nil, command.ErrAggregateNotFound
	}
	return append([]eventstream.Envelope(nil), events...), nil
}

func (s *memoryStore) Append(ctx context.Context, stream eventstream.StreamID, expectedVersion uint64, events ...eventstream.Envelope) error {
	s.appendCalls++
	if err := ctx.Err(); err != nil {
		return err
	}
	if s.appendErr != nil {
		return s.appendErr
	}
	current := uint64(len(s.events[stream]))
	if current != expectedVersion {
		return &command.VersionConflictError{Expected: expectedVersion, Actual: current}
	}
	for index, event := range events {
		wantVersion := expectedVersion + uint64(index) + 1
		if event.Aggregate.Type != stream.Type || event.Aggregate.ID != stream.ID || event.Aggregate.Version != wantVersion {
			return fmt.Errorf("invalid event at index %d: %#v", index, event)
		}
		s.events[stream] = append(s.events[stream], event)
	}
	return nil
}
