package shared

import (
	"fmt"
	"github.com/muhtutorials/actors/actor"
	"github.com/muhtutorials/actors/remote"
)

type Player struct{}

func NewPlayer() actor.Receiver {
	return &Player{}
}

func (p *Player) Receive(ctx *actor.Context) {
	switch msg := ctx.Message().(type) {
	case actor.Started:
	case *remote.TestMessage:
		fmt.Println("test message:", string(msg.Data))
	}
}
