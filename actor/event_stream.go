package actor

import (
	"context"
	"log/slog"
)

// eventSub is the message that will be sent to subscribe to the event stream.
type eventSub struct {
	pid *PID
}

// EventUnsub is the message that will be sent to unsubscribe from the event stream.
type eventUnsub struct {
	pid *PID
}

type eventStream struct {
	subs map[*PID]bool
}

func newEventStream() Producer {
	return func() Receiver {
		return &eventStream{
			subs: make(map[*PID]bool),
		}
	}
}

// Receive for the event stream. All system-wide events are sent here.
// Some events are specially handled, such as eventSub, eventUnsub for subscribing to events,
// DeadLetterSub, DeadLetterUnsub, for subscribing to EventDeadLetter.
func (e eventStream) Receive(ctx *Context) {
	switch message := ctx.Message().(type) {
	case eventSub:
		e.subs[message.pid] = true
	case eventUnsub:
		delete(e.subs, message.pid)
	default:
		// check if we should log the event, if so, log it with the relevant level, message and attributes
		logMsg, ok := ctx.Message().(EventLogger)
		if ok {
			level, msg, attrs := logMsg.Log()
			slog.Log(context.Background(), level, msg, attrs...)
		}
		for sub := range e.subs {
			ctx.Forward(sub)
		}
	}
}
