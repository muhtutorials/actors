package cluster

import "github.com/muhtutorials/actors/actor"

// memberToProviderPID creates a new PID from the member that
// will target the provider actor of the cluster.
func memberToProvider(m *Member) *actor.PID {
	return actor.NewPID(m.Host, "provider/"+m.ID)
}

// NewCID returns a new Cluster ID.
func NewCID(pid *actor.PID, id, kind, region string) *CID {
	return &CID{
		PID:    pid,
		ID:     id,
		Kind:   kind,
		Region: region,
	}
}

// Equals checks whether the given CID equals the caller.
func (cid *CID) Equals(other *CID) bool {
	return cid.ID == other.ID && cid.Kind == other.Kind
}

// PID returns the cluster PID of where the node agent can be reached.
func (m *Member) PID() *actor.PID {
	return actor.NewPID(m.Host, "cluster/"+m.ID)
}

func (m *Member) Equals(other *Member) bool {
	return m.ID == other.ID && m.Host == other.Host
}

func (m *Member) HasKind(kind string) bool {
	for _, k := range m.Kinds {
		if kind == k {
			return true
		}
	}
	return false
}
