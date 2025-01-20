package actor

import (
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

// Request sends the given message to the given PID as a "Request", returning
// a response that will resolve in the future. Calling Response.Result() will
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
		e.BroadcastEvent(EventEngineRemoteMissing{Target: pid, Sender: sender, Message: msg})
		return
	}
	e.remote.Send(pid, msg, sender)
}

// SendRepeater is a struct that can be used to send a repeating message to a given PID.
// If you need to have an actor wake up periodically, you can use a SendRepeater.
// It is started by the SendRepeat method and stopped by Stop method.
type SendRepeater struct {
	self     *PID
	target   *PID
	engine   *Engine
	message  any
	interval time.Duration
	cancelCh chan struct{}
}

func (sr SendRepeater) start() {
	ticker := time.NewTicker(sr.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				sr.engine.SendWithSender(sr.target, sr.message, sr.self)
			case <-sr.cancelCh:
				ticker.Stop()
				return
			}
		}
	}()
}

func (sr SendRepeater) Stop() {
	close(sr.cancelCh)
}

// SendRepeat will send the given message to the given PID each given interval.
// It will return a SendRepeater struct that can stop the repeating message by calling Stop().
func (e *Engine) SendRepeat(pid *PID, msg any, interval time.Duration) SendRepeater {
	sr := SendRepeater{
		self:     nil,
		target:   pid.CloneVT(),
		engine:   e,
		message:  msg,
		interval: interval,
		cancelCh: make(chan struct{}, 1),
	}
	sr.start()
	return sr
}

// Stop will send a non-graceful poisonPill message to the process that is associated with the given PID.
// The process will shut down immediately, once it has processed the poisonPill message.
func (e *Engine) Stop(pid *PID, wgs ...*sync.WaitGroup) *sync.WaitGroup {
	return e.sendPoisonPill(pid, false, wgs...)
}

// Poison will send a graceful poisonPill message to the process that is associated with the given PID.
// The process will shut down gracefully once it has processed all the messages in the inbox.
// If given a WaitGroup, it blocks till the process is completely shut down.
func (e *Engine) Poison(pid *PID, wg ...*sync.WaitGroup) *sync.WaitGroup {
	return e.sendPoisonPill(pid, true, wg...)
}

func (e *Engine) sendPoisonPill(pid *PID, graceful bool, wg ...*sync.WaitGroup) *sync.WaitGroup {
	waitGroup := new(sync.WaitGroup)
	if len(wg) > 0 {
		waitGroup = wg[0]
	}
	pill := poisonPill{
		wg:       waitGroup,
		graceful: graceful,
	}
	// if we didn't find a process, we will broadcast a EventDeadLetter.
	if e.Registry.Get(pid) == nil {
		e.BroadcastEvent(EventDeadLetter{
			Target:  pid,
			Message: pill,
			Sender:  nil,
		})
		return waitGroup
	}
	waitGroup.Add(1)
	e.SendLocal(pid, pill, nil)
	return waitGroup
}

// SendLocal will send the given message to the given PID. If the recipient is not in the
// registry, the message will be sent to the DeadLetter process instead. If there is no DeadLetter
// process registered, the function will panic.
func (e *Engine) SendLocal(pid *PID, msg any, sender *PID) {
	proc := e.Registry.Get(pid)
	if proc == nil {
		e.BroadcastEvent(EventDeadLetter{
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
	e.Send(e.eventStream, eventSub{pid: pid})
}

// Unsubscribe will unsubscribe the given PID from the event stream.
func (e *Engine) Unsubscribe(pid *PID) {
	e.Send(e.eventStream, eventUnsub{pid: pid})
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
