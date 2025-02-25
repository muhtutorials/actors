package types

import (
	"github.com/muhtutorials/actors/actor"
	"time"
)

type CancelOrderRequest struct {
	TradeID string
}

type TradeInfoRequest struct {
	TradeID string
}

type TradeInfoResponse struct {
	Price float64
}

type TradeOrderRequest struct {
	TradeID    string
	Token0     string
	Token1     string
	Chain      string
	Wallet     string
	PrivateKey string
	ExpiresAt  time.Time
}

// PriceRequest is used with "Repeat" to trigger price update
type PriceRequest struct{}

type PriceResponse struct {
	Ticker    string
	Price     float64
	UpdatedAt time.Time
}

type Subscribe struct {
	PID *actor.PID
}

type Unsubscribe struct {
	PID *actor.PID
}
