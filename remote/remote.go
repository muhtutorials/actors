package remote

import (
	"context"
	"crypto/tls"
	"fmt"
	"github.com/muhtutorials/actors/actor"
	"log/slog"
	"net"
	"storj.io/drpc/drpcmux"
	"storj.io/drpc/drpcserver"
	"sync"
	"sync/atomic"
)

const (
	initialized uint32 = iota
	running
	stopped
)

// Config holds remote's configuration.
type Config struct {
	TLSConfig *tls.Config
}

// NewConfig returns a new default remote's configuration.
func NewConfig() Config {
	return Config{}
}

// WithTLS sets the TLS config of the remote which will set
// the transport of the Remote to TLS.
func (cfg Config) WithTLS(c *tls.Config) Config {
	cfg.TLSConfig = c
	return cfg
}

type Remote struct {
	addr            string
	config          Config
	engine          *actor.Engine
	streamRouterPID *actor.PID
	state           atomic.Uint32
	// Stop closes this channel to signal the remote to stop listening.
	stopCh chan struct{}
	stopWG *sync.WaitGroup
}

func New(addr string, cfg Config) *Remote {
	r := &Remote{
		addr:   addr,
		config: cfg,
	}
	r.state.Store(initialized)
	return r
}

func (r *Remote) Start(e *actor.Engine) error {
	if r.state.Load() != initialized {
		return fmt.Errorf("remote already started")
	}
	r.state.Store(running)
	r.engine = e
	var (
		listener net.Listener
		err      error
	)
	if r.config.TLSConfig == nil {
		listener, err = net.Listen("tcp", r.addr)
	} else {
		slog.Debug("remote using TLS for listening")
		listener, err = tls.Listen("tcp", r.addr, r.config.TLSConfig)
	}
	if err != nil {
		return fmt.Errorf("remote failed to listen: %w", err)
	}
	fmt.Println("remote listening at:", r.addr)
	mux := drpcmux.New()
	if err = DRPCRegisterRemote(mux, newStreamReader(r)); err != nil {
		return fmt.Errorf("failed to register remote: %w", err)
	}
	server := drpcserver.New(mux)
	r.streamRouterPID = r.engine.Spawn(
		newStreamRouter(r.engine, r.config.TLSConfig),
		"router",
		actor.WithInboxSize(1024*1024),
	)
	slog.Debug("server started", "listenAddr", r.addr)
	r.stopCh = make(chan struct{})
	r.stopWG = &sync.WaitGroup{}
	r.stopWG.Add(1)
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		defer r.stopWG.Done()
		if err = server.Serve(ctx, listener); err != nil {
			slog.Error("DRPC server", "err", err)
		} else {
			slog.Debug("DRPC server stopped")
		}
	}()
	// wait for stopCh to be closed
	go func() {
		<-r.stopCh
		cancel()
	}()
	return nil
}

// Stop will stop the remote from listening.
func (r *Remote) Stop() *sync.WaitGroup {
	if r.state.Load() != running {
		slog.Warn("remote already stopped but stop has been called", "state", r.state.Load())
		return &sync.WaitGroup{} // return empty wait group so the caller can still wait without panicking
	}
	r.state.Store(stopped)
	r.stopCh <- struct{}{}
	return r.stopWG
}

// Send sends a message to the process with the PID over the network.
// Optionally, a sender PID can be provided to inform the receiving process who sent the
// message.
// Sending will work even if the remote is stopped. Receiving, however, will not work.
func (r *Remote) Send(pid *actor.PID, msg any, sender *actor.PID) {
	r.engine.Send(r.streamRouterPID, &streamMessage{
		target:  pid,
		message: msg,
		sender:  sender,
	})
}

// Address returns the listen address of the remote.
func (r *Remote) Address() string {
	return r.addr
}
