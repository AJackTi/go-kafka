package command

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/AJackTi/go-kafka/internal/eventstream"
	"github.com/AJackTi/go-kafka/internal/task"
)

type IDGenerator func() (string, error)

type Clock func() time.Time

type Option func(*Handler)

// WithIDGenerator injects deterministic IDs in tests and other controlled
// environments.
func WithIDGenerator(generator IDGenerator) Option {
	return func(handler *Handler) { handler.newID = generator }
}

// WithClock injects the event clock used by the command handler.
func WithClock(clock Clock) Option {
	return func(handler *Handler) { handler.now = clock }
}

type Handler struct {
	store EventStore
	newID IDGenerator
	now   Clock
}

func NewHandler(store EventStore, options ...Option) (*Handler, error) {
	if store == nil {
		return nil, fmt.Errorf("%w: event store is required", ErrInvalidCommand)
	}
	handler := &Handler{
		store: store,
		newID: func() (string, error) { return uuid.NewString(), nil },
		now:   func() time.Time { return time.Now().UTC() },
	}
	for _, option := range options {
		if option != nil {
			option(handler)
		}
	}
	if handler.newID == nil || handler.now == nil {
		return nil, fmt.Errorf("%w: id generator and clock are required", ErrInvalidCommand)
	}
	return handler, nil
}

func (h *Handler) Create(ctx context.Context, command CreateTask) (Receipt, error) {
	if err := contextError(ctx); err != nil {
		return Receipt{}, err
	}
	if err := command.validate(); err != nil {
		return Receipt{}, err
	}

	aggregateID := strings.TrimSpace(command.ID)
	if aggregateID == "" {
		var err error
		aggregateID, err = h.newID()
		if err != nil {
			return Receipt{}, fmt.Errorf("generate aggregate id: %w", err)
		}
	}
	if strings.TrimSpace(aggregateID) == "" {
		return Receipt{}, fmt.Errorf("%w: aggregate id is required", ErrInvalidCommand)
	}
	eventID, err := h.newID()
	if err != nil {
		return Receipt{}, fmt.Errorf("generate event id: %w", err)
	}
	envelope, err := eventstream.NewEnvelope(
		eventID,
		task.CreatedEventType,
		task.EventSchemaVersion,
		eventstream.AggregateRef{Type: task.AggregateType, ID: aggregateID, Version: 1},
		h.now(),
		command.attributes(),
	)
	if err != nil {
		return Receipt{}, err
	}
	stream := eventstream.StreamID{Type: task.AggregateType, ID: aggregateID}
	if err := h.store.Append(ctx, stream, 0, envelope); err != nil {
		return Receipt{}, err
	}
	return receiptFor(envelope), nil
}

func (h *Handler) Update(ctx context.Context, command UpdateTask) (Receipt, error) {
	if err := contextError(ctx); err != nil {
		return Receipt{}, err
	}
	if err := command.validate(); err != nil {
		return Receipt{}, err
	}

	stream := eventstream.StreamID{Type: task.AggregateType, ID: strings.TrimSpace(command.ID)}
	state, err := h.load(ctx, stream)
	if err != nil {
		return Receipt{}, err
	}
	if state.Deleted {
		return Receipt{}, ErrAggregateDeleted
	}
	if state.Version != command.ExpectedVersion {
		return Receipt{}, &VersionConflictError{Expected: command.ExpectedVersion, Actual: state.Version}
	}

	eventID, err := h.newID()
	if err != nil {
		return Receipt{}, fmt.Errorf("generate event id: %w", err)
	}
	envelope, err := eventstream.NewEnvelope(
		eventID,
		task.UpdatedEventType,
		task.EventSchemaVersion,
		eventstream.AggregateRef{Type: task.AggregateType, ID: stream.ID, Version: state.Version + 1},
		h.now(),
		command.attributes(),
	)
	if err != nil {
		return Receipt{}, err
	}
	if err := h.store.Append(ctx, stream, state.Version, envelope); err != nil {
		return Receipt{}, err
	}
	return receiptFor(envelope), nil
}

