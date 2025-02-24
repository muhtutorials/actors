package actor

import "github.com/zeebo/xxh3"

const pidSep = "/"

// NewPID returns a new Process ID given an address and an id.
func NewPID(addr, id string) *PID {
	return &PID{
		Address: addr,
		ID:      id,
	}
}

func (pid *PID) String() string {
	return pid.Address + pidSep + pid.ID
}

func (pid *PID) Equals(other *PID) bool {
	return pid.Address == other.Address && pid.ID == other.ID
}

func (pid *PID) Child(id string) *PID {
	childID := pid.ID + pidSep + id
	return NewPID(pid.Address, childID)
}

func (pid *PID) LookupKey() uint64 {
	key := []byte(pid.Address)
	// Converts string "pid.ID" to bytes and appends them to "key".
	key = append(key, pid.ID...)
	return xxh3.Hash(key)
}
