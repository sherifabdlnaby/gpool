package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"pipeline/work"
	"pipeline/workerpool"
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

	helloHandler := func(w http.ResponseWriter, req *http.Request) {
		panic("adsf")
		io.WriteString(w, "Hello, world!\n")
	}

	http.HandleFunc("/hello", helloHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))

	var workerPool worker.ConcurrencyBounder
	//workerPool = workerpooldispatch.NewWorkerPool(WORKER_COUNT)
	workerPool = workerpool.NewWorkerPool(WORKER_COUNT)
	workerPool.Start()
	ctx, _ := context.WithCancel(context.Background())
	//workerPool.Stop()
	workerPool.Stop()

	go func() {
		for i := 0; i < 10; i++ {
			go func(i int) {
				x := make(chan int, 1)
				err := workerPool.Enqueue(ctx, func() {
					panic("lol")
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
