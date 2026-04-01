package store

import (
	"context"
	"testing"
)

const queue = "queue"
const message = "message"

func newTestStore(t *testing.T) *SqliteStore {
	store, err := NewSqliteStore(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { store.Close() })
	return store
}

func TestEnqueueMessage(t *testing.T) {
	store := newTestStore(t)

	store.Enqueue(context.Background(), queue, []byte(message))

	msg, err := store.Peek(context.Background(), queue)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if string(msg.Body) != message {
		t.Fatalf("Message peeked did not match queued")
	}
}
