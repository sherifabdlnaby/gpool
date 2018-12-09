package worker

import (
	"context"
)

type ConcurrencyBounder interface {
	Start()
	Stop()
	Enqueue(context.Context, func()) error
}
