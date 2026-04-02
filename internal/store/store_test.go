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

func newTestStoreWithQueue(t *testing.T) *SqliteStore {
	store := newTestStore(t)

	ctx := context.Background()
	store.CreateQueue(ctx, queue, 2)

	return store
}

func TestEnqueueMessage(t *testing.T) {
	store := newTestStoreWithQueue(t)

	store.Enqueue(context.Background(), queue, []byte(message))

	msg, err := store.Peek(context.Background(), queue)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if string(msg.Body) != message {
		t.Fatalf("Message peeked did not match queued")
	}
}

func TestReceiveMessage(t *testing.T) {
	store := newTestStoreWithQueue(t)

	store.Enqueue(context.Background(), queue, []byte(message))

	msg, err := store.Receive(context.Background(), queue)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if string(msg.Body) != message {
		t.Fatalf("Message received did not match queued")
	}
}

func TestPeekAfterReceive(t *testing.T) {
	store := newTestStoreWithQueue(t)

	ctx := context.Background()

	store.Enqueue(ctx, queue, []byte("Shouldn't get"))
	store.Enqueue(ctx, queue, []byte("Should get"))

	_, err := store.Receive(context.Background(), queue)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	peeked, err := store.Peek(ctx, queue)

	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if string(peeked.Body) != "Should get" {
		t.Fatalf("Did not peek message past received")
	}
}

func TestAck(t *testing.T) {
	store := newTestStoreWithQueue(t)

	ctx := context.Background()

	store.Enqueue(ctx, queue, []byte(message))

	msg, err := store.Receive(ctx, queue)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	err = store.Ack(ctx, *msg.LockToken)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	peeked, err := store.Peek(ctx, queue)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if peeked != nil {
		t.Fatalf("Expected queue to be empty after message Acknowledged")
	}
}

func TestNack(t *testing.T) {
	store := newTestStoreWithQueue(t)

	ctx := context.Background()

	store.Enqueue(ctx, queue, []byte(message))

	msg, err := store.Receive(ctx, queue)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	err = store.Nack(ctx, *msg.LockToken)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	msg, err = store.Receive(ctx, queue)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if msg.DeliveryCount < 1 {
		t.Fatalf("Expected 1 delivery count, got %d", msg.DeliveryCount)
	}
}

func TestNackDLQ(t *testing.T) {
	store := newTestStoreWithQueue(t)

	ctx := context.Background()

	store.Enqueue(ctx, queue, []byte(message))

	msg, err := store.Receive(ctx, queue)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	err = store.Nack(ctx, *msg.LockToken)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	msg, err = store.Receive(ctx, queue)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	err = store.Nack(ctx, *msg.LockToken)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	peeked, err := store.Peek(ctx, queue)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if peeked != nil {
		t.Fatalf("Expected Peek to return nil after message entered DLQ")
	}
}
