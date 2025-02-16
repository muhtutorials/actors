package main

import (
	"fmt"
	"github.com/muhtutorials/actors/actor"
	"time"
)

type monitor struct {
	starts      int
	stops       int
	crashes     int
	deadLetters int
	message     string
}

func newMonitor() actor.Receiver {
	return &monitor{}
}

func (m *monitor) Receive(ctx *actor.Context) {
	switch ctx.Message().(type) {
	case actor.Initialized:
		ctx.Engine().Subscribe(ctx.PID())
	case actor.ActorStartedEvent:
		m.starts++
	case actor.ActorStoppedEvent:
		m.stops++
	case actor.ActorRestartedEvent:
		m.crashes++
	case actor.DeadLetterEvent:
		m.deadLetters++
	case query:
		ctx.Respond(query{
			starts:      m.starts,
			stops:       m.stops,
			crashes:     m.crashes,
			deadLetters: m.deadLetters,
		})
	}
}

type query struct {
	starts      int
	stops       int
	crashes     int
	deadLetters int
}

type message struct{}

type unstableActor struct {
	spawnTime time.Time
}

func newUnstableActor() actor.Receiver {
	return &unstableActor{}
}

func (a *unstableActor) Receive(ctx *actor.Context) {
	switch ctx.Message().(type) {
	case actor.Initialized:
		a.spawnTime = time.Now()
	case actor.Started:
		fmt.Println("actor started")
	case message:
		if time.Since(a.spawnTime) > time.Second {
			panic("aaahhh!!")
		}
	}
}

func main() {
	engine, _ := actor.NewEngine(actor.NewEngineConfig())
	mon := engine.Spawn(newMonitor, "monitor")
	ua := engine.Spawn(newUnstableActor, "unstable_actor", actor.WithMaxRestarts(10000))
	senderAtInterval := engine.SendAtInterval(ua, message{}, time.Millisecond*20)
	time.Sleep(time.Second * 5)
	senderAtInterval.Stop()
	<-engine.Kill(ua).Done()
	time.Sleep(time.Second * 1)
	resp, err := engine.Request(mon, query{}, time.Second).Result()
	if err != nil {
		fmt.Println("query:", err)
	}
	q := resp.(query)
	fmt.Printf("observed %d starts\n", q.starts)
	fmt.Printf("observed %d stops\n", q.stops)
	fmt.Printf("observed %d crashes\n", q.crashes)
	fmt.Printf("observed %d dead letters\n", q.deadLetters)
	<-engine.Kill(mon).Done()
	fmt.Println("done")
}
