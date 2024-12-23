package cluster

import (
	"fmt"
	"math"
	"math/rand"
)

type ActivationConfig struct {
	id           string
	region       string
	selectMember SelectMemberFunc
}

// NewActivationConfig returns a new default config.
func NewActivationConfig() ActivationConfig {
	return ActivationConfig{
		id:           fmt.Sprintf("%d", rand.Intn(math.MaxInt)),
		region:       "default",
		selectMember: SelectRandomMember,
	}
}

// WithID sets the id of the actor that will be activated on the cluster.
//
// Defaults to a random identifier.
func (cfg ActivationConfig) WithID(id string) ActivationConfig {
	cfg.id = id
	return cfg
}

// WithRegion sets the region where this actor should be spawned.
//
// Defaults to a "default".
func (cfg ActivationConfig) WithRegion(region string) ActivationConfig {
	cfg.region = region
	return cfg
}

// WithSelectMemberFunc sets the function that will be invoked during
// the activation process.
// It will select the member on which the actor will be activated/spawned.
func (cfg ActivationConfig) WithSelectMemberFunc(fn SelectMemberFunc) ActivationConfig {
	cfg.selectMember = fn
	return cfg
}

// SelectMemberFunc will be invoked during the activation process.
// Given the ActivationDetails the actor will be spawned on the returned member.
type SelectMemberFunc func(ActivationDetails) *Member

// ActivationDetails holds detailed information about an activation.
type ActivationDetails struct {
	// Region where the actor should be activated.
	Region string
	// Kind of actor.
	Kind string
	// A slice of members that's pre-filtered by the kind of the actor
	// that needs to be activated.
	Members []*Member
}

// SelectRandomMember selects a random member of the cluster.
func SelectRandomMember(details ActivationDetails) *Member {
	return details.Members[rand.Intn(len(details.Members))]
}
