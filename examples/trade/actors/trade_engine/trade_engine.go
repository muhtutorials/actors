package trade_engine

import (
	"fmt"
	"github.com/muhtutorials/actors/actor"
	"github.com/muhtutorials/actors/examples/trade/actors/price_watcher"
	"github.com/muhtutorials/actors/examples/trade/actors/trade_executor"
	"github.com/muhtutorials/actors/examples/trade/types"
	"github.com/muhtutorials/actors/safe_map"
	"log/slog"
	"time"
)

type TradeEngine struct {
	executors     *safe_map.SafeMap[string, *actor.PID]
	priceWatchers *safe_map.SafeMap[string, *actor.PID]
}

func NewTradeEngine() actor.Receiver {
	return &TradeEngine{
		executors:     safe_map.New[string, *actor.PID](),
		priceWatchers: safe_map.New[string, *actor.PID](),
	}
}

func (t *TradeEngine) Receive(ctx *actor.Context) {
	switch msg := ctx.Message().(type) {
	case actor.Started:
		fmt.Println("trade engine started")
	case types.TradeOrderRequest:
		t.spawnExecutor(ctx, msg)
	case types.TradeInfoRequest:
		t.handleTradeInfoRequest(ctx, msg)
	case types.CancelOrderRequest:
		t.handleCancelOrder(ctx, msg)
	case actor.Stopped:
		fmt.Println("trade engine stopped")
	}
}

func (t *TradeEngine) spawnExecutor(ctx *actor.Context, msg types.TradeOrderRequest) {
	slog.Info("TradeEngine.TradeOrderRequest", "tradeID", msg.TradeID, "wallet", msg.Wallet)
	// make sure there is a price watcher for this token pair
	priceWatcherPID := t.spawnPriceWatcher(ctx, msg)
	opts := &trade_executor.TradeExecutorOpts{
		PriceWatcherPID: priceWatcherPID,
		TradeID:         msg.TradeID,
		Ticker:          toTicker(msg.Token0, msg.Token1, msg.Chain),
		Token0:          msg.Token0,
		Token1:          msg.Token1,
		Chain:           msg.Chain,
		Wallet:          msg.Wallet,
		PrivateKey:      msg.PrivateKey,
		ExpiresAt:       msg.ExpiresAt,
	}
	pid := ctx.SpawnChild(trade_executor.NewTradeExecutor(opts), msg.TradeID)
	t.executors.Insert(opts.TradeID, pid)
}

func (t *TradeEngine) spawnPriceWatcher(ctx *actor.Context, msg types.TradeOrderRequest) *actor.PID {
	ticker := toTicker(msg.Token0, msg.Token1, msg.Chain)
	// look for existing price watcher in trade engine
	pid, _ := t.priceWatchers.Get(ticker)
	if pid != nil {
		return pid
	}
	opts := price_watcher.PriceWatcherOpts{
		Ticker: ticker,
		Token0: msg.Token0,
		Token1: msg.Token1,
		Chain:  msg.Chain,
	}
	pid = ctx.SpawnChild(price_watcher.NewPriceWatcher(opts), ticker)
	t.priceWatchers.Insert(ticker, pid)
	return pid
}

func (t *TradeEngine) handleTradeInfoRequest(ctx *actor.Context, msg types.TradeInfoRequest) {
	slog.Info("TradeEngine.TradeInfoRequest", "id", msg.TradeID)
	// get the executor
	pid, _ := t.executors.Get(msg.TradeID)
	if pid == nil {
		slog.Error("failed to get trade info", "err", "TradeExecutor PID not found", "TradeID", msg.TradeID)
		return
	}
	resp := ctx.Request(pid, types.TradeInfoRequest{}, time.Second*5)
	result, err := resp.Result()
	if err != nil {
		slog.Error("failed to get trade info", "err", err)
		return
	}
	v, ok := result.(types.TradeInfoResponse)
	if !ok {
		slog.Error("failed to get trade info", "err", "unknown response type")
		return
	}
	ctx.Respond(v)
}

func (t *TradeEngine) handleCancelOrder(ctx *actor.Context, msg types.CancelOrderRequest) {
	slog.Info("TradeEngine.CancelOrderRequest", "id", msg.TradeID)
	// get the executor
	pid, _ := t.executors.Get(msg.TradeID)
	if pid == nil {
		slog.Error("failed to cancel order", "err", "TradeExecutor PID not found", "TradeID", msg.TradeID)
		return
	}
	ctx.Send(pid, types.CancelOrderRequest{})
}

func toTicker(token0, token1, chain string) string {
	return fmt.Sprintf("%s-%s-%s", token0, token1, chain)
}
