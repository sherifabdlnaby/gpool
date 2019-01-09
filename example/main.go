package main

import (
	"context"
	"fmt"
	"github.com/sherifabdlnaby/gpool"
	"log"
	"time"
)

// size Workers / Concurrent jobs of the Pool
const size = 2

func main() {
	var pool gpool.Pool
	pool, _ = gpool.NewSemaphorePool(size)
	log.Println("Starting Pool...")

	pool.Start()
	defer pool.Stop()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		for i := 0; i < 10; i++ {

			// Small Interval for more readable output
			time.Sleep(500 * time.Millisecond)

			go func(i int) {
				x := make(chan int, 1)

				log.Printf("Job [%v] Enqueueing", i)

				err := pool.Enqueue(ctx, func() {
					time.Sleep(2000 * time.Millisecond)
					x <- i
				})

				if err != nil {
					log.Printf("Job [%v] was not enqueued. [%s]", i, err.Error())
					return
				}

				log.Printf("Job [%v] Enqueue-ed ", i)

				log.Printf("Job [%v] Receieved, Result: [%v]", i, <-x)
			}(i)
		}
	}()

	// Uncomment to demonstrate ctx cancel of jobs.
	//time.Sleep(100 * time.Millisecond)
	//cancel()

	time.Sleep(5000 * time.Millisecond)

	fmt.Println("Stopping...")

	pool.Stop()

	fmt.Println("Stopped")

	fmt.Println("Sleeping for couple of seconds so canceled job have a chance to print out their status")

	time.Sleep(4000 * time.Millisecond)
}
