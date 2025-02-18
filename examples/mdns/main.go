package main

import (
	"fmt"
	"github.com/muhtutorials/actors/actor"
	"github.com/muhtutorials/actors/examples/mdns/chat"
	"github.com/muhtutorials/actors/examples/mdns/discovery"
	"github.com/muhtutorials/actors/remote"
	"log/slog"
	"os"
)

func main() {
	ip := "127.0.0.1"
	port := 3000
	slog.SetDefault(slog.New(slog.NewTextHandler(
		os.Stdout,
		&slog.HandlerOptions{Level: slog.LevelDebug},
	)))
	rem := remote.New(fmt.Sprintf("%s:%d", ip, port), remote.NewConfig())
	engine, err := actor.NewEngine(actor.NewEngineConfig().WithRemote(rem))
	if err != nil {
		panic(err)
	}
	engine.Spawn(chat.New(), "chat", actor.WithID("chat"))
	engine.Spawn(discovery.NewMDNSDiscovery(
		discovery.WithAnnounceAddr(ip, port),
	), "mdns")
	select {}
}
