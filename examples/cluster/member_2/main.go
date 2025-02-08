package main

import (
	"fmt"
	"github.com/muhtutorials/actors/actor"
	"github.com/muhtutorials/actors/cluster"
	"github.com/muhtutorials/actors/examples/cluster/shared"
	"log"
	"reflect"
)

func main() {
	config := cluster.NewConfig().
		WithListenAddr("127.0.0.1:4000").
		WithID("B").
		WithRegion("us-west")
	clus, err := cluster.New(config)
	if err != nil {
		log.Fatal(err)
	}
	clus.RegisterKind(cluster.NewKindConfig(), "playerSession", shared.NewPlayer)
	eventPID := clus.Engine().SpawnFunc(func(ctx *actor.Context) {
		switch msg := ctx.Message().(type) {
		case cluster.ActivationEvent:
			fmt.Println("got activation event:", msg)
		default:
			fmt.Println("got:", reflect.TypeOf(msg))
		}
	}, "event")
	clus.Engine().Subscribe(eventPID)
	clus.Start()
	select {}
}
