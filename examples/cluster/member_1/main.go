package main

import (
	"fmt"
	"github.com/muhtutorials/actors/actor"
	"github.com/muhtutorials/actors/cluster"
	"github.com/muhtutorials/actors/examples/cluster/shared"
	"github.com/muhtutorials/actors/remote"
	"log"
)

func main() {
	config := cluster.NewConfig().
		WithListenAddr("127.0.0.1:3000").
		WithID("A").
		WithRegion("eu-west")
	clus, err := cluster.New(config)
	if err != nil {
		log.Fatal(err)
	}
	clus.RegisterKind(cluster.NewKindConfig(), "player", shared.NewPlayer)
	eventPID := clus.Engine().SpawnFunc(func(ctx *actor.Context) {
		switch msg := ctx.Message().(type) {
		case cluster.ActivationEvent:
			fmt.Printf("Actor activated (PID=%s)\n", msg.PID)
		case cluster.MemberJoinedEvent:
			if msg.Member.ID == "B" {
				conf := cluster.NewActivationConfig().
					WithID("C").
					WithRegion("us-west")
				playerPID := clus.Activate(conf, "player")
				message := &remote.TestMessage{Data: []byte("hello from member 1")}
				ctx.Send(playerPID, message)
			}
		}
	}, "event")
	clus.Engine().Subscribe(eventPID)
	clus.Start()
	select {}
}
