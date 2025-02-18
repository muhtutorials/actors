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
	serviceName        = "_actor_"
	domain             = "local."
	memberPingInterval = time.Second * 2
)

// MemberAddr represents a reachable member in the cluster.
type MemberAddr struct {
	ListenAddr string
	ID         string
}

type ProviderConfig struct {
	bootstrapMembers []MemberAddr
}

func NewProviderConfig() ProviderConfig {
	return ProviderConfig{}
}

func (c ProviderConfig) WithBootstrapMembers(members ...MemberAddr) ProviderConfig {
	c.bootstrapMembers = append(c.bootstrapMembers, members...)
	return c
}

type Provider struct {
	config       ProviderConfig
	cluster      *Cluster
	pid          *actor.PID
	members      *MemberSet
	membersAlive *MemberSet
	memberPinger actor.Repeater
	eventPID     *actor.PID
	resolver     *zeroconf.Resolver
	announcer    *zeroconf.Server
	context      context.Context
	cancel       context.CancelFunc
}

func NewProvider(cfg ProviderConfig) Producer {
	return func(c *Cluster) actor.Producer {
		return func() actor.Receiver {
			return &Provider{
				config:       cfg,
				cluster:      c,
				members:      NewMemberSet(),
				membersAlive: NewMemberSet(),
			}
		}
	}
}

func (p *Provider) Receive(ctx *actor.Context) {
	switch msg := ctx.Message().(type) {
	case actor.Started:
		p.handleActorStarted(ctx)
	case actor.Stopped:
		p.handleActorStopped()
	case *Handshake:
		p.handleHandshake(ctx, msg)
	case *Members:
		p.addMembers(msg.Members...)
	case MemberLeft:
		p.handleMemberLeft(msg)
	case MemberPing:
		p.handleMemberPing(ctx)
	case *actor.Ping:
	case actor.Initialized:
	default:
		slog.Warn("received unhandled message", "msg", msg, "type", reflect.TypeOf(msg))
	}
}

func (p *Provider) handleActorStarted(ctx *actor.Context) {
	p.pid = ctx.PID()
	p.members.Add(p.cluster.Member())
	p.memberPinger = ctx.Repeat(ctx.PID(), MemberPing{}, memberPingInterval)
	p.context, p.cancel = context.WithCancel(context.Background())
	p.sendMembersToAgent()
	p.start(ctx)
}

func (p *Provider) handleActorStopped() {
	p.cluster.engine.Unsubscribe(p.eventPID)
	p.memberPinger.Stop()
	p.announcer.Shutdown()
	p.cancel()
}

func (p *Provider) handleHandshake(ctx *actor.Context, hs *Handshake) {
	p.addMembers(hs.Member)
	members := p.members.Slice()
	p.cluster.engine.Send(ctx.Sender(), &Members{Members: members})
}

func (p *Provider) handleMemberLeft(msg MemberLeft) {
	member := p.members.GetByAddress(msg.ListenAddr)
	p.removeMember(member)
}

func (p *Provider) handleMemberPing(ctx *actor.Context) {
	p.members.ForEach(func(member *Member) bool {
		if member.Address != p.cluster.agentPID.Address {
			ping := &actor.Ping{From: ctx.PID()}
			ctx.Send(memberToProvider(member), ping)
		}
		return true
	})
}

func (p *Provider) addMembers(members ...*Member) {
	for _, member := range members {
		if !p.members.Contains(member) {
			p.members.Add(member)
		}
		p.sendMembersToAgent()
	}
}

func (p *Provider) removeMember(member *Member) {
	if p.members.Contains(member) {
		p.members.Remove(member)
	}
	p.sendMembersToAgent()
}

// Send all the current members to the local cluster agent.
func (p *Provider) sendMembersToAgent() {
	members := &Members{Members: p.members.Slice()}
	p.cluster.engine.Send(p.cluster.PID(), members)
}

func (p *Provider) start(ctx *actor.Context) {
	p.eventPID = ctx.SpawnChildFunc(p.handleEvent, "event")
	p.cluster.engine.Subscribe(p.eventPID)
	// Send handshake to all bootstrap members if there are any.
	for _, member := range p.config.bootstrapMembers {
		memberPID := actor.NewPID(member.ListenAddr, "provider/"+member.ID)
		p.cluster.engine.SendWithSender(memberPID, &Handshake{
			Member: p.cluster.Member(),
		}, ctx.PID())
	}
	p.initAutoDiscovery()
	p.startAutoDiscovery()
}

func (p *Provider) handleEvent(ctx *actor.Context) {
	msg, ok := ctx.Message().(actor.RemoteUnreachableEvent)
	if ok {
		ctx.Send(p.pid, MemberLeft{ListenAddr: msg.ListenAddr})
	}
}

func (p *Provider) initAutoDiscovery() {
	resolver, err := zeroconf.NewResolver()
	if err != nil {
		log.Fatal(err)
	}
	p.resolver = resolver
	host, portStr, err := net.SplitHostPort(p.cluster.agentPID.Address)
	if err != nil {
		log.Fatal(err)
	}
	port, err := strconv.Atoi(portStr)
	if err != nil {
		log.Fatal(err)
	}
	server, err := zeroconf.RegisterProxy(
		p.cluster.ID(),
		serviceName,
		domain,
		port,
		fmt.Sprintf("member_%s", p.cluster.ID()),
		[]string{host},
		[]string{"txtv=0", "lo=1", "la=2"},
		nil,
	)
	if err != nil {
		log.Fatal(err)
	}
	p.announcer = server
}

func (p *Provider) startAutoDiscovery() {
	entries := make(chan *zeroconf.ServiceEntry)
	go p.handleServiceEntries(entries)
	if err := p.resolver.Browse(p.context, serviceName, domain, entries); err != nil {
		slog.Error("[CLUSTER] discovery failed", "err", err)
		panic(err)
	}
}

func (p *Provider) handleServiceEntries(results <-chan *zeroconf.ServiceEntry) {
	for entry := range results {
		if entry.Instance != p.cluster.ID() {
			addr := fmt.Sprintf("%s:%d", entry.AddrIPv4[0], entry.Port)
			hs := &Handshake{Member: p.cluster.Member()}
			// create a reachable PID for this member
			memberPID := actor.NewPID(addr, "provider/"+entry.Instance)
			self := actor.NewPID(p.cluster.agentPID.Address, "provider/"+p.cluster.ID())
			p.cluster.engine.SendWithSender(memberPID, hs, self)
		}
	}
	slog.Debug("[CLUSTER] stopping discovery", "id", p.cluster.ID())
}
