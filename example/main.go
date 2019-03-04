package main

import (
	"context"
	"github.com/sherifabdlnaby/gpool"
	"log"
	"time"
)

// size Workers / Concurrent jobs of the Pool
const size = 2

func main() {
	concurrency := 2

	// Create and start pool.
	pool, _ := gpool.NewPool(concurrency)
	pool.Start()
	defer pool.Stop()

	// Create JOB
	resultChan1 := make(chan int)
	ctx := context.Background()
	job := func() {
		time.Sleep(2000 * time.Millisecond)
		resultChan1 <- 1337
	}

	// Enqueue Job
	err1 := pool.Enqueue(ctx, job)

	if err1 != nil {
		log.Printf("Job was not enqueued. Error: [%s]", err1.Error())
		return
	}

	log.Printf("Job Enqueued and started processing")

	log.Printf("Job Done, Received: %v", <-resultChan1)
}
