package WorkerPool

import "pipeline/Payload"

type Task interface {
	Run(p Payload.Payload) Payload.Payload
}
