package main

import (
	"github.com/muhtutorials/actors/actor"
	"github.com/muhtutorials/actors/examples/chat/types"
	"github.com/muhtutorials/actors/remote"
	"log/slog"
)

type clientMap map[string]*actor.PID

type userMap map[string]string

type server struct {
	clients clientMap // address: pid
	users   userMap   // address: username
	logger  *slog.Logger
}

func newServer() actor.Receiver {
	return &server{
		clients: make(clientMap),
		users:   make(userMap),
		logger:  slog.Default(),
	}
}

func (s *server) Receive(ctx *actor.Context) {
	switch msg := ctx.Message().(type) {
	case *types.Connect:
		addr := ctx.Sender().GetAddress()
		if _, ok := s.clients[addr]; ok {
			s.logger.Warn(
				"client already connected",
				"addr", ctx.Sender().GetAddress(),
				"id", ctx.Sender().GetID(),
			)
			return
		}
		if _, ok := s.users[addr]; ok {
			s.logger.Warn(
				"user already connected",
				"addr", ctx.Sender().GetAddress(),
				"id", ctx.Sender().GetID(),
			)
			return
		}
		s.clients[addr] = ctx.Sender()
		s.users[addr] = msg.Username
		slog.Info(
			"new client connected",
			"addr", ctx.Sender().GetAddress(),
			"id", ctx.Sender().GetID(),
			"username", msg.Username,
		)
	case *types.Disconnect:
		addr := ctx.Sender().GetAddress()
		_, ok := s.clients[addr]
		if !ok {
			s.logger.Warn(
				"unknown client disconnected",
				"addr", ctx.Sender().GetAddress(),
				"id", ctx.Sender().GetID(),
			)
			return
		}
		username, ok := s.users[addr]
		if !ok {
			s.logger.Warn(
				"unknown user disconnected",
				"addr", ctx.Sender().GetAddress(),
				"id", ctx.Sender().GetID(),
			)
			return
		}
		s.logger.Info("client disconnected", "username", username)
		delete(s.clients, addr)
		delete(s.users, username)
	case *types.Message:
		s.logger.Info("message received", "from", ctx.Sender(), "msg", msg.Msg)
		for _, pid := range s.clients {
			// Don't send message to the place where it came from.
			if !pid.Equals(ctx.Sender()) {
				s.logger.Info(
					"forwarding message",
					"addr", pid.Address,
					"id", pid.ID,
					"msg", ctx.Message(),
				)
				ctx.Forward(pid)
			}
		}
	}
}

func main() {
	rem := remote.New("127.0.0.1:3000", remote.NewConfig())
	engine, err := actor.NewEngine(actor.NewEngineConfig().WithRemote(rem))
	if err != nil {
		panic(err)
	}
	engine.Spawn(newServer, "server", actor.WithID("primary"))
	select {}
}
