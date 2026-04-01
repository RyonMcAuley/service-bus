# service-bus

A lightweight, embedded message queue library for Go, backed by SQLite. Built to explore and replicate the core mechanics of brokers like Azure Service Bus — peek-lock delivery, acknowledgement, dead-lettering, and retry — without any external infrastructure dependencies.

> ⚠️ Work in progress. Core messaging semantics are implemented; HTTP API and additional features are planned.

---

## Why

I work with Azure Service Bus daily and wanted to understand the mechanics from the ground up — how peek-lock actually works, how dead-letter queues get populated, and how delivery guarantees are enforced at the storage layer. This project is that exploration, implemented in Go with SQLite as the backing store.

---

## Features

- **Queue management** — create, delete, list queues with configurable max delivery attempts
- **Enqueue** — publish messages to a named queue
- **Peek** — inspect the next available message without consuming it
- **Receive (peek-lock)** — claim a message with a 30-second visibility lock; prevents double-delivery
- **Ack** — delete a successfully processed message by lock token
- **Nack** — release or dead-letter a failed message by lock token
- **Dead-letter queue** — messages exceeding `max_delivery` attempts are automatically promoted to DLQ
- **Queue stats** — inspect message counts and queue state

---

## How It Works

### Peek-Lock Delivery

Rather than deleting a message on receive, `service-bus` implements a **peek-lock** pattern:

1. `Receive` selects the next available message and sets a `lock_token` and `locked_until` timestamp (30s window)
2. The message is invisible to other consumers while locked
3. On `Ack`, the message is deleted — delivery confirmed
4. On `Nack`, the message is either returned to the queue (lock cleared, `visible_at` reset) or promoted to the DLQ if `delivery_count >= max_delivery`

This matches the core delivery guarantee model used by Azure Service Bus and Amazon SQS.

### Dead-Letter Queue

Each queue has a built-in DLQ — no separate queue required. Messages that exceed their maximum delivery attempts are flagged with `is_dlq = 1` and removed from normal receive flow. DLQ messages can be inspected independently.

### Storage

All state is persisted in a single SQLite file. The schema uses a composite index on `(queue_name, is_dlq, visible_at)` to make the receive hot path efficient.

SQLite is configured with:
- **WAL journal mode** — allows concurrent readers alongside the single writer without blocking
- **busy_timeout = 5000ms** — graceful handling of write contention rather than immediate failure
- **max open conns = 1** — prevents SQLite locking errors under concurrent access

---

## Schema

```sql
CREATE TABLE queues (
    name         TEXT PRIMARY KEY,
    max_delivery INTEGER NOT NULL DEFAULT 5,
    created_at   DATETIME NOT NULL
);

CREATE TABLE messages (
    id             TEXT PRIMARY KEY,
    queue_name     TEXT NOT NULL,
    body           BLOB NOT NULL,
    enqueued_at    DATETIME NOT NULL,
    visible_at     DATETIME NOT NULL,
    delivery_count INTEGER NOT NULL DEFAULT 0,
    lock_token     TEXT,
    locked_until   DATETIME,
    is_dlq         INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (queue_name) REFERENCES queues(name)
);

CREATE INDEX idx_messages_queue ON messages(queue_name, is_dlq, visible_at);
```

---

## Usage

```go
store, err := servicebus.NewSqliteStore("./myqueue.db")
if err != nil {
    log.Fatal(err)
}
defer store.Close()

ctx := context.Background()

// Create a queue with max 3 delivery attempts
err = store.CreateQueue(ctx, "jobs", 3)

// Enqueue a message
err = store.Enqueue(ctx, "jobs", []byte(`{"task":"send_email"}`))

// Receive (peek-lock)
msg, err := store.Receive(ctx, "jobs")

// Acknowledge successful processing
err = store.Ack(ctx, *msg.LockToken)

// Or nack on failure — returns to queue or promotes to DLQ
err = store.Nack(ctx, *msg.LockToken)
```

---

## Planned

- [ ] HTTP API (`net/http`) for language-agnostic usage
- [ ] Configurable lock duration (currently fixed at 30s)
- [ ] Message TTL / expiry
- [ ] DLQ drain / requeue support
- [ ] Metrics endpoint (queue depth, DLQ count)

---

## Dependencies

- [`modernc.org/sqlite`](https://pkg.go.dev/modernc.org/sqlite) — pure Go SQLite driver, no CGo required
- [`github.com/google/uuid`](https://pkg.go.dev/github.com/google/uuid) — lock token generation

---

## License

MIT
