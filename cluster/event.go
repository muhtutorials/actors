package cluster

import "github.com/muhtutorials/actors/actor"

// EventMemberJoin gets triggered each time a new member joins the cluster.
type EventMemberJoin struct {
	Member *Member
}

// EventMemberLeave gets triggered each time a member leaves the cluster.
type EventMemberLeave struct {
	Member *Member
}

// EventActivation gets triggered each time a new actor is activated somewhere on
// the cluster.
type EventActivation struct {
	PID *actor.PID
}

// EventDeactivation gets triggered each time an actor gets deactivated somewhere on
// the cluster.
type EventDeactivation struct {
	PID *actor.PID
}
