package main

import (
	"fmt"
	"github.com/muhtutorials/actors/actor"
	"log"
	"time"
)

type nameRequest struct{}

type nameResponse struct {
	name string
}

type nameResponder struct {
	name string
}

func newNameResponder() actor.Receiver {
	return &nameResponder{name: "Anon"}
}

func newCustomNameResponder(name string) actor.Producer {
	return func() actor.Receiver {
		return &nameResponder{name: name}
	}
}

func (r *nameResponder) Receive(ctx *actor.Context) {
	switch ctx.Message().(type) {
	case *nameRequest:
		ctx.Respond(&nameResponse{r.name})
	}
}

func main() {
	engine, err := actor.NewEngine(actor.NewEngineConfig())
	if err != nil {
		log.Fatal(err)
	}
	pid := engine.Spawn(newNameResponder, "responder")
	resp := engine.Request(pid, &nameRequest{}, time.Millisecond)
	result, err := resp.Result()
	if err != nil {
		log.Fatal(err)
	}
	if v, ok := result.(*nameResponse); ok {
		fmt.Println("received name:", v.name)
	}
	pid = engine.Spawn(newCustomNameResponder("Jack"), "custom_responder")
	resp = engine.Request(pid, &nameRequest{}, time.Millisecond)
	result, err = resp.Result()
	if err != nil {
		log.Fatal(err)
	}
	if v, ok := result.(*nameResponse); ok {
		fmt.Println("received name:", v.name)
	}
}
