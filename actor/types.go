package actor

import (
	"context"
	"sync"
)

// Remoter abstracts a remote that's tied to an engine.
type Remoter interface {
	Start(*Engine) error
	Stop() *sync.WaitGroup
	Send(*PID, any, *PID)
	Address() string
}

// Producer is a function that returns a Receiver.
type Producer func() Receiver

// Receiver (actor) receives and processes messages.
type Receiver interface {
	Receive(*Context)
}

// funcReceiver is used to turn a stateless actor into a producer.
// "func(*Context)" becomes "Producer".
// Its usage can be seen inside "Engine.SpawnFunc" method.
type funcReceiver struct {
	fn func(*Context)
}

func newFuncReceiver(fn func(ctx *Context)) Producer {
	return func() Receiver {
		return &funcReceiver{
			fn: fn,
		}
	}
}

func (r *funcReceiver) Receive(ctx *Context) {
	r.fn(ctx)
}

type (
	Initialized struct{}

	Started struct{}

	Stopped struct{}

	killProcess struct {
		cancel   context.CancelFunc
		graceful bool
	}

	InternalError struct {
		From string
		Err  error
	}
)
