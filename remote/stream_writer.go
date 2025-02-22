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
	"time"
)

const (
	connIdleTimeout = time.Minute * 10
	batchSize       = 1024
)

type streamWriter struct {
	pid         *actor.PID
	routerPID   *actor.PID
	writeToAddr string
	rawConn     net.Conn
	conn        *drpcconn.Conn
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
		serializer:  ProtoSerde{},
		engine:      e,
		inbox:       actor.NewInbox(batchSize),
	}
}

func (s *streamWriter) Start() {
	s.inbox.Start(s)
	s.init()
}

func (s *streamWriter) ShutDown() {
	event := actor.RemoteUnreachableEvent{ListenAddr: s.writeToAddr}
	s.engine.Send(s.routerPID, event)
	s.engine.BroadcastEvent(event)
	if s.stream != nil {
		_ = s.stream.Close()
	}
	_ = s.inbox.Stop()
	_ = s.engine.Processes.Delete(s.PID().ID)
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
			typeNameIndex int32
			targetIndex   int32
			senderIndex   int32
			streamMsg     = msgs[i].Message.(*streamMessage)
		)
		typeNameIndex, typeNames = lookUpTypeName(typeLookup, s.serializer.TypeName(streamMsg.message), typeNames)
		targetIndex, targets = lookUpPID(targetLookup, streamMsg.target, targets)
		senderIndex, senders = lookUpPID(senderLookup, streamMsg.sender, senders)
		data, err := s.serializer.Serialize(streamMsg.message)
		if err != nil {
			slog.Error("serialize", "err", err)
			continue
		}
		messages[i] = &Message{
			TypeNameIndex: typeNameIndex,
			TargetIndex:   targetIndex,
			SenderIndex:   senderIndex,
			Data:          data,
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

func (s *streamWriter) PID() *actor.PID {
	return s.pid
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
		s.ShutDown()
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
		s.ShutDown()
		return
	}
	s.stream = stream
	s.conn = conn
	slog.Debug("connected", "remote", s.writeToAddr)
	go func() {
		<-s.conn.Closed()
		slog.Debug("lost connection", "remote", s.writeToAddr)
		s.ShutDown()
	}()
}

func lookUpTypeName(lookup map[string]int32, typeName string, types []string) (int32, []string) {
	newIndex := int32(len(lookup))
	index, ok := lookup[typeName]
	if !ok {
		lookup[typeName] = newIndex
		index = newIndex
		types = append(types, typeName)
	}
	return index, types
}

func lookUpPID(lookup map[uint64]int32, pid *actor.PID, pids []*actor.PID) (int32, []*actor.PID) {
	if pid == nil {
		return 0, pids
	}
	newIndex := int32(len(lookup))
	key := pid.LookupKey()
	index, ok := lookup[key]
	if !ok {
		lookup[key] = newIndex
		index = newIndex
		pids = append(pids, pid)
	}
	return index, pids
}
