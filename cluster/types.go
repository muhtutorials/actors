package cluster

import "github.com/muhtutorials/actors/actor"

type (
	Activate struct {
		Config ActivationConfig
		Kind   string
	}

	Deactivate struct {
		PID *actor.PID
	}

	GetActivated struct {
		ID string
	}

	GetKinds struct{}

	GetMembers struct{}

	PingMembers struct{}

	MemberLeft struct {
		ListenAddr string
	}

	// MemberJoinedEvent gets triggered each time a new member joins the cluster.
	MemberJoinedEvent struct {
		Member *Member
	}

	// MemberLeftEvent gets triggered each time a member leaves the cluster.
	MemberLeftEvent struct {
		Member *Member
	}

	// ActivationEvent gets triggered each time a new actor is activated somewhere on
	// the cluster.
	ActivationEvent struct {
		PID *actor.PID
	}

	// DeactivationEvent gets triggered each time an actor gets deactivated somewhere on
	// the cluster.
	DeactivationEvent struct {
		PID *actor.PID
	}
)
