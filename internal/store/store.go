package store

import "context"

type MessageStore interface {
	CreateQueue(ctx context.Context, name string, maxDelivery int) error
	ListQueues(ctx context.Context) ([]*Queue, error)

	Enqueue(ctx context.Context, queueName string, body []byte) error

	GetStats(ctx context.Context, queueName string) (*Stats, error)

	Close() error
}
