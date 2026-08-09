# ADR 0001: Event-sourced task commands with a transactional outbox

- Status: accepted
- Date: 2026-08-09

## Context

The original task flow created a random aggregate for every HTTP command and
published a JSON array of loosely typed events directly to Kafka. Updates and
deletes therefore had a different aggregate ID from the URL, versions stayed at
zero, and a database write plus a Kafka write could not be made atomic. The
consumer was also unable to distinguish a malformed message from a successfully
applied one.

## Decision

Task commands target one stream identified by `(aggregate_type, aggregate_id)`.
The command handler owns validation, event IDs, timestamps, and optimistic
concurrency; it depends only on the `EventStore` port in
`internal/task/command`. The store's `Append` operation is the future atomic
boundary for the immutable event row and its outbox row. The command layer does
not call Kafka.

Every event uses one canonical `eventstream.Envelope`:

```json
{
  "id": "event-uuid",
  "type": "task.created",
  "schema_version": 1,
  "aggregate": {"type": "Task", "id": "task-uuid", "version": 1},
  "occurred_at": "2026-08-09T01:07:06Z",
  "data": {"title": "...", "name": "...", "image": "", "description": "", "status": "Doing"}
}
```

The aggregate ID is the Kafka key, and each Kafka value contains exactly one
envelope. The existing topic spelling `eventStore_Task` is retained while the
runtime is migrated; deployments that already contain the legacy JSON-array
format must use a new topic/consumer group or reset offsets before enabling the
canonical decoder. There is intentionally no silent dual-format decoder.

Task stream versions start at one and must be contiguous. Create appends version
1 with expected version 0. Update and delete require the caller's expected
version and append `expected + 1`. Delete is a tombstone: the stream remains
addressable, its version is retained, and later commands are rejected.

## Consequences

Positive:

- Aggregate ordering is stable because all events share one Kafka key.
- A stale command has a typed, inspectable conflict instead of overwriting state.
- Event payloads do not duplicate identity and can evolve through
  `schema_version`.
- The command domain can be tested without Kafka, MySQL, or network mocks.

Trade-offs:

- This is a wire-format migration from the old array/base64 representation.
- The first phase introduces a port; the MySQL event store, outbox publisher,
  projection, and runtime wiring follow in the next phase.
- Existing `Doing`/`Done` status spelling is preserved for compatibility.

## Rejected alternatives

1. **Keep publishing arrays of events.** This loses per-event ordering and
   makes retries/poison-message handling ambiguous.
2. **Write MySQL and Kafka in separate application calls.** A process crash
   between writes creates an unrecoverable dual-write gap.
3. **Let HTTP handlers construct envelopes.** That duplicates domain rules and
   makes non-HTTP command paths inconsistent.
