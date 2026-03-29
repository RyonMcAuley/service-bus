package store

import (
	"context"
	"database/sql"
	"github.com/google/uuid"
	_ "modernc.org/sqlite"
	"time"
)

type SqliteStore struct {
	db *sql.DB
}

func NewSqliteStore(path string) (*SqliteStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.Exec("PRAGMA journal_mode=WAL;")
	db.Exec("PRAGMA busy_timeout=5000;")

	s := &SqliteStore{db: db}

	if err := s.migrate(); err != nil {
		return nil, err
	}

	return s, nil
}

func (s *SqliteStore) Close() error {
	return s.db.Close()
}

func (s *SqliteStore) CreateQueue(ctx context.Context, name string, maxDelivery int) error {
	_, err := s.db.ExecContext(ctx, queryCreateQueue, name, maxDelivery, time.Now())
	return err
}

func (s *SqliteStore) ListQueues(ctx context.Context) ([]*Queue, error) {
	rows, err := s.db.QueryContext(ctx, queryListQueues)

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var queues []*Queue

	for rows.Next() {
		q := &Queue{}
		err := rows.Scan(&q.Name, &q.MaxDelivery, &q.CreatedAt)
		if err != nil {
			return nil, err
		}
		queues = append(queues, q)
	}
	return queues, rows.Err()
}

func (s *SqliteStore) Enqueue(ctx context.Context, queueName string, body []byte) error {
	_, err := s.db.ExecContext(ctx, queryEnqueue,
		uuid.NewString(),
		queueName,
		body,
		time.Now(), // EnqueuedAt
		time.Now(), // VisibleAt
		0,          // DeliveryCount
		false,      // IsDLQ
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *SqliteStore) Peek(ctx context.Context, qName string) (*Message, error) {
	m := s.db.QueryRowContext(ctx, queryPeek, qName, time.Now())

	msg := &Message{}
	err := m.Scan(&msg.ID, &msg.QueueName, &msg.Body, &msg.EnqueuedAt,
		&msg.VisibleAt, &msg.DeliveryCount, &msg.LockToken, &msg.LockedUntil, &msg.IsDLQ)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return msg, nil
}

func (s *SqliteStore) GetStats(ctx context.Context, qName string) (*Stats, error) {
	count := s.db.QueryRowContext(ctx, queryGetStats, qName)

	stats := &Stats{
		QueueName: qName,
	}
	err := count.Scan(&stats.MessageCount, &stats.DLQCount)
	if err != nil {
		return nil, err
	}
	return stats, nil
}
