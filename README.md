# service-bus

A lightweight local message queue server for Go, backed by SQLite. Built to explore and replicate the core mechanics of brokers like Azure Service Bus — peek-lock delivery, acknowledgement, dead-lettering, and retry — without any external infrastructure dependencies.

> ⚠️ Work in progress. Core messaging semantics and HTTP API are implemented; client library and additional features are planned.

---

## Why

I work with Azure Service Bus daily and wanted to understand the mechanics from the ground up — how peek-lock actually works, how dead-letter queues get populated, and how delivery guarantees are enforced at the storage layer. This project is that exploration, implemented in Go with SQLite as the backing store.

---

## Features

- **Queue management** — create, list queues with configurable max delivery attempts
- **Enqueue** — publish messages to a named queue
- **Peek** — inspect the next available message without consuming it
- **Receive (peek-lock)** — claim a message with a 30-second visibility lock; prevents double-delivery
- **Ack** — delete a successfully processed message by lock token
- **Nack** — release or dead-letter a failed message by lock token
- **Dead-letter queue** — messages exceeding `max_delivery` attempts are automatically promoted to DLQ
- **Queue stats** — inspect message counts and DLQ depth
- **HTTP API** — language-agnostic REST interface; use from any stack

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
- **foreign keys enforced** — queue existence is validated on enqueue

---

## Running the Server

```bash
make build
./bin/servicebus
```

The server listens on `:5800` by default and creates `./db/servicebus.db` on first run. Delete the file to reset all state.

---

## HTTP API

### Queues

#### Create a queue
```
POST /queues/{name}?maxDelivery=5
```
Creates a named queue. `maxDelivery` is optional and defaults to 5. Returns `409 Conflict` if the queue already exists.

```bash
curl -X POST http://localhost:5800/queues/jobs?maxDelivery=3
```

#### List queues
```
GET /queues
```

---

### Messages

#### Enqueue a message
```
POST /queues/{name}/messages
```
Request body is the raw message payload.

```bash
curl -X POST http://localhost:5800/queues/jobs \
  -d '{"task":"send_email","to":"user@example.com"}'
```

#### Receive a message (peek-lock)
```
GET /queues/{name}/messages
```
Returns the next available message and a `lockToken`. The message is locked for 30 seconds. Store the `lockToken` — it is required to Ack or Nack.

```json
{
  "id": "abc-123",
  "body": "{\"task\":\"send_email\"}",
  "lockToken": "xyz-789"
}
```

Returns `204 No Content` if the queue is empty.

#### Peek a message
```
GET /queues/{name}/messages?peek=true
```
Returns the next available message without locking it.

---

### Ack / Nack

Both endpoints accept the lock token returned from Receive.

#### Acknowledge (delete on success)
```
POST /ack
```
```json
{ "lockToken": "xyz-789" }
```

#### Nack (retry or dead-letter)
```
POST /nack
```
```json
{ "lockToken": "xyz-789" }
```
If `delivery_count >= max_delivery`, the message is moved to the DLQ. Otherwise it is returned to the queue.

---

### Stats

```
GET /queues/{name}/stats
```
```json
{
  "queueName": "jobs",
  "messageCount": 4,
  "dlqCount": 1
}
```

---

## Usage (Go embed)

The store layer can be used directly in Go projects without running the HTTP server:

```go
store, err := store.NewSqliteStore("./myqueue.db")
if err != nil {
    log.Fatal(err)
}
defer store.Close()

ctx := context.Background()

err = store.CreateQueue(ctx, "jobs", 3)
err = store.Enqueue(ctx, "jobs", []byte(`{"task":"send_email"}`))

msg, err := store.Receive(ctx, "jobs")
err = store.Ack(ctx, *msg.LockToken)

// or on failure:
err = store.Nack(ctx, *msg.LockToken)
```

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

## Planned

- [ ] Configurable lock duration (currently fixed at 30s)
- [ ] DLQ drain / requeue support
- [ ] Delete queue
- [ ] Purge queue
- [ ] Message TTL / expiry
- [ ] Go client package (`/client`)
- [ ] Configurable port and db path via CLI flags

---

## Dependencies

- [`modernc.org/sqlite`](https://pkg.go.dev/modernc.org/sqlite) — pure Go SQLite driver, no CGo required
- [`github.com/google/uuid`](https://pkg.go.dev/github.com/google/uuid) — lock token generation
- [`github.com/go-chi/chi`](https://pkg.go.dev/github.com/go-chi/chi) — HTTP router

---

## License

MIT
