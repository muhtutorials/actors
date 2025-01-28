package remote

import (
	"crypto/tls"
	"github.com/muhtutorials/actors/actor"
	"log/slog"
)

type streamMessage struct {
	target  *actor.PID
	message any
	sender  *actor.PID
}

type streamRouter struct {
	engine    *actor.Engine
	tlsConfig *tls.Config
	pid       *actor.PID
	streams   map[string]*actor.PID
}

func newStreamRouter(e *actor.Engine, cfg *tls.Config) actor.Producer {
	return func() actor.Receiver {
		return &streamRouter{
			engine:    e,
			tlsConfig: cfg,
			streams:   make(map[string]*actor.PID),
		}
	}
}

func (s *streamRouter) Receive(ctx *actor.Context) {
	switch msg := ctx.Message().(type) {
	case actor.Started:
		s.pid = ctx.PID()
	case *streamMessage:
		s.handleStreamMessage(msg)
	case actor.EventRemoteUnreachable:
		s.handleTerminateStream(msg)
	}
}

func (s *streamRouter) handleStreamMessage(msg *streamMessage) {
	var (
		swpid *actor.PID // stream writer PID
		ok    bool
		addr  = msg.target.Address
	)
	swpid, ok = s.streams[addr]
	if !ok {
		swpid = s.engine.SpawnProcess(newStreamWriter(s.pid, addr, s.tlsConfig, s.engine))
		s.streams[addr] = swpid
	}
	s.engine.Send(swpid, msg)
}

func (s *streamRouter) handleTerminateStream(msg actor.EventRemoteUnreachable) {
	streamWriterPID := s.streams[msg.ListenAddr]
	delete(s.streams, msg.ListenAddr)
	slog.Debug("stream terminated",
		"remote", msg.ListenAddr,
		"pid", streamWriterPID,
	)
}
