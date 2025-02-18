package discovery

import (
	"context"
	"fmt"
	"github.com/grandcat/zeroconf"
	"github.com/muhtutorials/actors/actor"
	"log/slog"
	"strings"
)

type mdns struct {
	id        string
	announcer *announcer
	resolver  *zeroconf.Resolver
	engine    *actor.Engine
	ctx       context.Context
	cancel    context.CancelFunc
}

func NewMDNSDiscovery(opts ...OptFunc) actor.Producer {
	options := applyOpts(opts...)
	ancr := newAnnouncer(options)
	ctx, cancel := context.WithCancel(context.Background())
	return func() actor.Receiver {
		return &mdns{
			id:        options.id,
			announcer: ancr,
			ctx:       ctx,
			cancel:    cancel,
		}
	}
}

func (m *mdns) Receive(ctx *actor.Context) {
	switch msg := ctx.Message().(type) {
	case actor.Initialized:
		m.engine = ctx.Engine()
		slog.Info("[DISCOVERY] initializing discovery")
		m.createResolver()
	case actor.Started:
		slog.Info("[DISCOVERY] starting discovery")
		go m.startDiscovery(m.ctx)
		m.announcer.start()
	case actor.Stopped:
		slog.Info("[DISCOVERY] stopping discovery")
		m.shutdown()
		_ = msg
	}
}

func (m *mdns) shutdown() {
	if m.announcer != nil {
		m.announcer.shutdown()
	}
	m.cancel()
}

func (m *mdns) createResolver() {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		panic(err)
	}
	m.resolver = resolver
}

// Starts multicast DNS discovery process.
// Searches matching entries with "serviceName" and "domain".
func (m *mdns) startDiscovery(ctx context.Context) {
	entries := make(chan *zeroconf.ServiceEntry)
	go func(results <-chan *zeroconf.ServiceEntry) {
		for entry := range results {
			m.sendDiscoveryEvent(entry)
		}
	}(entries)
	if err := m.resolver.Browse(ctx, serviceName, domain, entries); err != nil {
		slog.Info("[DISCOVERY] discovery failed")
		panic(err)
	}
	<-ctx.Done()
}

// Sends discovered peer as "DiscoveryEvent" to event stream.
func (m *mdns) sendDiscoveryEvent(entry *zeroconf.ServiceEntry) {
	// exclude self
	//if entry.Instance == m.id {
	//	return
	//}
	event := &DiscoveryEvent{ID: entry.Instance}
	for _, addr := range entry.AddrIPv4 {
		event.Addrs = append(
			event.Addrs,
			fmt.Sprintf("%s:%d", addr.String(), entry.Port),
		)
	}
	slog.Info(
		"[DISCOVERY] remote discovered",
		"ID", entry.Instance,
		"addrs", strings.Join(event.Addrs, ","),
	)
	if m.engine == nil {
		slog.Info("[DISCOVERY] engine not found")
		return
	}
	m.engine.BroadcastEvent(event)
}
