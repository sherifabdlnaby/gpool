package main

import (
	"context"
	"fmt"
	"github.com/sherifabdlnaby/gpool"
	"log"
	"time"
)

// WorkerCount Number of Workers / Concurrent jobs of the Pool
const WorkerCount = 2

func main() {
	var workerPool gpool.Pool

	//workerPool = workerpooldispatch.NewWorkerPool(WORKER_COUNT)
	workerPool = gpool.NewSemaphorePool(WorkerCount)

	log.Println("Starting Pool...")

	workerPool.Start()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for i := 0; i < 10; i++ {

			// Small Interval for more readable output
			time.Sleep(500 * time.Millisecond)

			go func(i int) {
				x := make(chan int, 1)

				log.Printf("Job [%v] Enqueueing", i)

				err := workerPool.Enqueue(ctx, func() {
					time.Sleep(2000 * time.Millisecond)
					x <- i
				})

				if err != nil {
					log.Printf("Job [%v] was not enqueued. [%s]", i, err.Error())
					return
				}

				log.Printf("Job [%v] Enqueue-ed ", i)

				log.Printf("Job [%v] Receieved [%v]", i, <-x)
			}(i)
		}
	}()

	// Uncomment to demonstrate ctx cancel of jobs.
	//time.Sleep(100 * time.Millisecond)
	//cancel()

	time.Sleep(5000 * time.Millisecond)

	fmt.Println("Stopping...")

	workerPool.Stop()

	fmt.Println("Stopped")

	fmt.Println("Sleeping for couple of seconds so canceled job have a chance to print out their status")

	time.Sleep(10000 * time.Millisecond)
}
