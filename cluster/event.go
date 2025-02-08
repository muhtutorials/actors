package cluster

import "github.com/muhtutorials/actors/actor"

// MemberJoinedEvent gets triggered each time a new member joins the cluster.
type MemberJoinedEvent struct {
	Member *Member
}

// MemberLeftEvent gets triggered each time a member leaves the cluster.
type MemberLeftEvent struct {
	Member *Member
}

// ActivationEvent gets triggered each time a new actor is activated somewhere on
// the cluster.
type ActivationEvent struct {
	PID *actor.PID
}

// DeactivationEvent gets triggered each time an actor gets deactivated somewhere on
// the cluster.
type DeactivationEvent struct {
	PID *actor.PID
}