func (h *Handler) Delete(ctx context.Context, command DeleteTask) (Receipt, error) {
	if err := contextError(ctx); err != nil {
		return Receipt{}, err
	}
	if err := command.validate(); err != nil {
		return Receipt{}, err
	}

	stream := eventstream.StreamID{Type: task.AggregateType, ID: strings.TrimSpace(command.ID)}
	state, err := h.load(ctx, stream)
	if err != nil {
		return Receipt{}, err
	}
	if state.Deleted {
		return Receipt{}, ErrAggregateDeleted
	}
	if state.Version != command.ExpectedVersion {
		return Receipt{}, &VersionConflictError{Expected: command.ExpectedVersion, Actual: state.Version}
	}

	eventID, err := h.newID()
	if err != nil {
		return Receipt{}, fmt.Errorf("generate event id: %w", err)
	}
	envelope, err := eventstream.NewEnvelope(
		eventID,
		task.DeletedEventType,
		task.EventSchemaVersion,
		eventstream.AggregateRef{Type: task.AggregateType, ID: stream.ID, Version: state.Version + 1},
		h.now(),
		struct{}{},
	)
	if err != nil {
		return Receipt{}, err
	}
	if err := h.store.Append(ctx, stream, state.Version, envelope); err != nil {
		return Receipt{}, err
	}
	return receiptFor(envelope), nil
}

func (h *Handler) load(ctx context.Context, stream eventstream.StreamID) (*task.State, error) {
	history, err := h.store.Load(ctx, stream)
	if err != nil {
		if errors.Is(err, ErrAggregateNotFound) {
			return nil, err
		}
		return nil, err
	}
	if len(history) == 0 {
		return nil, ErrAggregateNotFound
	}
	state, err := task.Rehydrate(stream, history)
	if err != nil {
		if errors.Is(err, task.ErrAggregateDeleted) {
			return nil, ErrAggregateDeleted
		}
		return nil, err
	}
	return state, nil
}

func receiptFor(envelope eventstream.Envelope) Receipt {
	return Receipt{
		AggregateID: envelope.Aggregate.ID,
		EventID:     envelope.ID,
		Version:     envelope.Aggregate.Version,
		Type:        envelope.Type,
		Event:       envelope,
	}
}

func contextError(ctx context.Context) error {
	if ctx == nil {
		return fmt.Errorf("%w: context is required", ErrInvalidCommand)
	}
	return ctx.Err()
}

func (c CreateTask) validate() error {
	if strings.TrimSpace(c.ID) != c.ID {
		return fmt.Errorf("%w: id cannot have surrounding whitespace", ErrInvalidCommand)
	}
	if c.ID != "" && strings.TrimSpace(c.ID) == "" {
		return fmt.Errorf("%w: id is required when provided", ErrInvalidCommand)
	}
	if err := c.attributes().Validate(); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidCommand, err)
	}
	return nil
}

func (c UpdateTask) validate() error {
	if strings.TrimSpace(c.ID) == "" || strings.TrimSpace(c.ID) != c.ID {
		return fmt.Errorf("%w: id is required", ErrInvalidCommand)
	}
	if c.ExpectedVersion == 0 {
		return fmt.Errorf("%w: expected version must be positive", ErrInvalidCommand)
	}
	if err := c.attributes().Validate(); err != nil {
		return fmt.Errorf("%w: %w", ErrInvalidCommand, err)
	}
	return nil
}

func (c DeleteTask) validate() error {
	if strings.TrimSpace(c.ID) == "" || strings.TrimSpace(c.ID) != c.ID {
		return fmt.Errorf("%w: id is required", ErrInvalidCommand)
	}
	if c.ExpectedVersion == 0 {
		return fmt.Errorf("%w: expected version must be positive", ErrInvalidCommand)
	}
	return nil
}
