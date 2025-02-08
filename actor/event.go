package actor

import (
	"log/slog"
	"time"
)

// EventLogger is an interface that events can implement. If they do,
// the event stream will log these events to slog.
type EventLogger interface {
	Log() (slog.Level, string, []any)
}

// ActorInitializedEvent is broadcast over the event stream before an actor
// receives and processes its "ActorStartedEvent".
type ActorInitializedEvent struct {
	PID       *PID
	Timestamp time.Time
}

func (e ActorInitializedEvent) Log() (slog.Level, string, []any) {
	return slog.LevelDebug, "Actor initialized", []any{"pid", e.PID}
}

// ActorStartedEvent is broadcast over the event stream each time
// a Receiver (actor) is spawned and activated. This means that at
// the point of receiving this event the Receiver (actor) is ready
// to process messages.
type ActorStartedEvent struct {
	PID       *PID
	Timestamp time.Time
}

func (e ActorStartedEvent) Log() (slog.Level, string, []any) {
	return slog.LevelDebug, "Actor started", []any{"pid", e.PID}
}

// ActorStoppedEvent is broadcast over the event stream each time
// a process is terminated.
type ActorStoppedEvent struct {
	PID       *PID
	Timestamp time.Time
}

func (e ActorStoppedEvent) Log() (slog.Level, string, []any) {
	return slog.LevelDebug, "Actor stopped", []any{"pid", e.PID}
}

// ActorRestartedEvent is broadcast when an actor crashes and gets restarted.
type ActorRestartedEvent struct {
	PID        *PID
	Timestamp  time.Time
	Stacktrace []byte
	Reason     any
	Restarts   int32
}

func (e ActorRestartedEvent) Log() (slog.Level, string, []any) {
	return slog.LevelError,
		"Actor crashed and restarted",
		[]any{"pid", e.PID.GetID(), "stack", string(e.Stacktrace), "reason", e.Reason, "restarts", e.Restarts}
}

// ActorMaxRestartsExceededEvent is broadcast if an actor crashes too many times.
type ActorMaxRestartsExceededEvent struct {
	PID       *PID
	Timestamp time.Time
}

func (e ActorMaxRestartsExceededEvent) Log() (slog.Level, string, []any) {
	return slog.LevelError, "Actor crashed too many times", []any{"pid", e.PID.GetID()}
}

// ActorDuplicateIDEvent is broadcast when trying to register the same name twice.
type ActorDuplicateIDEvent struct {
	PID *PID
}

func (e ActorDuplicateIDEvent) Log() (slog.Level, string, []any) {
	return slog.LevelError, "Actor name already exists", []any{"pid", e.PID.GetID()}
}

// EngineRemoteMissingEvent is broadcast when trying to send a message to a remote actor
// but the remote system is not available.
type EngineRemoteMissingEvent struct {
	Target  *PID
	Message any
	Sender  *PID
}

func (e EngineRemoteMissingEvent) Log() (slog.Level, string, []any) {
	return slog.LevelError, "Engine has no remote", []any{"sender", e.Target.GetID()}
}

// RemoteUnreachableEvent is broadcast when trying to send a message to
// a remote that is not reachable. The event will be broadcast after we
// retry to dial it "n" times.
type RemoteUnreachableEvent struct {
	// listen address of the remote we are trying to dial
	ListenAddr string
}

// DeadLetterEvent is broadcast when a message can't be delivered to its recipient.
type DeadLetterEvent struct {
	Target  *PID
	Message any
	Sender  *PID
}
