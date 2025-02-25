package trade_executor

import (
	"github.com/muhtutorials/actors/actor"
	"github.com/muhtutorials/actors/examples/trade/types"
	"log/slog"
	"time"
)

type TradeExecutorOpts struct {
	PriceWatcherPID *actor.PID
	TradeID         string
	Ticker          string
	Token0          string
	Token1          string
	Chain           string
	Wallet          string
	PrivateKey      string
	ExpiresAt       time.Time
}

type TradeExecutor struct {
	PID             *actor.PID
	Engine          *actor.Engine
	priceWatcherPID *actor.PID
	tradeID         string
	ticker          string
	token0          string
	token1          string
	chain           string
	wallet          string
	privateKey      string
	expiresAt       time.Time
	status          string
	price           float64
}

func NewTradeExecutor(opts *TradeExecutorOpts) actor.Producer {
	return func() actor.Receiver {
		return &TradeExecutor{
			priceWatcherPID: opts.PriceWatcherPID,
			tradeID:         opts.TradeID,
			ticker:          opts.Ticker,
			token0:          opts.Token0,
			token1:          opts.Token1,
			chain:           opts.Chain,
			wallet:          opts.Wallet,
			privateKey:      opts.PrivateKey,
			expiresAt:       opts.ExpiresAt,
			status:          "active",
		}
	}
}

func (t *TradeExecutor) Receive(ctx *actor.Context) {
	switch msg := ctx.Message().(type) {
	case actor.Started:
		slog.Info("trade executor started", "tradeID", t.tradeID, "wallet", t.wallet)
		t.PID = ctx.PID()
		t.Engine = ctx.Engine()
		t.Engine.Send(t.priceWatcherPID, types.Subscribe{PID: t.PID})
	case types.PriceResponse:
		t.handlePriceResponse(msg)
	case types.TradeInfoRequest:
		t.handleTradeInfoRequest(ctx)
	case types.CancelOrderRequest:
		slog.Info("TradeExecutor.CancelOrderRequest", "tradeID", t.tradeID, "wallet", t.wallet)
		t.status = "canceled"
		t.Kill()
	case actor.Stopped:
		slog.Info("trade executor stopped", "tradeID", t.tradeID, "wallet", t.wallet)
	}
}

func (t *TradeExecutor) handlePriceResponse(p types.PriceResponse) {
	if !t.expiresAt.IsZero() && time.Now().After(t.expiresAt) {
		slog.Info("trade expired", "tradeID", t.tradeID, "wallet", t.wallet)
		t.Kill()
		return
	}
	t.price = p.Price
	// do something with the price
	slog.Info("TradeExecutorActor.PriceResponse", "ticker", p.Ticker, "price", p.Price)
}

func (t *TradeExecutor) handleTradeInfoRequest(ctx *actor.Context) {
	slog.Info("TradeExecutor.TradeInfoRequest", "tradeID", t.tradeID, "wallet", t.wallet)
	ctx.Respond(types.TradeInfoResponse{
		Price: t.price,
	})
}

func (t *TradeExecutor) Kill() {
	if t.PID == nil {
		slog.Error("TradeExecutor.PID is <nil>")
		return
	}
	if t.Engine == nil {
		slog.Error("TradeExecutor.Engine is <nil>")
		return
	}
	t.Engine.Send(t.priceWatcherPID, types.Unsubscribe{PID: t.PID})
	t.Engine.Kill(t.PID)
}
