package store

import "time"

type Queue struct {
	Name        string
	MaxDelivery int
	CreatedAt   time.Time
}

type Message struct {
	ID            string
	QueueName     string
	Body          []byte
	EnqueuedAt    time.Time
	VisibleAt     time.Time
	DeliveryCount int
	LockToken     *string
	LockedUntil   *time.Time
	IsDLQ         bool
}

type Stats struct {
	QueueName         string
	ActiveMessages    int
	AvailableMessages int
	DLQCount          int
}
