package main

import (
	"fmt"
	"github.com/muhtutorials/actors/actor"
	"time"
)

type Hooker interface {
	OnInit(ctx *actor.Context)
	OnStart(ctx *actor.Context)
	OnStop(ctx *actor.Context)
	Receive(ctx *actor.Context)
}

type foo struct{}

func newFoo() actor.Receiver {
	return &foo{}
}

func (f *foo) OnInit(_ *actor.Context) {
	fmt.Println("foo initialized")
}

func (f *foo) OnStart(_ *actor.Context) {
	fmt.Println("foo started")
}

func (f *foo) OnStop(_ *actor.Context) {
	fmt.Println("foo stopped")
}

func (f *foo) Receive(ctx *actor.Context) {
	switch msg := ctx.Message().(type) {
	case string:
		fmt.Println("foo received message:", msg)
	}
}

func WithHooks() func(actor.ReceiveFunc) actor.ReceiveFunc {
	return func(next actor.ReceiveFunc) actor.ReceiveFunc {
		return func(ctx *actor.Context) {
			switch ctx.Message().(type) {
			case actor.Initialized:
				ctx.Receiver().(Hooker).OnInit(ctx)
			case actor.Started:
				ctx.Receiver().(Hooker).OnStart(ctx)
			case actor.Stopped:
				ctx.Receiver().(Hooker).OnStop(ctx)
			}
			next(ctx)
		}
	}
}

func main() {
	engine, err := actor.NewEngine(actor.NewEngineConfig())
	if err != nil {
		panic(err)
	}
	pid := engine.Spawn(newFoo, "foo", actor.WithMiddleware(WithHooks()))
	engine.Send(pid, "hey!")
	time.Sleep(time.Second)
	<-engine.Kill(pid).Done()
}
