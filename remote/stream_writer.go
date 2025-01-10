package remote

import (
	"context"
	"crypto/tls"
	"errors"
	"github.com/muhtutorials/actors/actor"
	"io"
	"log/slog"
	"net"
	"storj.io/drpc/drpcconn"
	"sync"
	"time"
)

const (
	connIdleTimeout       = time.Minute * 10
	streamWriterBatchSize = 1024
)

type streamWriter struct {
	pid         *actor.PID
	routerPID   *actor.PID
	writeToAddr string
	conn        *drpcconn.Conn
	rawConn     net.Conn
	tlsConfig   *tls.Config
	stream      DRPCRemote_ReceiveStream
	serializer  Serializer
	engine      *actor.Engine
	inbox       actor.Inboxer
}

func newStreamWriter(routerPID *actor.PID, addr string, cfg *tls.Config, e *actor.Engine) actor.Processor {
	return &streamWriter{
		pid:         actor.NewPID(e.Address(), "stream/"+addr),
		routerPID:   routerPID,
		writeToAddr: addr,
		tlsConfig:   cfg,
		serializer:  ProtoSerializer{},
		engine:      e,
		inbox:       actor.NewInbox(streamWriterBatchSize),
	}
}

func (s *streamWriter) PID() *actor.PID {
	return s.pid
}

func (s *streamWriter) Send(_ *actor.PID, msg any, sender *actor.PID) {
	s.inbox.Send(actor.Envelope{
		Sender:  sender,
		Message: msg,
	})
}

func (s *streamWriter) Invoke(msgs []actor.Envelope) {
	var (
		typeLookup   = make(map[string]int32)
		typeNames    = make([]string, 0)
		targetLookup = make(map[uint64]int32)
		targets      = make([]*actor.PID, 0)
		senderLookup = make(map[uint64]int32)
		senders      = make([]*actor.PID, 0)
		messages     = make([]*Message, len(msgs))
	)
	for i := 0; i < len(msgs); i++ {
		var (
			typeID   int32
			targetID int32
			senderID int32
			stream   = msgs[i].Message.(*streamDeliver)
		)
		typeID, typeNames = LookUpTypeName(typeLookup, s.serializer.TypeName(stream.message), typeNames)
		targetID, targets = lookUpPIDs(targetLookup, stream.target, targets)
		senderID, senders = lookUpPIDs(senderLookup, stream.sender, senders)
		b, err := s.serializer.Serialize(stream.message)
		if err != nil {
			slog.Error("serialize", "err", err)
			continue
		}
		messages[i] = &Message{
			TypeNameIndex: typeID,
			TargetIndex:   targetID,
			SenderIndex:   senderID,
			Data:          b,
		}
		envelope := &Envelope{
			TypeNames: typeNames,
			Targets:   targets,
			Senders:   senders,
			Messages:  messages,
		}
		if err = s.stream.Send(envelope); err != nil {
			if errors.Is(err, io.EOF) {
				_ = s.conn.Close()
				return
			}
			slog.Error("stream writer failed to send message", "err", err)
		}
		// refresh connection deadline
		if err = s.rawConn.SetDeadline(time.Now().Add(connIdleTimeout)); err != nil {
			slog.Error("failed to set context deadline", "err", err)
		}
	}
}

func (s *streamWriter) Start() {
	s.inbox.Start(s)
	s.init()
}

func (s *streamWriter) ShutDown(wg *sync.WaitGroup) {
	event := actor.EventRemoteUnreachable{ListenAddr: s.writeToAddr}
	s.engine.Send(s.routerPID, event)
	s.engine.BroadcastEvent(event)
	if s.stream != nil {
		_ = s.stream.Close()
	}
	_ = s.inbox.Stop()
	s.engine.Registry.Remove(s.PID())
	if wg != nil {
		wg.Done()
	}
}

func (s *streamWriter) init() {
	var (
		rawConn    net.Conn
		err        error
		delay      = time.Millisecond * 500
		maxRetries = 3
	)
	for i := 0; i < maxRetries; i++ {
		// here we try to connect to the remote address
		if s.tlsConfig == nil {
			rawConn, err = net.Dial("tcp", s.writeToAddr)
			if err != nil {
				d := delay * time.Duration(i*2)
				slog.Error("net.Dial", "err", err, "remote", s.writeToAddr, "retry", i, "max", maxRetries, "delay", d)
				time.Sleep(d)
				continue
			}
		} else {
			slog.Debug("remote using TLS for writing")
			rawConn, err = tls.Dial("tcp", s.writeToAddr, s.tlsConfig)
			if err != nil {
				d := delay * time.Duration(i*2)
				slog.Error("tls.Dial", "err", err, "remote", s.writeToAddr, "retry", i, "max", maxRetries, "delay", d)
				time.Sleep(d)
				continue
			}
		}
		break
	}
	// We could not reach the remote after retrying "n" times. Hence, shut down the stream writer
	// and broadcast "EventRemoteUnreachable".
	if rawConn == nil {
		s.ShutDown(nil)
		return
	}
	s.rawConn = rawConn
	if err = rawConn.SetDeadline(time.Now().Add(connIdleTimeout)); err != nil {
		slog.Error("failed to set deadline on raw connection", "err", err)
		return
	}
	conn := drpcconn.New(rawConn)
	client := NewDRPCRemoteClient(conn)
	stream, err := client.Receive(context.Background())
	if err != nil {
		slog.Error("receive", "err", err, "remote", s.writeToAddr)
		s.ShutDown(nil)
		return
	}
	s.stream = stream
	s.conn = conn
	slog.Debug("connected", "remote", s.writeToAddr)
	go func() {
		<-s.conn.Closed()
		slog.Debug("lost connection", "remote", s.writeToAddr)
		s.ShutDown(nil)
	}()
}

func LookUpTypeName(m map[string]int32, name string, types []string) (int32, []string) {
	maximum := int32(len(m))
	id, ok := m[name]
	if !ok {
		m[name] = maximum
		id = maximum
		types = append(types, name)
	}
	return id, types
}

func lookUpPIDs(m map[uint64]int32, pid *actor.PID, pids []*actor.PID) (int32, []*actor.PID) {
	if pid == nil {
		return 0, pids
	}
	maximum := int32(len(m))
	key := pid.LookupKey()
	id, ok := m[key]
	if !ok {
		m[key] = maximum
		id = maximum
		pids = append(pids, pid)
	}
	return id, pids
}
