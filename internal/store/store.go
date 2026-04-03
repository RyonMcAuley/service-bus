package store

import "context"

type MessageStore interface {
	CreateQueue(ctx context.Context, name string, maxDelivery int) error
	ListQueues(ctx context.Context) ([]*Queue, error)
	Peek(ctx context.Context, queueName string) (*Message, error)

	Enqueue(ctx context.Context, queueName string, body []byte) error
	Receive(ctx context.Context, queueName string) (*Message, error)
	Ack(ctx context.Context, lockToken string) error
	Nack(ctx context.Context, lockToken string) error

	GetStats(ctx context.Context, queueName string) (*Stats, error)

	Close() error
}
