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
	"log"
	"pipeline/pool"
	"pipeline/work"
	"time"
)

const WORKER_COUNT = 2
const JOB_COUNT = 5

func main() {
	log.Println("starting application...")

	workerPool := pool.WorkerPool{}

	workerPool.StartWorkerPool(WORKER_COUNT)

	for i, job := range work.CreateJobs(JOB_COUNT) {
		workerPool.Work <- pool.Work{Job: job, ID: i}
	}

	workerPool.StopWorkers()

	time.Sleep(1000 * time.Second)
}
