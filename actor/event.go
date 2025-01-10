package actor

import (
	"log/slog"
	"time"
)

// EventLogger is an interface that the various events can choose to implement.
// If they do, the event stream will log these events to slog.
type EventLogger interface {
	Log() (slog.Level, string, []any)
}

// EventActorStarted is broadcast over the event stream each time
// a Receiver (Actor) is spawned and activated. This means that at
// the point of receiving this event the Receiver (Actor) is ready
// to process messages.
type EventActorStarted struct {
	PID       *PID
	Timestamp time.Time
}

func (e EventActorStarted) Log() (slog.Level, string, []any) {
	// todo: should it be `e.PID` or `e.PID.GetID()`?
	return slog.LevelDebug, "Actor started.", []any{"pid", e.PID}
}

// EventActorInitialized is broadcast over the event stream before an actor
// received and processed its started event.
type EventActorInitialized struct {
	PID       *PID
	Timestamp time.Time
}

func (e EventActorInitialized) Log() (slog.Level, string, []any) {
	return slog.LevelDebug, "Actor initialized.", []any{"pid", e.PID}
}

// EventActorStopped is broadcast over the event stream each time
// a process is terminated.
type EventActorStopped struct {
	PID       *PID
	Timestamp time.Time
}

func (e EventActorStopped) Log() (slog.Level, string, []any) {
	return slog.LevelDebug, "Actor stopped.", []any{"pid", e.PID}
}

// EventActorRestarted is broadcast when an actor crashes and gets restarted.
type EventActorRestarted struct {
	PID        *PID
	Timestamp  time.Time
	Stacktrace []byte
	Reason     any
	Restarts   int32
}

func (e EventActorRestarted) Log() (slog.Level, string, []any) {
	return slog.LevelError, "Actor crashed and restarted.",
		[]any{"pid", e.PID.GetID(), "stack", string(e.Stacktrace),
			"reason", e.Reason, "restarts", e.Restarts}
}

// EventActorMaxRestartsExceeded is created if an actor crashes too many times.
type EventActorMaxRestartsExceeded struct {
	PID       *PID
	Timestamp time.Time
}

func (e EventActorMaxRestartsExceeded) Log() (slog.Level, string, []any) {
	return slog.LevelError, "Actor crashed too many times.", []any{"pid", e.PID.GetID()}
}

// EventActorDuplicateID is published if we try to register the same name twice.
type EventActorDuplicateID struct {
	PID *PID
}

func (e EventActorDuplicateID) Log() (slog.Level, string, []any) {
	return slog.LevelError, "Actor name already exists.", []any{"pid", e.PID.GetID()}
}

// EventEngineRemoteMissing is published if we try to send a message to a remote actor but the remote
// system is not available.
type EventEngineRemoteMissing struct {
	Target  *PID
	Message any
	Sender  *PID
}

func (e EventEngineRemoteMissing) Log() (slog.Level, string, []any) {
	return slog.LevelError, "Engine has no remote.", []any{"sender", e.Target.GetID()}
}

// EventRemoteUnreachable is published when trying to send a message to
// a remote that is not reachable. The event will be published after we
// retry to dial it "n" times.
type EventRemoteUnreachable struct {
	// listen address of the remote we are trying to dial
	ListenAddr string
}

// EventDeadLetter is delivered to the dead letter actor when
// a message can't be delivered to its recipient.
type EventDeadLetter struct {
	Target  *PID
	Message any
	Sender  *PID
}
