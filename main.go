/* package main

import (
	"pipeline/payload"
	"pipeline/StringReverser"
)

func main() {
    var processPlugin StringReverser.StringReverser
	processPlugin.Init("{'json' : 'config and stuff '}")
	processPlugin.Start()
	processPlugin.Run(payload.payload{})
	processPlugin.Close()
}*/

package main

import (
	"fmt"
	"log"
	"pipeline/Payload"
	"pipeline/StringReverser"
	"pipeline/workerpool"
	"time"
)

const WORKER_COUNT = 3

func main() {
	log.Println("starting application...")

	workerPool := workerpool.NewWorkerPool(WORKER_COUNT)

	workerPool.Start()
	defer workerPool.Stop()

	sr := StringReverser.StringReverser{}

	/*	result0, _ := workerPool.EnqueueWithTimeout(sr, payload.payload{ID: 123, Json: "SHERIF"}, time.Second * 1000)
		result1, _ := workerPool.EnqueueWithTimeout(sr, <-result0, time.Second * 1000)
		result2, _ := workerPool.EnqueueWithTimeout(sr, <-result1, time.Second * 1000)
		result3, _ := workerPool.EnqueueWithTimeout(sr, <-result2, time.Second * 1000)
		result4, _ := workerPool.EnqueueWithTimeout(sr, <-result3, time.Second * 1000)
		result5, _ := workerPool.EnqueueWithTimeout(sr, <-result4, time.Second * 1000)
		result6, _ := workerPool.EnqueueWithTimeout(sr, <-result5, time.Second * 1000)
		result7, _ := workerPool.EnqueueWithTimeout(sr, <-result6, time.Second * 1000)
		result8, _ := workerPool.EnqueueWithTimeout(sr, <-result7, time.Second * 1000)
		result9, _ := workerPool.EnqueueWithTimeout(sr, <-result8, time.Second * 1000)
		result10, _ := workerPool.EnqueueWithTimeout(sr, <-result9, time.Second * 1000)
		workerPool.EnqueueWithTimeout(sr, <-result10, time.Second * 1000)*/

	result1, _ := workerPool.EnqueueWithTimeout(sr, Payload.Payload{ID: 1, Json: "SHERIF"}, time.Millisecond*1000)
	result2, _ := workerPool.EnqueueWithTimeout(sr, Payload.Payload{ID: 2, Json: "SHERIF"}, time.Millisecond*1000)
	result3, _ := workerPool.EnqueueWithTimeout(sr, Payload.Payload{ID: 3, Json: "SHERIF"}, time.Millisecond*1000)
	result4, _ := workerPool.EnqueueWithTimeout(sr, Payload.Payload{ID: 4, Json: "SHERIF"}, time.Millisecond*1000)
	result5, _ := workerPool.EnqueueWithTimeout(sr, Payload.Payload{ID: 5, Json: "SHERIF"}, time.Millisecond*1000)
	result6, _ := workerPool.EnqueueWithTimeout(sr, Payload.Payload{ID: 6, Json: "SHERIF"}, time.Millisecond*1000)
	result7, _ := workerPool.EnqueueWithTimeout(sr, Payload.Payload{ID: 7, Json: "SHERIF"}, time.Millisecond*1000)

	fmt.Println(<-result7)
	fmt.Println(<-result2)
	fmt.Println(<-result4)
	fmt.Println(<-result6)
	fmt.Println(<-result5)
	fmt.Println(<-result3)
	fmt.Println(<-result1)

}
