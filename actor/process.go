package actor

import (
	"bytes"
	"fmt"
	"github.com/DataDog/gostackparse"
	"log/slog"
	"runtime/debug"
	"sync"
	"time"
)

type Envelope struct {
	Sender  *PID
	Message any
}

// Processor is an interface that abstracts the way a process behaves.
type Processor interface {
	PID() *PID
	Send(*PID, any, *PID)
	Invoke([]Envelope)
	Start()
	ShutDown(*sync.WaitGroup)
}

type process struct {
	pid        *PID
	context    *Context
	inbox      Inboxer
	messageBuf []Envelope
	restarts   int32
	Opts
}

func newProcess(e *Engine, opts Opts) *process {
	pid := NewPID(e.address, opts.Kind+pidSep+opts.ID)
	ctx := NewContext(pid, e, opts.Context)
	return &process{
		pid:     pid,
		context: ctx,
		inbox:   NewInbox(opts.InboxSize),
		Opts:    opts,
	}
}

func applyMiddleware(receiveFunc ReceiveFunc, middleware ...MiddlewareFunc) ReceiveFunc {
	for i := len(middleware) - 1; i >= 0; i-- {
		receiveFunc = middleware[i](receiveFunc)
	}
	return receiveFunc
}

func (p *process) PID() *PID { return p.pid }

func (p *process) Send(_ *PID, msg any, sender *PID) {
	p.inbox.Send(Envelope{Sender: sender, Message: msg})
}

func (p *process) Invoke(msgs []Envelope) {
	// Number of messages that need to be processed.
	nMessages := len(msgs)
	// Number of messages that are processed.
	nProcessed := 0
	// FIXME: We could use nProcessed here, but for some reason placing nProcessed++ on the
	// bottom of the function freezes some tests. Hence, I created a new counter
	// for bookkeeping.
	processed := 0
	defer func() {
		// If we recovered, we buffer up all the messages that we could not process
		// so we can retry them on the next restart.
		if v := recover(); v != nil {
			p.context.message = Stopped{}
			p.context.receiver.Receive(p.context)
			p.messageBuf = make([]Envelope, nMessages-nProcessed)
			for i := 0; i < nMessages-nProcessed; i++ {
				p.messageBuf[i] = msgs[i+nProcessed]
			}
			p.tryRestart(v)
		}
	}()
	for i := 0; i < nMessages; i++ {
		nProcessed++
		msg := msgs[i]
		if pill, ok := msg.Message.(poisonPill); ok {
			// If we need to stop gracefully, we process all the messages
			// from the inbox, otherwise we ignore and clean up.
			if pill.graceful {
				unprocessed := msgs[processed:]
				for _, m := range unprocessed {
					p.invokeMessage(m)
				}
			}
			p.cleanUp(pill.wg)
			return
		}
		p.invokeMessage(msg)
		processed++
	}
}

func (p *process) invokeMessage(msg Envelope) {
	// Suppress poison pill messages here. they're private to the actor engine.
	if _, ok := msg.Message.(poisonPill); ok {
		return
	}
	p.context.sender = msg.Sender
	p.context.message = msg.Message
	rcv := p.context.receiver
	if len(p.Opts.MiddleWare) > 0 {
		applyMiddleware(rcv.Receive, p.Opts.MiddleWare...)(p.context)
	} else {
		rcv.Receive(p.context)
	}
}

func (p *process) Start() {
	rcv := p.Producer()
	p.context.receiver = rcv
	defer func() {
		if v := recover(); v != nil {
			p.context.message = Stopped{}
			p.context.receiver.Receive(p.context)
			p.tryRestart(v)
		}
	}()
	p.context.message = Initialized{}
	applyMiddleware(rcv.Receive, p.Opts.MiddleWare...)(p.context)
	p.context.engine.BroadcastEvent(EventActorInitialized{
		PID:       p.pid,
		Timestamp: time.Now(),
	})
	p.context.message = Started{}
	applyMiddleware(rcv.Receive, p.Opts.MiddleWare...)(p.context)
	p.context.engine.BroadcastEvent(EventActorStarted{
		PID:       p.pid,
		Timestamp: time.Now(),
	})
	// If we have messages in our buffer, invoke them.
	if len(p.messageBuf) > 0 {
		p.Invoke(p.messageBuf)
		p.messageBuf = nil
	}
	p.inbox.Start(p)
}

func (p *process) tryRestart(v any) {
	// InternalError does not take the maximum restarts into account.
	// For now, InternalError is getting triggered when we are dialing
	// a remote node. By doing this, we can keep dialing until it comes
	// back up. NOTE: not sure if that is the best option. What if that
	// node never comes back up again?
	if err, ok := v.(*InternalError); ok {
		slog.Error(err.From, "err", err.Err)
		time.Sleep(p.Opts.RestartDelay)
		p.Start()
		return
	}
	stackTrace := cleanTrace(debug.Stack())
	// If we reach the max restarts, we shut down the inbox and clean
	// everything up.
	if p.restarts == p.MaxRestarts {
		p.context.engine.BroadcastEvent(EventActorMaxRestartsExceeded{
			PID:       p.pid,
			Timestamp: time.Now(),
		})
		p.cleanUp(nil)
		return
	}
	p.restarts++
	// Restart the process after its restartDelay.
	p.context.engine.BroadcastEvent(EventActorRestarted{
		PID:        p.pid,
		Timestamp:  time.Now(),
		Stacktrace: stackTrace,
		Reason:     v,
		Restarts:   p.restarts,
	})
	time.Sleep(p.Opts.RestartDelay)
	p.Start()
}

func (p *process) cleanUp(wg *sync.WaitGroup) {
	if p.context.parentCtx != nil {
		_ = p.context.parentCtx.children.Delete(p.pid.ID)
	}
	if p.context.children.Len() > 0 {
		children := p.context.Children()
		for _, pid := range children {
			p.context.engine.Poison(pid).Wait()
		}
	}
	_ = p.inbox.Stop()
	p.context.engine.registry.Remove(p.pid)
	p.context.message = Stopped{}
	applyMiddleware(p.context.receiver.Receive, p.Opts.MiddleWare...)(p.context)
	p.context.engine.BroadcastEvent(EventActorStopped{
		PID: p.pid, Timestamp: time.Now(),
	})
	if wg != nil {
		wg.Done()
	}
}

func (p *process) ShutDown(wg *sync.WaitGroup) {
	p.cleanUp(wg)
}

func cleanTrace(stack []byte) []byte {
	goroutines, err := gostackparse.Parse(bytes.NewReader(stack))
	if err != nil {
		slog.Error("failed to parse stack trace")
		return stack
	}
	if len(goroutines) != 1 {
		slog.Error("expected only one goroutine", "goroutines", len(goroutines))
		return stack
	}
	// Skip the first frames.
	goroutines[0].Stack = goroutines[0].Stack[4:]
	buf := bytes.NewBuffer(nil)
	_, _ = fmt.Fprintf(buf, "goroutine %d [%s]\n", goroutines[0].ID, goroutines[0].State)
	for _, frame := range goroutines[0].Stack {
		_, _ = fmt.Fprintf(buf, "%s\n", frame.Func)
		_, _ = fmt.Fprintf(buf, "\t%s:%d\n", frame.File, frame.Line)
	}
	return buf.Bytes()
}
