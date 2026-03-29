package main

import (
	"context"
	"fmt"

	"github.com/RyonMcAuley/servicebus/internal/store"
)

func main() {
	fmt.Println("Starting up...")
	store, err := store.NewSqliteStore("./db/servicebus.db")
	if err != nil {
		panic(err)
	}
	fmt.Println("DB connection established")

	initQueue(store)
	initData(store)

	store.Close()
}

// Create test-queue and list queues
func initQueue(store *store.SqliteStore) {
	fmt.Println("Creating queue...")
	queueName := "test-queue"
	err := store.CreateQueue(context.Background(), queueName, 5)
	if err != nil {
		fmt.Println("Queue already created: " + queueName)
	} else {
		fmt.Println("New queue created: " + queueName)
	}
	queues, err := store.ListQueues(context.Background())
	if err != nil {
		panic(err)
	}
	printQueues(queues)
}

// Add something to the queue & peek it
func initData(store *store.SqliteStore) {
	fmt.Println("Queueing message...")
	queueName := "test-queue"
	err := store.Enqueue(context.Background(), queueName, []byte("test"))
	if err != nil {
		panic(err)
	}
	fmt.Println("Message queued")

	test, err := store.Peek(context.Background(), queueName)
	fmt.Println(string(test.Body))
}

// Print names in the queue array
func printQueues(qs []*store.Queue) {
	fmt.Println("All existing queues: ")
	for _, q := range qs {
		fmt.Println("- " + q.Name)
	}
	fmt.Println()
}
