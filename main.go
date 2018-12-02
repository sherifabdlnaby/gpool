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
	"pipeline/pool"
	"pipeline/work"
	"time"
)

const WORKER_COUNT = 2
const JOB_COUNT = 500

func main() {
	log.Println("starting application...")

	workerPool := pool.WorkerPool{}

	workerPool.StartWorkerPool(WORKER_COUNT)

	for i, job := range work.CreateJobs(JOB_COUNT) {
		select {
		case workerPool.Work <- pool.Work{Job: job, ID: i}:
		case <-time.After(1 * time.Second):
			log.Println(fmt.Sprintf("Job [%d] Timedout", i))
		}

	}

	workerPool.Stop()
}
