package main

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"github.com/muhtutorials/actors/actor"
	"github.com/muhtutorials/actors/examples/trade/actors/trade_engine"
	"github.com/muhtutorials/actors/examples/trade/types"
	"os"
	"time"
)

func main() {
	engine, err := actor.NewEngine(actor.NewEngineConfig())
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	tradeEnginePID := engine.Spawn(trade_engine.NewTradeEngine, "trade_engine")
	time.Sleep(time.Second)
	tradeOrder := types.TradeOrderRequest{
		TradeID:    GenerateID(),
		Token0:     "token0",
		Token1:     "token1",
		Chain:      "ETH",
		Wallet:     "wallet",
		PrivateKey: "pk",
		// empty ExpiresAt indicates no expiration
	}
	fmt.Println("sending trade order")
	engine.Send(tradeEnginePID, tradeOrder)
	time.Sleep(time.Second)
	resp := engine.Request(
		tradeEnginePID,
		types.TradeInfoRequest{TradeID: tradeOrder.TradeID},
		time.Second*5,
	)
	result, err := resp.Result()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	tradeInfo, ok := result.(types.TradeInfoResponse)
	if !ok {
		fmt.Println("trade info response error")
		os.Exit(1)
	}
	fmt.Println("trade info:", tradeInfo.Price)
	time.Sleep(time.Second)
	fmt.Println("canceling trade order")
	engine.Send(tradeEnginePID, types.CancelOrderRequest{TradeID: tradeOrder.TradeID})
	time.Sleep(time.Second)
	<-engine.Kill(tradeEnginePID).Done()
}

func GenerateID() string {
	id := make([]byte, 16)
	_, _ = rand.Read(id)
	return hex.EncodeToString(id)
}
