package main

import (
	"fmt"
	"github.com/muhtutorials/actors/actor"
	"github.com/muhtutorials/actors/examples/remote/types"
	"github.com/muhtutorials/actors/remote"
	"time"
)

func main() {
	rem := remote.New("127.0.0.1:4000", remote.NewConfig())
	engine, err := actor.NewEngine(actor.NewEngineConfig().WithRemote(rem))
	if err != nil {
		panic(err)
	}
	serverPID := actor.NewPID("127.0.0.1:3000", "server/primary")
	for {
		engine.Send(serverPID, &types.Message{Data: "hey"})
		fmt.Println("sent message to:", serverPID)
		time.Sleep(time.Second)
	}
}
