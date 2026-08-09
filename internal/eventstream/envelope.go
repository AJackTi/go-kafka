// Package eventstream defines the versioned event contract shared by command
// handlers, event stores, and message transports.
package eventstream

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"
)

var (
	ErrInvalidEnvelope = errors.New("invalid event envelope")
	ErrInvalidMessage  = errors.New("invalid event message")
)

// AggregateRef identifies the stream to which an event belongs.
type AggregateRef struct {
	Type    string `json:"type"`
	ID      string `json:"id"`
	Version uint64 `json:"version"`
}

// StreamID identifies an aggregate stream without a version.
type StreamID struct {
	Type string
	ID   string
}

// Envelope is the canonical, one-event-per-message transport format.
type Envelope struct {
	ID            string          `json:"id"`
	Type          string          `json:"type"`
	SchemaVersion uint16          `json:"schema_version"`
	Aggregate     AggregateRef    `json:"aggregate"`
	OccurredAt    time.Time       `json:"occurred_at"`
	Data          json.RawMessage `json:"data"`
}

// Message is the transport-neutral representation sent to a broker.
type Message struct {
	Topic string
	Key   []byte
	Value []byte
}

// NewEnvelope creates and validates an envelope from a JSON-serializable
// payload. Timestamps are normalized to UTC for deterministic persistence.
func NewEnvelope(
	id string,
	eventType string,
	schemaVersion uint16,
	aggregate AggregateRef,
	occurredAt time.Time,
	payload any,
) (Envelope, error) {
	data, err := json.Marshal(payload)
	if err != nil {
		return Envelope{}, fmt.Errorf("marshal event payload: %w", err)
	}

	envelope := Envelope{
		ID:            id,
		Type:          eventType,
		SchemaVersion: schemaVersion,
		Aggregate:     aggregate,
		OccurredAt:    normalizeTime(occurredAt),
		Data:          append(json.RawMessage(nil), data...),
	}
	if err := envelope.Validate(); err != nil {
		return Envelope{}, err
	}

	return envelope, nil
}

// Validate checks the wire-level invariants that every consumer can rely on.
func (e Envelope) Validate() error {
	if strings.TrimSpace(e.ID) == "" || strings.TrimSpace(e.Type) == "" || e.ID != strings.TrimSpace(e.ID) || e.Type != strings.TrimSpace(e.Type) {
		return fmt.Errorf("%w: id and type are required", ErrInvalidEnvelope)
	}
	if e.SchemaVersion == 0 {
		return fmt.Errorf("%w: schema_version must be positive", ErrInvalidEnvelope)
	}
	if strings.TrimSpace(e.Aggregate.Type) == "" || strings.TrimSpace(e.Aggregate.ID) == "" || e.Aggregate.Type != strings.TrimSpace(e.Aggregate.Type) || e.Aggregate.ID != strings.TrimSpace(e.Aggregate.ID) {
		return fmt.Errorf("%w: aggregate type and id are required", ErrInvalidEnvelope)
	}
	if e.Aggregate.Version == 0 {
		return fmt.Errorf("%w: aggregate version must be positive", ErrInvalidEnvelope)
	}
	if e.OccurredAt.IsZero() {
		return fmt.Errorf("%w: occurred_at is required", ErrInvalidEnvelope)
	}
	if len(e.Data) == 0 || !json.Valid(e.Data) {
		return fmt.Errorf("%w: data must be valid JSON", ErrInvalidEnvelope)
	}

	return nil
}

// Encode serializes one validated envelope. An array is intentionally not a
// valid output shape: each broker message carries exactly one event.
func Encode(e Envelope) ([]byte, error) {
	if err := e.Validate(); err != nil {
		return nil, err
	}
	e.OccurredAt = normalizeTime(e.OccurredAt)

	return json.Marshal(e)
}

// Decode parses exactly one envelope and rejects trailing JSON values.
func Decode(data []byte) (Envelope, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	var envelope Envelope
	if err := decoder.Decode(&envelope); err != nil {
		return Envelope{}, fmt.Errorf("%w: decode: %w", ErrInvalidEnvelope, err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		if err == nil {
			return Envelope{}, fmt.Errorf("%w: multiple JSON values", ErrInvalidEnvelope)
		}
		return Envelope{}, fmt.Errorf("%w: trailing data: %w", ErrInvalidEnvelope, err)
	}
	if err := envelope.Validate(); err != nil {
		return Envelope{}, err
	}
	envelope.OccurredAt = normalizeTime(envelope.OccurredAt)

	return envelope, nil
}

// StreamID returns the stream reference carried by the envelope.
func (e Envelope) StreamID() StreamID {
	return StreamID{Type: e.Aggregate.Type, ID: e.Aggregate.ID}
}

func normalizeTime(value time.Time) time.Time {
	return value.Round(0).UTC()
}

// NewMessage maps an envelope to a broker message. The aggregate ID is the
// key so all versions of one stream remain ordered in Kafka.
func NewMessage(topicPrefix string, envelope Envelope) (Message, error) {
	if err := envelope.Validate(); err != nil {
		return Message{}, err
	}
	topicPrefix = strings.TrimSpace(topicPrefix)
	if topicPrefix == "" {
		return Message{}, fmt.Errorf("%w: topic prefix is required", ErrInvalidMessage)
	}

	value, err := Encode(envelope)
	if err != nil {
		return Message{}, err
	}

	return Message{
		Topic: topicPrefix + "_" + envelope.Aggregate.Type,
		Key:   []byte(envelope.Aggregate.ID),
		Value: value,
	}, nil
}
