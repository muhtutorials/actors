package main

import (
	"fmt"
	"github.com/muhtutorials/actors/actor"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type handler struct{}

func newHandler() actor.Receiver {
	return &handler{}
}

func (handler) Receive(ctx *actor.Context) {
	switch msg := ctx.Message().(type) {
	case []byte:
		fmt.Println("got message to handle:", string(msg))
	case actor.Stopped:
		for i := 0; i < 3; i++ {
			fmt.Println("\r handler stopping in", 3-i)
			time.Sleep(time.Second)
		}
		fmt.Println("handler stopped")
	}
}

type session struct {
	conn net.Conn
}

func newSession(conn net.Conn) actor.Producer {
	return func() actor.Receiver {
		return &session{conn: conn}
	}
}

func (s *session) Receive(ctx *actor.Context) {
	switch ctx.Message().(type) {
	case actor.Started:
		fmt.Println("new connection:", s.conn.RemoteAddr())
		go s.readLoop(ctx)
	case actor.Stopped:
		s.conn.Close()
	}
}

func (s *session) readLoop(ctx *actor.Context) {
	buf := make([]byte, 1024)
	for {
		n, err := s.conn.Read(buf)
		if err != nil {
			fmt.Println(err)
			break
		}
		msg := make([]byte, n)
		// Copy shared buffer to prevent race conditions.
		copy(msg, buf[:n])
		// Send to the handler to process to message.
		ctx.Send(ctx.Parent().Child("handler/default"), msg)
	}
	// Loop is done due to an error, or we need to close due to server shutdown.
	ctx.Send(ctx.Parent(), &connRemove{pid: ctx.PID()})
}

type connAdd struct {
	pid  *actor.PID
	conn net.Conn
}

type connRemove struct {
	pid *actor.PID
}

type server struct {
	addr     string
	listener net.Listener
	sessions map[*actor.PID]net.Conn
}

func newServer(addr string) actor.Producer {
	return func() actor.Receiver {
		return &server{
			addr:     addr,
			sessions: make(map[*actor.PID]net.Conn),
		}
	}
}

func (s *server) Receive(ctx *actor.Context) {
	switch msg := ctx.Message().(type) {
	case actor.Initialized:
		lis, err := net.Listen("tcp", s.addr)
		if err != nil {
			panic(err)
		}
		s.listener = lis
		// start the handler that will handle the incoming messages from clients/sessions.
		ctx.SpawnChild(newHandler, "handler", actor.WithID("default"))
	case actor.Started:
		fmt.Println("server started")
		go s.acceptLoop(ctx)
	case actor.Stopped:
		// On stop all the child sessions will automatically get the stop
		// message and close all their underlying connections.
	case *connAdd:
		fmt.Printf("added new connection: addr %s, pid %v\n", msg.conn.RemoteAddr(), msg.pid)
		s.sessions[msg.pid] = msg.conn
	case *connRemove:
		fmt.Printf("removed connection: pid %v", msg.pid)
		delete(s.sessions, msg.pid)
	}
}

func (s *server) acceptLoop(ctx *actor.Context) {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			fmt.Println(err)
			break
		}
		pid := ctx.SpawnChild(
			newSession(conn),
			"session",
			actor.WithID(conn.RemoteAddr().String()),
		)
		ctx.Send(ctx.PID(), &connAdd{
			pid:  pid,
			conn: conn,
		})
	}
}

func main() {
	engine, err := actor.NewEngine(actor.NewEngineConfig())
	if err != nil {
		panic(err)
	}
	serverPID := engine.Spawn(newServer(":3000"), "server")
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	<-engine.Kill(serverPID).Done()
}
