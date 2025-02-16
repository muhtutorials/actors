package main

import (
	"fmt"
	"github.com/muhtutorials/actors/actor"
	"time"
)

type event struct {
	msg string
}

func main() {
	engine, _ := actor.NewEngine(actor.NewEngineConfig())
	actorA := engine.SpawnFunc(func(ctx *actor.Context) {
		switch msg := ctx.Message().(type) {
		case actor.Started:
			fmt.Println("actor A started")
		case actor.ActorStartedEvent:
			fmt.Println("another actor started:", msg.PID)
		case event:
			fmt.Println("actor A event:", msg)
		}
	}, "actor_a")
	engine.Subscribe(actorA)
	opt := func(opts *actor.Opts) {
		opts.ID = "patrick_star"
	}
	actorB := engine.SpawnFunc(func(ctx *actor.Context) {
		switch msg := ctx.Message().(type) {
		case actor.Started:
			fmt.Println("actor B started")
			engine.Subscribe(ctx.PID())
		case event:
			fmt.Println("actor B event:", msg)
		}
	}, "actor_b", opt)
	defer func() {
		engine.Unsubscribe(actorA)
		engine.Unsubscribe(actorB)
	}()
	time.Sleep(time.Millisecond)
	engine.BroadcastEvent(event{msg: "hey!"})
	time.Sleep(time.Millisecond)
}
