package actor

import (
	"context"
	"time"
)

const (
	defaultInboxSize    = 1024
	defaultMaxRestarts  = 3
	defaultRestartDelay = 500 * time.Millisecond
)

type ReceiveFunc = func(*Context)

type MiddlewareFunc = func(ReceiveFunc) ReceiveFunc

type Opts struct {
	Producer     Producer
	ID           string
	Kind         string
	InboxSize    int
	MaxRestarts  int32
	RestartDelay time.Duration
	Context      context.Context
	MiddleWare   []MiddlewareFunc
}

type OptFunc func(*Opts)

func DefaultOpts(p Producer) Opts {
	return Opts{
		Producer:     p,
		InboxSize:    defaultInboxSize,
		MaxRestarts:  defaultMaxRestarts,
		RestartDelay: defaultRestartDelay,
		Context:      context.Background(),
		MiddleWare:   []MiddlewareFunc{},
	}
}

func WithID(id string) OptFunc {
	return func(opts *Opts) {
		opts.ID = id
	}
}

func WithInboxSize(size int) OptFunc {
	return func(opts *Opts) {
		opts.InboxSize = size
	}
}

func WithMaxRestarts(n int) OptFunc {
	return func(opts *Opts) {
		opts.MaxRestarts = int32(n)
	}
}

func WithRestartDelay(d time.Duration) OptFunc {
	return func(opts *Opts) {
		opts.RestartDelay = d
	}
}

func WithContext(ctx context.Context) OptFunc {
	return func(opts *Opts) {
		opts.Context = ctx
	}
}

func WithMiddleware(mw ...MiddlewareFunc) OptFunc {
	return func(opts *Opts) {
		opts.MiddleWare = append(opts.MiddleWare, mw...)
	}
}
