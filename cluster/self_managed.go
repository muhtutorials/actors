package cluster

import (
	"context"
	"fmt"
	"github.com/grandcat/zeroconf"
	"github.com/muhtutorials/actors/actor"
	"log"
	"log/slog"
	"net"
	"reflect"
	"strconv"
	"time"
)

const (
	serviceName        = "_some_"
	domain             = "local."
	memberPingInterval = time.Second * 2
)

// MemberAddr represents a reachable node in the cluster.
type MemberAddr struct {
	ListenAddr string
	ID         string
}

type memberLeave struct {
	ListenAddr string
}

type memberPing struct{}

type SelfManagedConfig struct {
	bootstrapMembers []MemberAddr
}

func NewSelfManagedConfig() SelfManagedConfig {
	return SelfManagedConfig{}
}

func (c SelfManagedConfig) WithBootstrapMember(m MemberAddr) SelfManagedConfig {
	c.bootstrapMembers = append(c.bootstrapMembers, m)
	return c
}

type SelfManaged struct {
	config       SelfManagedConfig
	cluster      *Cluster
	pid          *actor.PID
	members      *MemberSet
	membersAlive *MemberSet
	memberPinger actor.SenderAtInterval
	eventSubPID  *actor.PID
	resolver     *zeroconf.Resolver
	announcer    *zeroconf.Server
	context      context.Context
	cancel       context.CancelFunc
}

func NewSelfManagedProvider(cfg SelfManagedConfig) Producer {
	return func(c *Cluster) actor.Producer {
		return func() actor.Receiver {
			return &SelfManaged{
				config:       cfg,
				cluster:      c,
				members:      NewMemberSet(),
				membersAlive: NewMemberSet(),
			}
		}
	}
}

func (s *SelfManaged) Receive(ctx *actor.Context) {
	switch msg := ctx.Message().(type) {
	case actor.Started:
		s.handleActorStarted(ctx)
	case actor.Stopped:
		s.handleActorStopped()
	case *Handshake:
		s.handleHandshake(ctx, msg)
	case *Members:
		s.addMembers(msg.Members...)
	case memberLeave:
		s.handleMemberLeave(msg)
	case memberPing:
		s.handleMemberPing(ctx)
	case actor.Ping:
	case actor.Initialized:
		_ = msg
	default:
		slog.Warn("received unhandled message", "msg", msg, "type", reflect.TypeOf(msg))
	}
}

func (s *SelfManaged) handleActorStarted(ctx *actor.Context) {
	s.pid = ctx.PID()
	s.members.Add(s.cluster.Member())
	s.memberPinger = ctx.SendAtInterval(ctx.PID(), memberPing{}, memberPingInterval)
	s.context, s.cancel = context.WithCancel(context.Background())
	s.sendMembersToAgent()
	s.start(ctx)
}

func (s *SelfManaged) handleActorStopped() {
	// todo: does order matter here?
	s.cluster.engine.Unsubscribe(s.eventSubPID)
	s.memberPinger.Stop()
	s.announcer.Shutdown()
	s.cancel()
}

func (s *SelfManaged) handleHandshake(ctx *actor.Context, h *Handshake) {
	s.addMembers(h.Member)
	members := s.members.Slice()
	s.cluster.engine.Send(ctx.Sender(), &Members{Members: members})
}

func (s *SelfManaged) handleMemberLeave(msg memberLeave) {
	member := s.members.GetByHost(msg.ListenAddr)
	s.removeMember(member)
}

func (s *SelfManaged) handleMemberPing(ctx *actor.Context) {
	s.members.ForEach(func(member *Member) bool {
		if member.Host != s.cluster.agentPID.Address {
			ping := &actor.Ping{From: ctx.PID()}
			ctx.Send(memberToProvider(member), ping)
		}
		return true
	})
}

func (s *SelfManaged) addMembers(members ...*Member) {
	for _, member := range members {
		if !s.members.Contains(member) {
			s.members.Add(member)
		}
		s.sendMembersToAgent()
	}
}

func (s *SelfManaged) removeMember(member *Member) {
	if s.members.Contains(member) {
		s.members.Remove(member)
	}
	s.sendMembersToAgent()
}

// Send all the current members to the local cluster agent.
func (s *SelfManaged) sendMembersToAgent() {
	members := &Members{Members: s.members.Slice()}
	s.cluster.engine.Send(s.cluster.PID(), members)
}

func (s *SelfManaged) start(ctx *actor.Context) {
	s.eventSubPID = ctx.SpawnChildFunc(s.handleEventStream, "event")
	s.cluster.engine.Subscribe(s.eventSubPID)
	// Send handshake to all bootstrap members if any.
	for _, member := range s.config.bootstrapMembers {
		memberPID := actor.NewPID(member.ListenAddr, "provider/"+member.ID)
		s.cluster.engine.SendWithSender(memberPID, &Handshake{
			Member: s.cluster.Member(),
		}, ctx.PID())
	}
	s.initAutoDiscovery()
	s.startAutoDiscovery()
}

func (s *SelfManaged) handleEventStream(ctx *actor.Context) {
	msg, ok := ctx.Message().(actor.EventRemoteUnreachable)
	if ok {
		ctx.Send(s.pid, memberLeave{ListenAddr: msg.ListenAddr})
	}
}

func (s *SelfManaged) initAutoDiscovery() {
	resolver, err := zeroconf.NewResolver()
	if err != nil {
		log.Fatal(err)
	}
	s.resolver = resolver
	host, portStr, err := net.SplitHostPort(s.cluster.agentPID.Address)
	if err != nil {
		log.Fatal(err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatal(err)
	}
	server, err := zeroconf.RegisterProxy(
		s.cluster.ID(),
		serviceName,
		domain,
		port,
		fmt.Sprintf("member_%s", s.cluster.ID()),
		[]string{host},
		[]string{"txtv=0", "lo=1", "la=2"},
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}
	s.announcer = server
}

func (s *SelfManaged) startAutoDiscovery() {
	entries := make(chan *zeroconf.ServiceEntry)
	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			if entry.Instance != s.cluster.ID() {
				host := fmt.Sprintf("%s:%d", entry.AddrIPv4[0], entry.Port)
				hs := &Handshake{Member: s.cluster.Member()}
				// create a reachable PID for this member.
				memberPID := actor.NewPID(host, "provider/"+entry.Instance)
				self := actor.NewPID(s.cluster.agentPID.Address, "provider/"+s.cluster.ID())
				s.cluster.engine.SendWithSender(memberPID, hs, self)
			}
		}
		slog.Debug("[CLUSTER] stopping discovery", "id", s.cluster.ID())
	}(entries)
	if err := s.resolver.Browse(s.context, serviceName, domain, entries); err != nil {
		slog.Error("[CLUSTER] discovery failed", "err", err)
		panic(err)
	}
}
