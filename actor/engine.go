package actor

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

// Remoter abstracts a remote that's tied to an engine.
type Remoter interface {
	Start(*Engine) error
	Stop() *sync.WaitGroup
	Send(*PID, any, *PID)
	Address() string
}

// Producer is a function that can return a Receiver.
type Producer func() Receiver

// Receiver (actor) receives and processes messages.
type Receiver interface {
	Receive(*Context)
}

// Engine represents the actor engine.
type Engine struct {
	Registry    *Registry
	address     string
	remote      Remoter
	eventStream *PID
}

// EngineConfig holds the configuration of the engine.
type EngineConfig struct {
	remote Remoter
}

// NewEngineConfig returns a new default EngineConfig.
func NewEngineConfig() EngineConfig {
	return EngineConfig{}
}

// WithRemote sets the remote which will configure the engine so it's capable
// to send and receive messages over the network.
func (cfg EngineConfig) WithRemote(r Remoter) EngineConfig {
	cfg.remote = r
	return cfg
}

// NewEngine returns a new actor Engine given an EngineConfig.
func NewEngine(cfg EngineConfig) (*Engine, error) {
	e := new(Engine)
	e.Registry = NewRegistry(e) // need to init the registry in case we want a custom dead letter
	e.address = LocalLookupAddr
	if cfg.remote != nil {
		e.remote = cfg.remote
		e.address = cfg.remote.Address()
		if err := cfg.remote.Start(e); err != nil {
			return nil, fmt.Errorf("failed to start remote: %w", err)
		}
	}
	e.eventStream = e.Spawn(newEventStream(), "event_stream")
	return e, nil
}

// Spawn spawns a process.
func (e *Engine) Spawn(p Producer, kind string, optFns ...OptFunc) *PID {
	opts := DefaultOpts(p)
	opts.Kind = kind
	for _, fn := range optFns {
		fn(&opts)
	}
	// check if we got an ID, generate otherwise
	if opts.ID == "" {
		opts.ID = strconv.Itoa(rand.Intn(math.MaxInt))
	}
	proc := newProcess(e, opts)
	return e.SpawnProcess(proc)
}

// SpawnFunc spawns the given function as a stateless receiver/actor.
func (e *Engine) SpawnFunc(fn func(*Context), kind string, optFns ...OptFunc) *PID {
	return e.Spawn(newFuncReceiver(fn), kind, optFns...)
}

// SpawnProcess spawns the given Processor. This function is useful when working
// with custom created processes. Take a look at the streamWriter as an example.
func (e *Engine) SpawnProcess(p Processor) *PID {
	e.Registry.Add(p)
	return p.PID()
}

// Address returns the address of the actor engine. When there is
// no remote configured, the "local" address will be used, otherwise
// the listen address of the remote.
func (e *Engine) Address() string {
	return e.address
}

// Request sends the message to the given PID as a "Request", returning
// a response that will resolve in the future. Calling "Response.Result" will
// block until the deadline is exceeded or the response is being resolved.
func (e *Engine) Request(pid *PID, msg any, timeout time.Duration) *Response {
	resp := NewResponse(e, timeout)
	e.Registry.Add(resp)
	e.SendWithSender(pid, msg, resp.PID())
	return resp
}

// SendWithSender will send the given message to the given PID with the
// given sender. Receivers can check the sender by calling Context.Sender().
func (e *Engine) SendWithSender(pid *PID, msg any, sender *PID) {
	e.send(pid, msg, sender)
}

// Send sends the given message to the given PID. If the message cannot be
// delivered due to the fact that the given process is not registered
// the message will be sent to the DeadLetter process instead.
func (e *Engine) Send(pid *PID, msg any) {
	e.send(pid, msg, nil)
}

// BroadcastEvent will broadcast the given message over the event stream, notifying all
// actors that are subscribed.
func (e *Engine) BroadcastEvent(msg any) {
	if e.eventStream != nil {
		e.send(e.eventStream, msg, nil)
	}
}

