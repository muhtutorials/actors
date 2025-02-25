package price_watcher

import (
	"github.com/muhtutorials/actors/actor"
	"github.com/muhtutorials/actors/examples/trade/types"
	"log/slog"
	"math/rand"
	"time"
)

type PriceWatcherOpts struct {
	Ticker string
	Token0 string
	Token1 string
	Chain  string
}

type PriceWatcher struct {
	PID         *actor.PID
	Engine      *actor.Engine
	repeater    actor.Repeater
	ticker      string
	token0      string
	token1      string
	chain       string
	price       float64
	updatedAt   time.Time
	subscribers map[*actor.PID]struct{}
}

func NewPriceWatcher(opts PriceWatcherOpts) actor.Producer {
	return func() actor.Receiver {
		return &PriceWatcher{
			ticker:      opts.Ticker,
			token0:      opts.Token0,
			token1:      opts.Token1,
			chain:       opts.Chain,
			subscribers: make(map[*actor.PID]struct{}),
		}
	}
}

func (p *PriceWatcher) Receive(ctx *actor.Context) {
	switch msg := ctx.Message().(type) {
	case actor.Started:
		slog.Info("price watcher started", "ticker", p.ticker)
		p.PID = ctx.PID()
		p.Engine = ctx.Engine()
		// create a repeater to trigger price updates every 200ms
		p.repeater = p.Engine.Repeat(p.PID, types.PriceRequest{}, time.Millisecond*200)
	case types.PriceRequest:
		p.refresh()
	case types.Subscribe:
		slog.Info("price watcher subscribe", "ticker", p.ticker, "subscriber", msg.PID)
		p.subscribers[msg.PID] = struct{}{}
	case types.Unsubscribe:
		slog.Info("price watcher unsubscribe", "ticker", p.ticker, "subscriber", msg.PID)
		delete(p.subscribers, msg.PID)
	case actor.Stopped:
		slog.Info("price watcher stopped", "ticker", p.ticker)
	}
}

func (p *PriceWatcher) refresh() {
	if len(p.subscribers) == 0 {
		slog.Info("no subscribers, killing price watcher", "ticker", p.ticker)
		p.Kill()
		return
	}
	// price fluctuation is generated randomly for example
	p.price += float64(rand.Intn(10))
	p.updatedAt = time.Now()
	for pid := range p.subscribers {
		p.Engine.Send(pid, types.PriceResponse{
			Ticker:    p.ticker,
			Price:     p.price,
			UpdatedAt: p.updatedAt,
		})
	}
}

func (p *PriceWatcher) Kill() {
	p.repeater.Stop()
	if p.PID == nil {
		slog.Error("PriceWatcher.PID is <nil>", "ticker", p.ticker)
		return
	}
	if p.Engine == nil {
		slog.Error("PriceWatcher.Engine is <nil>", "ticker", p.ticker)
		return
	}
	p.Engine.Kill(p.PID)
}
