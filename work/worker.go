package dpool

import (
	"context"
	"errors"
)

type Pool interface {
	Start()
	Stop()
	Enqueue(context.Context, func()) error
	TryEnqueue(func()) bool
}

var (
	ErrPoolClosed = errors.New("pool is closed")
	ErrJobTimeout = errors.New("job canceled")
)
