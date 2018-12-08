package main

import (
	"fmt"
	"pipeline/semp"
	"pipeline/work"
	"sync"
)

type stringHasher struct {
}

func (sh *stringHasher) Run(s int) int {

	//log.Println(fmt.Sprintf("RUNNING A PROCESS AT Plugin: [%s] [v%s] with PAYLOAD: %s", s.GetIdentifier().Name, s.GetIdentifier().Version, p))
	//time.Sleep(1000 * time.Millisecond)

	return s
}

const WORKER_COUNT = 5

var sr = stringHasher{}

func main() {

	var workerPool worker.ConcurrencyBounder
	//workerPool = workerpooldispatch.NewWorkerPool(WORKER_COUNT)
	workerPool = semp.NewSempWorker(WORKER_COUNT)
	workerPool.Start()
	wg := sync.WaitGroup{}

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			x, _ := workerPool.Enqueue(&sr, 3)
			fmt.Println(<-x)
		}(i)
	}
	wg.Wait()
	workerPool.Stop()
}
