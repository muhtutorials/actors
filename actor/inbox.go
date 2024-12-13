package actor

import (
	"github.com/muhtutorials/actors/ring_buffer"
	"runtime"
	"sync/atomic"
)

const (
	defaultThroughput = 300
	messageBatchSize  = 1024 * 4
)

const (
	stopped int32 = iota
	starting
	idle
	running
)

type Scheduler interface {
	Schedule(func())
	Throughput() int
}

type goScheduler int

func (goScheduler) Schedule(fn func()) {
	go fn()
}

func (s goScheduler) Throughput() int {
	return int(s)
}

func NewScheduler(throughput int) Scheduler {
	return goScheduler(throughput)
}

type Inboxer interface {
	Send(Envelope)
	Start(Processor)
	Stop() error
}

type Inbox struct {
	rb         *ring_buffer.RingBuffer[Envelope]
	proc       Processor
	scheduler  Scheduler
	procStatus atomic.Int32
}

func NewInbox(size int) *Inbox {
	in := &Inbox{
		rb:        ring_buffer.New[Envelope](int64(size)),
		scheduler: NewScheduler(defaultThroughput),
	}
	in.procStatus.Store(stopped)
	return in
}

func (in *Inbox) Send(msg Envelope) {
	in.rb.Push(msg)
	in.schedule()
}

func (in *Inbox) Start(proc Processor) {
	// Transition to "starting" and then "idle" to ensure no race condition on in.proc.
	if in.procStatus.CompareAndSwap(stopped, starting) {
		in.proc = proc
		in.procStatus.Swap(idle)
		in.schedule()
	}
}

func (in *Inbox) Stop() error {
	in.procStatus.Swap(stopped)
	return nil
}

func (in *Inbox) schedule() {
	if in.procStatus.CompareAndSwap(idle, running) {
		in.scheduler.Schedule(in.process)
	}
}

func (in *Inbox) process() {
	in.run()
	in.procStatus.CompareAndSwap(running, idle)
}

func (in *Inbox) run() {
	i, t := 0, in.scheduler.Throughput()
	for in.procStatus.Load() != stopped {
		if i > t {
			i = 0
			runtime.Gosched()
		}
		i++
		if messages, ok := in.rb.PopN(messageBatchSize); ok && len(messages) > 0 {
			in.proc.Invoke(messages)
		} else {
			return
		}
	}
}
