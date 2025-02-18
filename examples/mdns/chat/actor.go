package chat

import (
	"github.com/muhtutorials/actors/actor"
	"github.com/muhtutorials/actors/examples/chat/types"
	"github.com/muhtutorials/actors/examples/mdns/discovery"
	"log/slog"
	"reflect"
)

type server struct {
	engine *actor.Engine
}

func New() actor.Producer {
	return func() actor.Receiver {
		return &server{}
	}
}

func (s *server) Receive(ctx *actor.Context) {
	switch msg := ctx.Message().(type) {
	case actor.Initialized:
		s.engine = ctx.Engine()
		ctx.Engine().Subscribe(ctx.PID())
	case actor.Started:
		_ = msg
	case actor.Stopped:
		ctx.Engine().Unsubscribe(ctx.PID())
	case *discovery.DiscoveryEvent:
		// a new remote actor has been discovered. We can now send messages to it.
		pid := actor.NewPID(msg.Addrs[0], "chat/chat")
		chatMessage := &types.Message{
			Msg: "Hey, there!",
		}
		slog.Info(
			"sending hello",
			"to", pid,
			"msg", chatMessage.Msg,
			"from", ctx.PID(),
		)
		s.engine.SendWithSender(pid, chatMessage, ctx.PID())
	case *types.Message:
		slog.Info(
			"received message",
			"from", ctx.Sender(),
			"msg", msg.Msg,
		)
	case actor.DeadLetterEvent:
		slog.Warn(
			"dead letter",
			"target", msg.Target,
			"msg", msg.Message,
			"sender", msg.Sender,
		)
	default:
		slog.Warn(
			"unknown message",
			"type", reflect.TypeOf(msg).String(),
			"msg", msg,
		)
	}
}
