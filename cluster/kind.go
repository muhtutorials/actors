package cluster

import "github.com/muhtutorials/actors/actor"

// KindConfig holds configuration for a registered kind.
type KindConfig struct{}

// NewKindConfig returns a default kind configuration.
func NewKindConfig() KindConfig {
	return KindConfig{}
}

// kind is a type of actor that can be activated from any member of the cluster.
type kind struct {
	config   KindConfig
	name     string
	producer actor.Producer
}

// newKind returns a new kind.
func newKind(cfg KindConfig, name string, p actor.Producer) kind {
	return kind{
		config:   cfg,
		name:     name,
		producer: p,
	}
}
