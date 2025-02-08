package actor

import (
	"context"
	"github.com/muhtutorials/actors/safe_map"
	"log/slog"
	"math"
	"math/rand"
	"strconv"
	"time"
)

type Context struct {
	pid      *PID
	sender   *PID
	engine   *Engine
	receiver Receiver
	message  any
	context  context.Context
	// The context of the parent if we are a child.
	// We need this parentCtx, so we can remove the child from the parent context
	// when the child dies.
	parentCtx *Context
	children  *safe_map.SafeMap[string, *PID]
}

func NewContext(pid *PID, e *Engine, ctx context.Context) *Context {
	return &Context{
		pid:      pid,
		engine:   e,
		context:  ctx,
		children: safe_map.New[string, *PID](),
	}
}

// Context returns a context.Context, user defined on spawn or
// a context.Background as default.
func (c *Context) Context() context.Context {
	return c.context
}

func (c *Context) Receiver() Receiver {
	return c.receiver
}

func (c *Context) Request(pid *PID, msg any, timeout time.Duration) *Response {
	return c.engine.Request(pid, msg, timeout)
}

// Respond will send message to the sender of the current received message.
func (c *Context) Respond(msg any) {
	if c.sender == nil {
		slog.Warn("context has no sender", "method", "actor.Context.Respond", "pid", c.PID())
		return
	}
	c.engine.Send(c.sender, msg)
}

// SpawnChild will spawn the Producer as a child of the current Context.
// If the parent process dies, all the children will be automatically shut down gracefully.
// Hence, all children will receive the Stopped message.
func (c *Context) SpawnChild(p Producer, name string, optFns ...OptFunc) *PID {
	opts := DefaultOpts(p)
	opts.Kind = c.PID().ID + pidSep + name
	for _, fn := range optFns {
		fn(&opts)
	}
	// check if we got an ID, generate otherwise
	if opts.ID == "" {
		opts.ID = strconv.Itoa(rand.Intn(math.MaxInt))
	}
	proc := newProcess(c.engine, opts)
	proc.context.parentCtx = c
	pid := c.engine.SpawnProcess(proc)
	c.children.Insert(pid.ID, pid)
	return proc.PID()
}

// SpawnChildFunc spawns the function as a child Receiver of the current Context.
func (c *Context) SpawnChildFunc(fn func(*Context), name string, optFns ...OptFunc) *PID {
	return c.SpawnChild(newFuncReceiver(fn), name, optFns...)
}

// Send will send a message to the PID. This will also set the sender of the message to
// the PID of the current Context. Hence, the receiver of the message can call "Sender" to know
// the PID of the process that sent this message.
func (c *Context) Send(pid *PID, msg any) {
	c.engine.SendWithSender(pid, msg, c.pid)
}

// SendAtInterval will send a message to a PID at a provided interval.
// It will return a "SenderAtInterval" struct that can stop the repeated
// sending of a message by calling "Stop".
func (c *Context) SendAtInterval(pid *PID, msg any, interval time.Duration) SenderAtInterval {
	r := SenderAtInterval{
		self:     c.pid,
		target:   pid.CloneVT(),
		engine:   c.engine,
		message:  msg,
		interval: interval,
		cancelCh: make(chan struct{}, 1),
	}
	r.Start()
	return r
}

// Forward will forward the current received message to the PID.
// This will also set the "forwarder" as the sender of the message.
func (c *Context) Forward(pid *PID) {
	c.engine.SendWithSender(pid, c.message, c.pid)
}

// GetProcPID returns the PID of the process found by the id.
// Returns nil when it could not find any process.
func (c *Context) GetProcPID(id string) *PID {
	proc := c.engine.Registry.GetByID(id)
	if proc != nil {
		return proc.PID()
	}
	return nil
}

// PID returns the PID of the process that belongs to the context.
func (c *Context) PID() *PID {
	return c.pid
}

// Sender returns the PID of the process, when available, that sent the
// current received message.
func (c *Context) Sender() *PID {
	return c.sender
}

// Engine returns a pointer to the underlying Engine.
func (c *Context) Engine() *Engine {
	return c.engine
}

// Message returns the message that is currently being received.
func (c *Context) Message() any {
	return c.message
}

// Parent returns the PID of the process that spawned the current process.
func (c *Context) Parent() *PID {
	if c.parentCtx != nil {
		return c.parentCtx.pid
	}
	return nil
}

// Children returns all child PIDs for the current process.
func (c *Context) Children() []*PID {
	pids := make([]*PID, c.children.Len())
	i := 0
	c.children.ForEach(func(_ string, pid *PID) {
		pids[i] = pid
		i++
	})
	return pids
}
