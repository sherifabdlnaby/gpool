/* package main

import (
	"pipeline/Payload"
	"pipeline/StringReverser"
)

func main() {
    var processPlugin StringReverser.StringReverser
	processPlugin.Init("{'json' : 'config and stuff '}")
	processPlugin.Start()
	processPlugin.Run(Payload.Payload{})
	processPlugin.Close()
}*/

package main

import (
	"fmt"
	"log"
	"pipeline/Payload"
	"pipeline/pool"
)

const WORKER_COUNT = 2
const JOB_COUNT = 500

func main() {
	log.Println("starting application...")

	workerPool := pool.WorkerPool{}

	workerPool.Init(WORKER_COUNT)

	workerPool.Start()

	result1, _ := workerPool.Enqueue(Payload.Payload{ID: 123, Json: "SHERIF"})
	result2, _ := workerPool.Enqueue(Payload.Payload{ID: 124, Json: "SHERIF"})
	result3, _ := workerPool.Enqueue(Payload.Payload{ID: 125, Json: "SHERIF"})
	result4, _ := workerPool.Enqueue(Payload.Payload{ID: 126, Json: "SHERIF"})
	result5, _ := workerPool.Enqueue(Payload.Payload{ID: 127, Json: "SHERIF"})
	result6, _ := workerPool.Enqueue(Payload.Payload{ID: 128, Json: "SHERIF"})
	result7, _ := workerPool.Enqueue(Payload.Payload{ID: 129, Json: "SHERIF"})

	fmt.Println(<-result7)
	fmt.Println(<-result2)
	fmt.Println(<-result4)
	fmt.Println(<-result6)
	fmt.Println(<-result5)
	fmt.Println(<-result3)
	fmt.Println(<-result1)

	workerPool.Stop()
}
