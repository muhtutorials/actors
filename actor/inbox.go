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

type Inboxer interface {
	Start(Processor)
	Stop() error
	Send(Envelope)
}

type Inbox struct {
	buf        *ring_buffer.RingBuffer[Envelope]
	proc       Processor
	scheduler  Scheduler
	procStatus atomic.Int32
}

func NewInbox(size int) *Inbox {
	inbox := &Inbox{
		buf:       ring_buffer.New[Envelope](int64(size)),
		scheduler: NewScheduler(defaultThroughput),
	}
	inbox.procStatus.Store(stopped)
	return inbox
}

func (in *Inbox) Start(proc Processor) {
	// Transition to "starting" and then "idle" to ensure no race condition on "in.proc".
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

func (in *Inbox) Send(msg Envelope) {
	in.buf.Push(msg)
	in.schedule()
}

func (in *Inbox) schedule() {
	if in.procStatus.CompareAndSwap(idle, running) {
		in.scheduler.Schedule(in.process)
	}
}

func (in *Inbox) process() {
	in.run()
	if in.procStatus.CompareAndSwap(running, idle) && in.buf.Len() > 0 {
		// Messages might have been added to the ring buffer
		// between the last pop and the transition to "idle".
		// if this is the case, then we should schedule again.
		in.schedule()
	}
}

func (in *Inbox) run() {
	i, throughput := 0, in.scheduler.Throughput()
	for in.procStatus.Load() != stopped {
		if i > throughput {
			i = 0
			runtime.Gosched()
		}
		i++
		if messages, ok := in.buf.PopN(messageBatchSize); ok {
			in.proc.Invoke(messages)
		} else {
			return
		}
	}
}

type Scheduler interface {
	Schedule(func())
	Throughput() int
}

type goScheduler int

func NewScheduler(throughput int) Scheduler {
	return goScheduler(throughput)
}

func (goScheduler) Schedule(fn func()) {
	go fn()
}

func (s goScheduler) Throughput() int {
	return int(s)
}
