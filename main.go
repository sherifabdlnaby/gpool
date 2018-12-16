package main

import (
	"context"
	"fmt"
	"gpool/gpool"
	"log"
	"time"
)

type stringHasher struct {
}

func (sh *stringHasher) Run(s int) int {

	//log.Println(fmt.Sprintf("RUNNING A PROCESS AT Plugin: [%s] [v%s] with PAYLOAD: %s", s.GetIdentifier().Name, s.GetIdentifier().Version, p))
	//time.Sleep(1000 * time.Millisecond)

	return s
}

const WORKER_COUNT = 1

var sr = stringHasher{}

func main() {
	var workerPool gpool.Pool
	//workerPool = workerpooldispatch.NewWorkerPool(WORKER_COUNT)
	workerPool = gpool.NewSemaphorePool(WORKER_COUNT)
	workerPool.St()
	ctx, _ := context.WithCancel(context.Background())
	//workerPool.Stop()

	//cancel()
	go func() {
		for i := 0; i < 10; i++ {
			go func(i int) {
				x := make(chan int, 1)
				fmt.Println("ENQUEUED")
				err := workerPool.Enqueue(ctx, func() {
					time.Sleep(100 * time.Millisecond)
					x <- sr.Run(i)
				})
				if err != nil {
					log.Println(err.Error())
					return
				}
				fmt.Println(<-x)
			}(i)
		}
	}()

	time.Sleep(100 * time.Millisecond)
	//cancel()

	fmt.Println("Stopping")

	workerPool.Stop()

	fmt.Println("DONE")

	time.Sleep(50000 * time.Millisecond)
}
