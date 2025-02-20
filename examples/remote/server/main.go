package main

import (
	"fmt"
	"github.com/muhtutorials/actors/actor"
	"github.com/muhtutorials/actors/examples/remote/types"
	"github.com/muhtutorials/actors/remote"
)

type server struct{}

func newServer() actor.Receiver {
	return &server{}
}

func (s *server) Receive(ctx *actor.Context) {
	switch msg := ctx.Message().(type) {
	case actor.Started:
		fmt.Println("server started")
	case *types.Message:
		fmt.Println("server got message:", msg.Data)
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
