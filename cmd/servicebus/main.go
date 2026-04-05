package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/RyonMcAuley/servicebus/internal/store"
	"github.com/RyonMcAuley/servicebus/internal/transport"
)

func main() {
	store, err := store.NewSqliteStore("./db/servicebus.db")
	if err != nil {
		fmt.Println("Error connecting to database")
		panic(err)
	}
	defer store.Close()

	r := transport.NewRouter(store)
	log.Fatal(http.ListenAndServe(":5800", r))
}

// Print stats for all q's
func getStats(store *store.SqliteStore) {
	fmt.Println("Getting stats...")
	// queueName := "test-queue"

	allQueues, err := store.ListQueues(context.Background())
	if err != nil {
		panic(err)
	}

	for _, queue := range allQueues {
		stats, err := store.GetStats(context.Background(), queue.Name)
		if err != nil {
			panic(err)
		}

		fmt.Println("Queue: " + stats.QueueName)
		fmt.Println("Active: ", stats.MessageCount)
		fmt.Println("DLQ: ", stats.DLQCount)
	}
}

// Create test-queue and list queues
func initQueue(store *store.SqliteStore) {
	queueName := "test-queue-2"
	err := store.CreateQueue(context.Background(), queueName, 5)
	if err != nil {
		fmt.Println("Queue already exists: " + queueName)
	}
	// queues, err := store.ListQueues(context.Background())
	// if err != nil {
	// 	panic(err)
	// }
	// printQueues(queues)
}

// Add something to the queue & peek it
func initData(store *store.SqliteStore) {
	queueName := "test-queue-2"
	err := store.Enqueue(context.Background(), queueName, []byte("test"))
	if err != nil {
		fmt.Println("Error queuing message")
		panic(err)
	}

	// test, err := store.Peek(context.Background(), queueName)
	// fmt.Println(string(test.Body))
}

// Print names in the queue array
func printQueues(qs []*store.Queue) {
	fmt.Println("All existing queues: ")
	for _, q := range qs {
		fmt.Println("- " + q.Name)
	}
	fmt.Println()
}
