package actor

import (
	"context"
	"fmt"
	"github.com/muhtutorials/actors/safe_map"
	"math"
	"math/rand"
	"strconv"
	"time"
)

const LocalLookupAddr = "local"

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

// Engine represents the actor engine.
type Engine struct {
	Processes   *safe_map.SafeMap[string, Processor]
	address     string
	remote      Remoter
	eventStream *PID
}

// NewEngine returns a new actor Engine given an EngineConfig.
func NewEngine(cfg EngineConfig) (*Engine, error) {
	e := new(Engine)
	e.Processes = safe_map.New[string, Processor](1024)
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
	// check if we got an ID, generate one otherwise
	if opts.ID == "" {
		opts.ID = strconv.Itoa(rand.Intn(math.MaxInt))
	}
	return e.SpawnProcess(newProcess(e, opts))
}

// SpawnFunc spawns the given function as a stateless receiver/actor.
func (e *Engine) SpawnFunc(fn func(*Context), kind string, optFns ...OptFunc) *PID {
	return e.Spawn(newFuncReceiver(fn), kind, optFns...)
}

// SpawnProcess spawns the given Processor. This function is useful when working
// with custom created processes. Take a look at the streamWriter as an example.
func (e *Engine) SpawnProcess(proc Processor) *PID {
	pid := proc.PID()
	id := pid.ID
	if _, err := e.Processes.Get(id); err == nil {
		e.BroadcastEvent(ActorDuplicateIDEvent{PID: pid})
		return pid
	}
	e.Processes.Insert(id, proc)
	proc.Start()
	return pid
}

// Send sends the given message to the given PID. If the message cannot be
// delivered due to the fact that the given process is not registered
// the message will be sent to the DeadLetter process instead.
func (e *Engine) Send(pid *PID, msg any) {
	e.send(pid, msg, nil)
}

// SendWithSender will send the given message to the given PID with the
// given sender. Receivers can check the sender by calling Context.Sender method.
func (e *Engine) SendWithSender(pid *PID, msg any, sender *PID) {
	e.send(pid, msg, sender)
}

func (e *Engine) send(pid *PID, msg any, sender *PID) {
	if pid == nil {
		return
	}
	// check if it's a local message
	if e.address == pid.Address {
		e.SendLocally(pid, msg, sender)
		return
	}
	if e.remote == nil {
		e.BroadcastEvent(EngineRemoteMissingEvent{Target: pid, Sender: sender, Message: msg})
		return
	}
	e.remote.Send(pid, msg, sender)
}

// SendLocally will send the given message to the given PID. If the recipient is not in the
// "Processes", the message will be broadcast as a "DeadLetterEvent".
func (e *Engine) SendLocally(pid *PID, msg any, sender *PID) {
	proc, err := e.Processes.Get(pid.ID)
	if err != nil {
		e.BroadcastEvent(DeadLetterEvent{
			Target:  pid,
			Message: msg,
			Sender:  nil,
		})
		return
	}
	proc.Send(pid, msg, sender)
}

// BroadcastEvent will broadcast the given message over the event stream, notifying all
// actors that are subscribed.
func (e *Engine) BroadcastEvent(msg any) {
	if e.eventStream != nil {
		e.send(e.eventStream, msg, nil)
	}
}

// Subscribe will subscribe the given PID to the event stream.
func (e *Engine) Subscribe(pid *PID) {
	e.Send(e.eventStream, subscribeEvent{pid: pid})
}

// Unsubscribe will unsubscribe the given PID from the event stream.
func (e *Engine) Unsubscribe(pid *PID) {
	e.Send(e.eventStream, unsubscribeEvent{pid: pid})
}

// Request sends the message to the given PID as a "Request", returning
// a response that will resolve in the future. Calling "Response.Result" will
// block until the deadline is exceeded or the response is being resolved.
func (e *Engine) Request(pid *PID, msg any, timeout time.Duration) *Response {
	resp := NewResponse(e, timeout)
	e.Processes.Insert(resp.PID().ID, resp)
	e.SendWithSender(pid, msg, resp.PID())
	return resp
}

// Repeat will send a message to a PID at a provided interval.
// It will return a "Repeater" that can be used to stop it by calling Stop method.
func (e *Engine) Repeat(target *PID, msg any, interval time.Duration) Repeater {
	return e.repeat(target, msg, nil, interval)
}

// RepeatWithSender does the same thing as Repeat method only with sender specified.
func (e *Engine) RepeatWithSender(target *PID, msg any, sender *PID, interval time.Duration) Repeater {
	return e.repeat(target, msg, sender, interval)
}

func (e *Engine) repeat(pid *PID, msg any, sender *PID, interval time.Duration) Repeater {
	r := Repeater{
		target:   pid,
		message:  msg,
		sender:   sender,
		interval: interval,
		engine:   e,
		cancelCh: make(chan struct{}, 1),
	}
	r.Start()
	return r
}

// Repeater is used to send a message at an interval to a given PID.
type Repeater struct {
	target   *PID
	message  any
	sender   *PID
	interval time.Duration
	engine   *Engine
	cancelCh chan struct{}
}

func (r Repeater) Start() {
	ticker := time.NewTicker(r.interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				r.engine.SendWithSender(r.target, r.message, r.sender)
			case <-r.cancelCh:
				ticker.Stop()
				return
			}
		}
	}()
}

func (r Repeater) Stop() {
	close(r.cancelCh)
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
	// if the process isn't found, a "DeadLetterEvent" is broadcast
	if _, err := e.Processes.Get(pid.ID); err != nil {
		e.BroadcastEvent(DeadLetterEvent{
			Target:  pid,
			Message: kill,
			Sender:  nil,
		})
		cancel()
		return ctx
	}
	e.SendLocally(pid, kill, nil)
	return ctx
}

// Address returns the address of the actor engine. When there is
// no remote configured, the "local" address will be used, otherwise
// the listen address of the remote.
func (e *Engine) Address() string {
	return e.address
}
