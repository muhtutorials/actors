package actor

import (
	"context"
	"log/slog"
)

type eventStream struct {
	subs map[*PID]struct{}
}

func newEventStream() Producer {
	return func() Receiver {
		return &eventStream{
			subs: make(map[*PID]struct{}),
		}
	}
}

// Receive for the event stream. All system-wide events are sent here.
// Some events are specially handled, such as "subscribeEvent" and "unsubscribeEvent" for subscribing to events.
func (e eventStream) Receive(ctx *Context) {
	switch message := ctx.Message().(type) {
	case subscribeEvent:
		e.subs[message.pid] = struct{}{}
	case unsubscribeEvent:
		delete(e.subs, message.pid)
	default:
		// check if we should log the event, if so, log it with the relevant level, message and attributes
		eventLogger, ok := ctx.Message().(EventLogger)
		if ok {
			level, msg, attrs := eventLogger.Log()
			slog.Log(context.Background(), level, msg, attrs...)
		}
		for sub := range e.subs {
			ctx.Forward(sub)
		}
	}
}

// subscribeEvent is the message that will be sent to subscribe to the event stream.
type subscribeEvent struct {
	pid *PID
}

// unsubscribeEvent is the message that will be sent to unsubscribe from the event stream.
type unsubscribeEvent struct {
	pid *PID
}
