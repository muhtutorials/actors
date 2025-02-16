package cluster

import (
	"github.com/muhtutorials/actors/actor"
	"log/slog"
	"reflect"
)

type Agent struct {
	cluster    *Cluster
	kinds      map[string]struct{}
	localKinds map[string]kind
	members    *MemberSet
	// All the actors that are available cluster-wide.
	activated map[string]*actor.PID
}

func NewAgent(c *Cluster) actor.Producer {
	kinds := make(map[string]struct{})
	localKinds := make(map[string]kind)
	for _, k := range c.kinds {
		kinds[k.name] = struct{}{}
		localKinds[k.name] = k
	}
	return func() actor.Receiver {
		return &Agent{
			cluster:    c,
			kinds:      kinds,
			localKinds: localKinds,
			members:    NewMemberSet(),
			activated:  make(map[string]*actor.PID),
		}
	}
}

func (a *Agent) Receive(ctx *actor.Context) {
	switch msg := ctx.Message().(type) {
	case *ActorTopology:
		a.handleActorTopology(msg)
	case *Members:
		a.handleMembers(msg.Members)
	case *Activation:
		a.handleActivation(msg)
	case *Deactivation:
		a.handleDeactivation(msg)
	case *ActivationRequest:
		resp := a.handleActivationRequest(msg)
		ctx.Respond(resp)
	case Activate:
		pid := a.activate(msg.Config, msg.Kind)
		ctx.Respond(pid)
	case Deactivate:
		a.broadcast(&Deactivation{PID: msg.PID})
	case GetActive:
		pid := a.activated[msg.ID]
		ctx.Respond(pid)
	case GetKinds:
		a.handleGetKinds(ctx)
	case GetMembers:
		ctx.Respond(a.members.Slice())
	}
}

func (a *Agent) handleActorTopology(msg *ActorTopology) {
	for _, actorInfo := range msg.Actors {
		a.addActivated(actorInfo.PID)
	}
}

func (a *Agent) handleMembers(members []*Member) {
	joined := NewMemberSet(members...).Difference(a.members.Slice())
	left := a.members.Difference(members)
	for _, member := range joined {
		a.addMember(member)
	}
	for _, member := range left {
		a.removeMember(member)
	}
}

// A new kind is activated on this cluster.
func (a *Agent) handleActivation(msg *Activation) {
	a.addActivated(msg.PID)
	a.cluster.engine.BroadcastEvent(ActivationEvent{PID: msg.PID})
}

func (a *Agent) handleDeactivation(msg *Deactivation) {
	a.removeActivated(msg.PID)
	a.cluster.engine.Kill(msg.PID)
	a.cluster.engine.BroadcastEvent(DeactivationEvent{PID: msg.PID})
}

func (a *Agent) handleActivationRequest(msg *ActivationRequest) *ActivationResponse {
	if !a.isLocalKind(msg.Kind) {
		slog.Error(
			"received activation request but kind not registered locally on this cluster",
			"kind", msg.Kind,
		)
		return &ActivationResponse{Success: false}
	}
	k := a.localKinds[msg.Kind]
	pid := a.cluster.engine.Spawn(k.producer, msg.Kind, actor.WithID(msg.ID))
	return &ActivationResponse{
		PID:     pid,
		Success: true,
	}
}

func (a *Agent) activate(cfg ActivationConfig, kind string) *actor.PID {
	members := a.members.FilterByKind(kind)
	if len(members) == 0 {
		slog.Warn("could not find any members with kind", "kind", kind)
		return nil
	}
	selectedMember := cfg.selectMember(ActivationDetails{
		Region:  cfg.region,
		Members: members,
	})
	if selectedMember == nil {
		slog.Warn("activator did not find a member to activate on")
		return nil
	}
	req := &ActivationRequest{ID: cfg.id, Kind: kind}
	activatorPID := actor.NewPID(selectedMember.Address, "cluster/"+selectedMember.ID)
	var activationResp *ActivationResponse
	// Local activation.
	if selectedMember.Address == a.cluster.engine.Address() {
		activationResp = a.handleActivationRequest(req)
	} else {
		resp, err := a.cluster.engine.
			Request(activatorPID, req, a.cluster.config.requestTimeout).
			Result()
		if err != nil {
			slog.Error("failed activation request", "err", err)
			return nil
		}
		r, ok := resp.(*ActivationResponse)
		if !ok {
			slog.Error(
				"expected `*ActivationResponse`",
				"got", reflect.TypeOf(resp),
			)
			return nil
		}
		if !r.Success {
			slog.Error("activation unsuccessful", "msg", r)
			return nil
		}
		activationResp = r
	}
	a.broadcast(&Activation{PID: activationResp.PID})
	return activationResp.PID
}

func (a *Agent) handleGetKinds(ctx *actor.Context) {
	kinds := make([]string, len(a.kinds))
	i := 0
	for k := range a.kinds {
		kinds[i] = k
		i++
	}
	ctx.Respond(kinds)
}

func (a *Agent) addActivated(pid *actor.PID) {
	if _, ok := a.activated[pid.ID]; !ok {
		a.activated[pid.ID] = pid
		slog.Debug("new actor available on cluster", "pid", pid)
	}
}

func (a *Agent) removeActivated(pid *actor.PID) {
	delete(a.activated, pid.ID)
	slog.Debug("actor removed from cluster", "pid", pid)
}

func (a *Agent) addMember(member *Member) {
	a.members.Add(member)
	// Track cluster-wide available kinds.
	for _, k := range member.Kinds {
		if _, ok := a.kinds[k]; !ok {
			a.kinds[k] = struct{}{}
		}
	}
	var actorInfoSlice []*ActorInfo
	for _, pid := range a.activated {
		actorInfo := &ActorInfo{PID: pid}
		actorInfoSlice = append(actorInfoSlice, actorInfo)
	}
	// Send our "ActorTopology" to this member.
	if len(actorInfoSlice) > 0 {
		a.cluster.engine.Send(member.PID(), &ActorTopology{Actors: actorInfoSlice})
	}
	// Broadcast "MemberJoinedEvent".
	a.cluster.engine.BroadcastEvent(MemberJoinedEvent{Member: member})
	slog.Debug(
		"[CLUSTER] member joined",
		"addr", member.Address,
		"id", member.ID,
		"region", member.Region,
		"kinds", member.Kinds,
	)
}

func (a *Agent) removeMember(member *Member) {
	a.members.Remove(member)
	a.rebuildKinds()
	// Remove all the active kinds that were running on the member that left the cluster.
	for _, pid := range a.activated {
		if pid.Address == member.Address {
			a.removeActivated(pid)
		}
	}
	a.cluster.engine.BroadcastEvent(MemberLeftEvent{Member: member})
	slog.Debug(
		"[CLUSTER] member left",
		"addr", member.Address,
		"id", member.ID,
		"region", member.Region,
		"kinds", member.Kinds,
	)
}

func (a *Agent) rebuildKinds() {
	clear(a.kinds)
	a.members.ForEach(func(member *Member) bool {
		for _, k := range member.Kinds {
			if _, ok := a.kinds[k]; !ok {
				a.kinds[k] = struct{}{}
			}
		}
		return true
	})
}

func (a *Agent) isLocalKind(name string) bool {
	_, ok := a.localKinds[name]
	return ok
}

func (a *Agent) broadcast(msg any) {
	a.members.ForEach(func(member *Member) bool {
		a.cluster.engine.Send(member.PID(), msg)
		return true
	})
}
