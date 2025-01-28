package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/muhtutorials/actors/actor"
	"github.com/muhtutorials/actors/examples/chat/types"
	"github.com/muhtutorials/actors/remote"
	"log/slog"
	"math/rand"
	"os"
	"strconv"
)

type client struct {
	username  string
	serverPID *actor.PID
	logger    *slog.Logger
}

func newClient(username string, serverPID *actor.PID) actor.Producer {
	return func() actor.Receiver {
		return &client{
			username:  username,
			serverPID: serverPID,
			logger:    slog.Default(),
		}
	}
}

func (c *client) Receive(ctx *actor.Context) {
	switch msg := ctx.Message().(type) {
	case actor.Started:
		ctx.Send(c.serverPID, &types.Connect{
			Username: c.username,
		})
	case *actor.Stopped:
		c.logger.Info("client stopped")
	case *types.Message:
		fmt.Printf("%s: %s\n", msg.Username, msg.Msg)
	}
}

func main() {
	var (
		listenAddr = *flag.String("listen", "", "specify address to listen to, will pick a random port if not specified")
		connectTo  = *flag.String("connect", "127.0.0.1:3000", "the address of the server to connect to")
		username   = *flag.String("username", "", "your chat username")
	)
	flag.Parse()
	if listenAddr == "" {
		listenAddr = fmt.Sprintf("127.0.0.1:%d", rand.Int31n(50000)+10000)
	}
	if username == "" {
		username = "anon_" + strconv.Itoa(int(rand.Int63()))
	}
	rem := remote.New(listenAddr, remote.NewConfig())
	engine, err := actor.NewEngine(actor.NewEngineConfig().WithRemote(rem))
	if err != nil {
		slog.Error("failed to create engine", "err", err)
		os.Exit(1)
	}
	serverPID := actor.NewPID(connectTo, "server/primary")
	clientPID := engine.Spawn(newClient(username, serverPID), "client", actor.WithID(username))
	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Type 'quit' and press enter to exit.")
	for scanner.Scan() {
		msg := &types.Message{
			Username: username,
			Msg:      scanner.Text(),
		}
		// We use SendWithSender here so the server knows who
		// is sending the message.
		if msg.Msg == "quit" {
			break
		}
		engine.SendWithSender(serverPID, msg, clientPID)
	}
	if err = scanner.Err(); err != nil {
		slog.Error("failed to read message from stdin", "err", err)
	}
	// After breaking out of the loop on error let the server know
	// we need to disconnect.
	engine.SendWithSender(serverPID, &types.Disconnect{}, clientPID)
	engine.Poison(clientPID).Wait()
	slog.Info("client disconnected")
}
