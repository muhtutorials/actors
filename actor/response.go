package actor

import (
	"context"
	"math"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

type Response struct {
	pid     *PID
	engine  *Engine
	result  chan any
	timeout time.Duration
}

func NewResponse(e *Engine, timeout time.Duration) *Response {
	return &Response{
		pid:     NewPID(e.address, "response"+pidSep+strconv.Itoa(rand.Intn(math.MaxInt32))),
		engine:  e,
		result:  make(chan any, 1),
		timeout: timeout,
	}
}

func (r *Response) Result() (any, error) {
	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer func() {
		cancel()
		r.engine.Registry.Remove(r.pid)
	}()
	select {
	case resp := <-r.result:
		return resp, nil
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

func (r *Response) PID() *PID {
	return r.pid
}

func (r *Response) Send(_ *PID, msg any, _ *PID) {
	r.result <- msg
}

func (r *Response) Invoke([]Envelope) {}

func (r *Response) Start() {}

func (r *Response) ShutDown(_ *sync.WaitGroup) {}