func (e *Engine) send(pid *PID, msg any, sender *PID) {
	if pid == nil {
		return
	}
	if e.isLocalMessage(pid) {
		e.SendLocal(pid, msg, sender)
		return
	}
	if e.remote == nil {
		e.BroadcastEvent(EngineRemoteMissingEvent{Target: pid, Sender: sender, Message: msg})
		return
	}
	e.remote.Send(pid, msg, sender)
}

// SenderAtInterval is a struct that can be used to send a message at an interval to a given PID.
// If you need to have an actor wake up periodically, you can use a SenderAtInterval.
// It is started by the "SendAtInterval" method and stopped by "Stop" method.
type SenderAtInterval struct {
	self     *PID
	target   *PID
	engine   *Engine
	message  any
	interval time.Duration
	cancelCh chan struct{}
}

func (r SenderAtInterval) Start() {
	ticker := time.NewTicker(r.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				r.engine.SendWithSender(r.target, r.message, r.self)
			case <-r.cancelCh:
				ticker.Stop()
				return
			}
		}
	}()
}

func (r SenderAtInterval) Stop() {
	close(r.cancelCh)
}

// SendAtInterval will send a message to a PID at a provided interval.
// It will return a "SenderAtInterval" struct that can stop the repeated sending of a message by calling "Stop".
func (e *Engine) SendAtInterval(pid *PID, msg any, interval time.Duration) SenderAtInterval {
	r := SenderAtInterval{
		self:     nil,
		target:   pid.CloneVT(),
		engine:   e,
		message:  msg,
		interval: interval,
		cancelCh: make(chan struct{}, 1),
	}
	r.Start()
	return r
}

// Stop will send a non-graceful "killProcess" message to the process that is associated with the given PID.
// The process will shut down immediately, once it has processed the "killProcess" message.
func (e *Engine) Stop(pid *PID) context.Context {
	return e.sendKillProcess(context.Background(), pid, false)
}

// Kill will send a graceful "killProcess" message to the process that is associated with the given PID.
// The process will shut down gracefully once it has processed all the messages in the inbox.
// A context is returned that can be used to block or wait until the process is stopped.
func (e *Engine) Kill(pid *PID) context.Context {
	return e.sendKillProcess(context.Background(), pid, true)
}

// KillWithCtx behaves the exact same way as "Kill", the only difference is that it accepts
// a context as the first argument. The context can be used for custom timeouts and manual
// cancellation.
func (e *Engine) KillWithCtx(ctx context.Context, pid *PID) context.Context {
	return e.sendKillProcess(ctx, pid, true)
}

func (e *Engine) sendKillProcess(ctx context.Context, pid *PID, graceful bool) context.Context {
	ctx, cancel := context.WithCancel(ctx)
	kill := killProcess{
		cancel:   cancel,
		graceful: graceful,
	}
	// if the process isn't found a "DeadLetterEvent" is broadcast.
	if e.Registry.Get(pid) == nil {
		e.BroadcastEvent(DeadLetterEvent{
			Target:  pid,
			Message: kill,
			Sender:  nil,
		})
		cancel()
		return ctx
	}
	e.SendLocal(pid, kill, nil)
	return ctx
}

// SendLocal will send the given message to the given PID. If the recipient is not in the
// registry, the message will be sent to the DeadLetter process instead. If there is no DeadLetter
// process registered, the function will panic.
func (e *Engine) SendLocal(pid *PID, msg any, sender *PID) {
	proc := e.Registry.Get(pid)
	if proc == nil {
		e.BroadcastEvent(DeadLetterEvent{
			Target:  pid,
			Message: msg,
			Sender:  nil,
		})
		return
	}
	proc.Send(pid, msg, sender)
}

// Subscribe will subscribe the given PID to the event stream.
func (e *Engine) Subscribe(pid *PID) {
	e.Send(e.eventStream, subscribeEvent{pid: pid})
}

// Unsubscribe will unsubscribe the given PID from the event stream.
func (e *Engine) Unsubscribe(pid *PID) {
	e.Send(e.eventStream, unsubscribeEvent{pid: pid})
}

func (e *Engine) isLocalMessage(pid *PID) bool {
	if pid == nil {
		return false
	}
	return e.address == pid.Address
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
