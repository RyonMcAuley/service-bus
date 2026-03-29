package store

const schemaMigration = `
CREATE TABLE IF NOT EXISTS queues (
    name            TEXT PRIMARY KEY,
    max_delivery    INTEGER NOT NULL DEFAULT 5,
    created_at      DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS messages (
    id              TEXT PRIMARY KEY,
    queue_name      TEXT NOT NULL,
    body            BLOB NOT NULL,
    enqueued_at     DATETIME NOT NULL,
    visible_at      DATETIME NOT NULL,
    delivery_count  INTEGER NOT NULL DEFAULT 0,
    lock_token      TEXT,
    locked_until    DATETIME,
    is_dlq          INTEGER NOT NULL DEFAULT 0,
    FOREIGN KEY (queue_name) REFERENCES queues(name)
);

CREATE INDEX IF NOT EXISTS idx_messages_queue 
    ON messages(queue_name, is_dlq, visible_at);
`

const queryCreateQueue = `
INSERT INTO queues (name, max_delivery, created_at)
VALUES (?, ?, ?)
`

const queryListQueues = `
SELECT name, max_delivery, created_at
	FROM queues
`

const queryEnqueue = `
INSERT INTO messages (id, queue_name, body, enqueued_at, visible_at, delivery_count, is_dlq)
VALUES (?,?,?,?,?,?,?)
`

const queryPeek = `
SELECT id, queue_name, body, enqueued_at, visible_at, delivery_count, lock_token, locked_until, is_dlq
FROM messages
WHERE queue_name = ?
AND is_dlq = 0
AND visible_at <= ?
ORDER BY enqueued_at ASC
LIMIT 1
`
