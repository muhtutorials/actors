package main

import (
	"fmt"
	"github.com/muhtutorials/actors/actor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"time"
)

type PromMetrics struct {
	msgCounter prometheus.Counter
	msgLatency prometheus.Histogram
}

func NewPromMetrics(prefix string) *PromMetrics {
	msgCounter := promauto.NewCounter(prometheus.CounterOpts{
		Name: fmt.Sprintf("%s_actor_msg_counter", prefix),
		Help: "actor message counter",
	})
	msgLatency := promauto.NewHistogram(prometheus.HistogramOpts{
		Name: fmt.Sprintf("%s_actor_msg_latency", prefix),
		Help: "actor message latency",
		// Buckets is a predefined range of values used to categorize and count
		// observations in histogram metrics. Each bucket represents a cumulative
		// count of all observations less than or equal to its upper bound.
		// "0.1, 0.5, 1" are buckets into which latency of 0.1s, 0.5s and 1s will go.
		Buckets: []float64{0.1, 0.5, 1},
	})
	return &PromMetrics{
		msgCounter: msgCounter,
		msgLatency: msgLatency,
	}
}

func (p *PromMetrics) WithMetrics() func(actor.ReceiveFunc) actor.ReceiveFunc {
	return func(next actor.ReceiveFunc) actor.ReceiveFunc {
		return func(ctx *actor.Context) {
			start := time.Now()
			p.msgCounter.Inc()
			next(ctx)
			secs := time.Since(start).Seconds()
			p.msgLatency.Observe(secs)
		}
	}
}

type message struct {
	data string
}

type foo struct{}

func newFoo() actor.Receiver {
	return &foo{}
}

func (f *foo) Receive(ctx *actor.Context) {
	switch msg := ctx.Message().(type) {
	case actor.Started:
		fmt.Println("foo started")
	case message:
		fmt.Println("foo received message:", msg.data)
	}
}

type bar struct{}

func newBar() actor.Receiver {
	return &bar{}
}

func (b *bar) Receive(ctx *actor.Context) {
	switch msg := ctx.Message().(type) {
	case actor.Started:
		fmt.Println("bar started")
	case message:
		fmt.Println("bar received message:", msg.data)
	}
}

func main() {
	go func() {
		if err := http.ListenAndServe(":8000", promhttp.Handler()); err != nil {
			panic(err)
		}
	}()
	engine, err := actor.NewEngine(actor.NewEngineConfig())
	if err != nil {
		panic(err)
	}
	var (
		fooMetrics = NewPromMetrics("foo")
		barMetrics = NewPromMetrics("bar")
		fooPID     = engine.Spawn(newFoo, "foo", actor.WithMiddleware(fooMetrics.WithMetrics()))
		barPID     = engine.Spawn(newBar, "bar", actor.WithMiddleware(barMetrics.WithMetrics()))
	)
	for i := 0; i < 10; i++ {
		engine.Send(fooPID, message{data: "hey"})
		engine.Send(barPID, message{data: "hi"})
		time.Sleep(time.Second)
	}
	select {}
}
