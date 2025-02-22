package main

import (
	"fmt"
	"github.com/muhtutorials/actors/actor"
	"sync"
	"time"
)

type message struct {
	data string
}

type foo struct {
	done func()
}

func newFoo(fn func()) actor.Producer {
	return func() actor.Receiver {
		return &foo{done: fn}
	}
}

func (f *foo) Receive(ctx *actor.Context) {
	switch msg := ctx.Message().(type) {
	case actor.Started:
		fmt.Println("foo started")
	case *message:
		if msg.data == "failed" {
			panic("failed processing this message")
		}
		fmt.Println("restarted and processed next message successfully:", msg.data)
		f.done()
	case actor.Stopped:
		fmt.Println("foo stopped")
	}
}

func main() {
	engine, err := actor.NewEngine(actor.NewEngineConfig())
	if err != nil {
		panic(err)
	}
	var wg sync.WaitGroup
	wg.Add(1)
	pid := engine.Spawn(newFoo(wg.Done), "foo")
	engine.Send(pid, &message{data: "failed"})
	time.Sleep(time.Millisecond)
	engine.Send(pid, &message{data: "hey"})
	wg.Wait()
}
