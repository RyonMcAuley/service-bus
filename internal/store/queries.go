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
	AND lock_token IS NULL
ORDER BY enqueued_at ASC
	LIMIT 1
`

const queryGetStats = `
SELECT
	COUNT(CASE WHEN is_dlq = 0 THEN 1 END) as message_count,
	COUNT(CASE WHEN is_dlq = 1 THEN 1 END) as dlq_count
	FROM messages
	WHERE queue_name = ?
`

const queryReceive = `
SELECT id, queue_name, body, enqueued_at, visible_at, delivery_count, lock_token, locked_until, is_dlq
	FROM messages
	WHERE queue_name = ?
	AND is_dlq = 0
	AND visible_at <= ?
	AND lock_token IS NULL
ORDER BY enqueued_at ASC
	LIMIT 1
`

const queryReceiveUpdate = `
UPDATE messages
SET lock_token = ?, locked_until = ?, delivery_count = delivery_count + 1
	WHERE id = ?
`

const queryAck = `
DELETE
FROM messages
WHERE lock_token = ?
`

const queryNackFind = `
SELECT id, queue_name, body, enqueued_at, visible_at, delivery_count, lock_token, locked_until, is_dlq, q.max_delivery
	FROM messages
	JOIN queues q on q.name = messages.queue_name
	WHERE lock_token = ?
LIMIT 1
`

const queryNackDLQ = `
UPDATE messages
SET lock_token = NULL, locked_until = NULL, is_dlq = 1
WHERE id = ?
`

const queryNackRetry = `
UPDATE messages
SET lock_token = NULL, locked_until = NULL, visible_at = ?
WHERE id = ?
`
