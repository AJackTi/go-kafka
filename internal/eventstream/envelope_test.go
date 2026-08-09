package eventstream_test

import (
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/AJackTi/go-kafka/internal/eventstream"
)

func TestEnvelopeRoundTripPreservesCanonicalFields(t *testing.T) {
	t.Parallel()

	when := time.Date(2026, time.August, 9, 8, 7, 6, 500, time.FixedZone("ICT", 7*60*60))
	envelope, err := eventstream.NewEnvelope(
		"evt-123",
		"task.created",
		1,
		eventstream.AggregateRef{Type: "Task", ID: "task-123", Version: 1},
		when,
		map[string]string{"title": "ship it"},
	)
	if err != nil {
		t.Fatalf("NewEnvelope() error = %v", err)
	}

	encoded, err := eventstream.Encode(envelope)
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	decoded, err := eventstream.Decode(encoded)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if decoded.ID != envelope.ID || decoded.Type != envelope.Type {
		t.Fatalf("identity changed after round trip: %#v", decoded)
	}
	if decoded.Aggregate != envelope.Aggregate {
		t.Fatalf("aggregate changed after round trip: %#v", decoded.Aggregate)
	}
	if !decoded.OccurredAt.Equal(when.UTC()) {
		t.Fatalf("occurred_at = %s, want %s", decoded.OccurredAt, when.UTC())
	}
	if string(decoded.Data) != string(envelope.Data) {
		t.Fatalf("data changed after round trip: %s", decoded.Data)
	}
}

func TestEnvelopeRejectsInvalidContract(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		envelope eventstream.Envelope
		wantErr  error
	}{
		{
			name:    "missing event id",
			wantErr: eventstream.ErrInvalidEnvelope,
			envelope: eventstream.Envelope{
				Type:          "task.created",
				SchemaVersion: 1,
				Aggregate:     eventstream.AggregateRef{Type: "Task", ID: "task-123", Version: 1},
				OccurredAt:    time.Now(),
				Data:          json.RawMessage(`{}`),
			},
		},
		{
			name:    "zero aggregate version",
			wantErr: eventstream.ErrInvalidEnvelope,
			envelope: eventstream.Envelope{
				ID:            "evt-123",
				Type:          "task.created",
				SchemaVersion: 1,
				Aggregate:     eventstream.AggregateRef{Type: "Task", ID: "task-123"},
				OccurredAt:    time.Now(),
				Data:          json.RawMessage(`{}`),
			},
		},
		{
			name:    "zero schema version",
			wantErr: eventstream.ErrInvalidEnvelope,
			envelope: eventstream.Envelope{
				ID:         "evt-123",
				Type:       "task.created",
				Aggregate:  eventstream.AggregateRef{Type: "Task", ID: "task-123", Version: 1},
				OccurredAt: time.Now(),
				Data:       json.RawMessage(`{}`),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.envelope.Validate(); !errors.Is(err, tt.wantErr) {
				t.Fatalf("Validate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestDecodeRejectsTrailingJSON(t *testing.T) {
	t.Parallel()

	_, err := eventstream.Decode([]byte(`{"id":"evt-123","type":"task.created","schema_version":1,"aggregate":{"type":"Task","id":"task-123","version":1},"occurred_at":"2026-08-09T01:07:06Z","data":{}} {}`))
	if !errors.Is(err, eventstream.ErrInvalidEnvelope) {
		t.Fatalf("Decode() error = %v, want ErrInvalidEnvelope", err)
	}
}

func TestMessageUsesAggregateKeyAndSingleEnvelope(t *testing.T) {
	t.Parallel()

	envelope, err := eventstream.NewEnvelope(
		"evt-123",
		"task.created",
		1,
		eventstream.AggregateRef{Type: "Task", ID: "task-123", Version: 1},
		time.Date(2026, time.August, 9, 1, 2, 3, 0, time.UTC),
		map[string]string{"title": "ship it"},
	)
	if err != nil {
		t.Fatalf("NewEnvelope() error = %v", err)
	}

	message, err := eventstream.NewMessage("eventStore", envelope)
	if err != nil {
		t.Fatalf("NewMessage() error = %v", err)
	}
	if message.Topic != "eventStore_Task" {
		t.Fatalf("topic = %q, want eventStore_Task", message.Topic)
	}
	if string(message.Key) != "task-123" {
		t.Fatalf("key = %q, want task-123", message.Key)
	}
	decoded, err := eventstream.Decode(message.Value)
	if err != nil {
		t.Fatalf("Decode(message.Value) error = %v", err)
	}
	if decoded.ID != "evt-123" {
		t.Fatalf("message contains event %q, want evt-123", decoded.ID)
	}
}

func TestEncodeUsesStableSnakeCaseAndUTC(t *testing.T) {
	t.Parallel()

	when := time.Date(2026, time.August, 9, 8, 7, 6, 500, time.FixedZone("ICT", 7*60*60))
	envelope, err := eventstream.NewEnvelope(
		"evt-123",
		"task.created",
		1,
		eventstream.AggregateRef{Type: "Task", ID: "task-123", Version: 1},
		when,
		map[string]string{"title": "ship it"},
	)
	if err != nil {
		t.Fatalf("NewEnvelope() error = %v", err)
	}

	encoded, err := eventstream.Encode(envelope)
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}
	want := `{"id":"evt-123","type":"task.created","schema_version":1,"aggregate":{"type":"Task","id":"task-123","version":1},"occurred_at":"2026-08-09T01:07:06.0000005Z","data":{"title":"ship it"}}`
	if string(encoded) != want {
		t.Fatalf("encoded envelope = %s, want %s", encoded, want)
	}
}

func FuzzDecodeNeverPanics(f *testing.F) {
	f.Add([]byte(`{"id":"evt-123","type":"task.created","schema_version":1,"aggregate":{"type":"Task","id":"task-123","version":1},"occurred_at":"2026-08-09T01:07:06Z","data":{}}`))
	f.Add([]byte(`not json`))
	f.Fuzz(func(t *testing.T, data []byte) {
		if _, err := eventstream.Decode(data); err != nil {
			return
		}
	})
}
