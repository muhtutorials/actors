package main

import (
	"fmt"
	"github.com/muhtutorials/actors/actor"
	"time"
)

var restarts = 0

type message struct {
	data string
}

type barReceiver struct {
	data string
}

func newBarReceiver(data string) actor.Producer {
	return func() actor.Receiver {
		return &barReceiver{
			data: data,
		}
	}
}

func (r *barReceiver) Receive(ctx *actor.Context) {
	switch msg := ctx.Message().(type) {
	case actor.Started:
		time.Sleep(time.Second)
		if restarts < 2 {
			restarts++
			panic("need to restart")
		}
		fmt.Println("bar recovered and started with initial state:", r.data)
	case message:
		fmt.Println("bar received a message:", msg.data)
	case actor.Stopped:
		fmt.Println("bar stopped")
	}
}

type fooReceiver struct {
	barPID *actor.PID
}

func newFooReceiver() actor.Receiver {
	return &fooReceiver{}
}

func (r *fooReceiver) Receive(ctx *actor.Context) {
	switch msg := ctx.Message().(type) {
	case actor.Started:
		fmt.Println("foo started")
	case message:
		r.barPID = ctx.SpawnChild(newBarReceiver(msg.data), "bar", actor.WithID(msg.data))
		fmt.Println("received a message and starting bar:", r.barPID)
	case actor.Stopped:
		fmt.Println("foo stopped")
	}
}

func main() {
	engine, err := actor.NewEngine(actor.NewEngineConfig())
	if err != nil {
		panic(err)
	}
	pid := engine.Spawn(newFooReceiver, "foo")
	engine.Send(pid, message{data: ":-)"})
	time.Sleep(time.Second * 8)
	engine.Poison(pid)
	time.Sleep(time.Second)
}
