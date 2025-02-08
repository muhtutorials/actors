package actor

import (
	"sync"
)

const LocalLookupAddr = "local"

type Registry struct {
	lookup map[string]Processor
	engine *Engine
	mu     sync.RWMutex
}

func NewRegistry(e *Engine) *Registry {
	return &Registry{
		lookup: make(map[string]Processor, 1024),
		engine: e,
	}
}

// GetPID returns the process id associated for the given kind and its id.
// Returns nil if the process was not found.
func (r *Registry) GetPID(kind, id string) *PID {
	proc := r.GetByID(kind + pidSep + id)
	if proc != nil {
		return proc.PID()
	}
	return nil
}

// Get returns the processor for the given PID, if it exists.
// If it doesn't exist, nil is returned so the caller must check for that
// and direct the message to the dead letter processor instead.
func (r *Registry) Get(pid *PID) Processor {
	if pid == nil {
		return nil
	}
	r.mu.RLock()
	defer r.mu.RUnlock()
	if proc, ok := r.lookup[pid.ID]; ok {
		return proc
	}
	return nil
}

func (r *Registry) GetByID(id string) Processor {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.lookup[id]
}

func (r *Registry) Add(proc Processor) {
	r.mu.Lock()
	id := proc.PID().ID
	if _, ok := r.lookup[id]; ok {
		r.mu.Unlock()
		r.engine.BroadcastEvent(ActorDuplicateIDEvent{PID: proc.PID()})
		return
	}
	r.lookup[id] = proc
	r.mu.Unlock()
	proc.Start()
}

// Remove removes the given PID from the registry.
func (r *Registry) Remove(pid *PID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.lookup, pid.ID)
}
