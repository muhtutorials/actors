package actor

import (
	"bytes"
	"context"
	"fmt"
	"github.com/DataDog/gostackparse"
	"log/slog"
	"runtime/debug"
	"time"
)

type Envelope struct {
	Sender  *PID
	Message any
}

// Processor is an interface that abstracts the way a process behaves.
type Processor interface {
	Start()
	ShutDown()
	Send(*PID, any, *PID)
	Invoke([]Envelope)
	PID() *PID
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

func (p *process) Start() {
	p.context.receiver = p.Producer()
	defer func() {
		if v := recover(); v != nil {
			p.context.message = Stopped{}
			p.receive()
			p.tryRestart(v)
		}
	}()
	p.context.message = Initialized{}
	p.receive()
	p.context.engine.BroadcastEvent(ActorInitializedEvent{
		PID:       p.pid,
		Timestamp: time.Now(),
	})
	p.context.message = Started{}
	p.receive()
	p.context.engine.BroadcastEvent(ActorStartedEvent{
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

func (p *process) ShutDown() {
	p.cleanUp(nil)
}

func (p *process) Send(_ *PID, msg any, sender *PID) {
	p.inbox.Send(Envelope{Sender: sender, Message: msg})
}

func (p *process) Invoke(msgs []Envelope) {
	// Number of messages that need to be processed.
	nMessages := len(msgs)
	// Number of messages that have been processed.
	nProcessed := 0
	defer func() {
		// if the process recovers, all the unprocessed messages are saved to the buffer
		// so that they can be processed after the process restarts.
		if v := recover(); v != nil {
			p.context.message = Stopped{}
			p.receive()
			nUnprocessed := nMessages - nProcessed
			p.messageBuf = make([]Envelope, nUnprocessed)
			for i := 0; i < nUnprocessed; i++ {
				p.messageBuf[i] = msgs[nProcessed+i]
			}
			p.tryRestart(v)
		}
	}()
	for i := 0; i < nMessages; i++ {
		msg := msgs[i]
		if kill, ok := msg.Message.(killProcess); ok {
			// If we need to stop gracefully, we process all the messages
			// from the inbox, otherwise we ignore and clean up.
			if kill.graceful {
				unprocessed := msgs[nProcessed:]
				for _, m := range unprocessed {
					p.invokeMessage(m)
				}
			}
			p.cleanUp(kill.cancel)
			return
		}
		p.invokeMessage(msg)
		nProcessed++
	}
}

func (p *process) PID() *PID { return p.pid }

func (p *process) receive() {
	receiveFunc := p.context.receiver.Receive
	// Apply middleware if there is any.
	mwLen := len(p.Opts.MiddleWare)
	if mwLen > 0 {
		for i := mwLen - 1; i >= 0; i-- {
			receiveFunc = p.Opts.MiddleWare[i](receiveFunc)
		}
	}
	receiveFunc(p.context)
}

func (p *process) tryRestart(v any) {
	// InternalError does not take the maximum restarts into account.
	// For now, InternalError is getting triggered when we are dialing
	// a remote node. By doing this, we can keep dialing until it comes
	// back up. NOTE: not sure if that is the best option. What if that
	// node never comes back up again?
	if err, ok := v.(*InternalError); ok {
		slog.Error("internal error", "from", err.From, "err", err.Err)
		time.Sleep(p.Opts.RestartDelay)
		p.Start()
		return
	}
	// If we reach the max restarts, we shut down the inbox and clean
	// everything up.
	if p.restarts == p.MaxRestarts {
		p.context.engine.BroadcastEvent(ActorMaxRestartsExceededEvent{
			PID:       p.pid,
			Timestamp: time.Now(),
		})
		p.cleanUp(nil)
		return
	}
	stackTrace := cleanTrace(debug.Stack())
	// Restart the process after its "restartDelay".
	p.restarts++
	p.context.engine.BroadcastEvent(ActorRestartedEvent{
		PID:        p.pid,
		Timestamp:  time.Now(),
		Stacktrace: stackTrace,
		Reason:     v,
		Restarts:   p.restarts,
	})
	time.Sleep(p.Opts.RestartDelay)
	p.Start()
}

func (p *process) cleanUp(cancel context.CancelFunc) {
	defer cancel()
	if p.context.parentCtx != nil {
		_ = p.context.parentCtx.children.Delete(p.pid.ID)
	}
	if p.context.children.Len() > 0 {
		children := p.context.Children()
		for _, pid := range children {
			<-p.context.engine.Kill(pid).Done()
		}
	}
	_ = p.inbox.Stop()
	_ = p.context.engine.Processes.Delete(p.pid.ID)
	p.context.message = Stopped{}
	p.receive()
	p.context.engine.BroadcastEvent(ActorStoppedEvent{
		PID:       p.pid,
		Timestamp: time.Now(),
	})
}

func (p *process) invokeMessage(msg Envelope) {
	if _, ok := msg.Message.(killProcess); ok {
		return
	}
	p.context.sender = msg.Sender
	p.context.message = msg.Message
	p.receive()
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
