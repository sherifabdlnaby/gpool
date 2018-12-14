package main

import (
	"context"
	"fmt"
	"log"
	"pipeline/work"
	"time"
)

type stringHasher struct {
}

func (sh *stringHasher) Run(s int) int {

	//log.Println(fmt.Sprintf("RUNNING A PROCESS AT Plugin: [%s] [v%s] with PAYLOAD: %s", s.GetIdentifier().Name, s.GetIdentifier().Version, p))
	//time.Sleep(1000 * time.Millisecond)

	return s
}

const WORKER_COUNT = 0

var sr = stringHasher{}

func main() {
	var workerPool dpool.Pool
	//workerPool = workerpooldispatch.NewWorkerPool(WORKER_COUNT)
	workerPool = dpool.NewSemaphorePool(WORKER_COUNT)
	workerPool.Start()
	ctx, cancel := context.WithCancel(context.Background())
	//workerPool.Stop()
	//workerPool.Stop()
	cancel()
	time.Sleep(3000 * time.Millisecond)
	go func() {
		for i := 0; i < 10; i++ {
			go func(i int) {
				x := make(chan int, 1)
				err := workerPool.Enqueue(ctx, func() {
					x <- sr.Run(i)
				})
				if err != nil {
					log.Println(err.Error())
					return
				}
				fmt.Print(<-x)
				fmt.Println("LOL")
			}(i)
		}
	}()

	time.Sleep(4000 * time.Millisecond)
	//cancel()

	fmt.Println("DONE")
	time.Sleep(50000 * time.Millisecond)
}
