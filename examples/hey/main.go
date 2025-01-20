package main

import (
	"fmt"
	"github.com/muhtutorials/actors/actor"
	"log"
)

type terminator struct{}

func newTerminator() actor.Receiver {
	return &terminator{}
}

func (t *terminator) Receive(ctx *actor.Context) {
	switch msg := ctx.Message().(type) {
	case actor.Initialized:
		fmt.Println("terminator has initialized")
	case actor.Started:
		fmt.Println("terminator has started")
	case actor.Stopped:
		fmt.Println("terminator has stopped")
	case message:
		fmt.Println("message:", msg.data)
	}
}

type message struct {
	data string
}

func main() {
	engine, err := actor.NewEngine(actor.NewEngineConfig())
	if err != nil {
		log.Fatal(err)
	}
	pid := engine.Spawn(newTerminator, "cyborg")
	engine.SpawnFunc(func(c *actor.Context) {
		switch msg := c.Message().(type) {
		case actor.Started:
			fmt.Println("started")
			_ = msg
		}
	}, "foo")
	engine.Send(pid, message{data: "I'll be back!"})
	engine.Poison(pid).Wait()
}
