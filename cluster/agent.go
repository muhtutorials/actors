package cluster

import (
	"github.com/muhtutorials/actors/actor"
	"log/slog"
	"reflect"
)

type (
	activate struct {
		config ActivationConfig
		kind   string
	}
	deactivate struct {
		pid *actor.PID
	}
	getActive struct {
		id string
	}
	getKinds   struct{}
	getMembers struct{}
)

type Agent struct {
	cluster    *Cluster
	kinds      map[string]bool
	localKinds map[string]kind
	members    *MemberSet
	// All the actors that are available cluster-wide.
	activated map[string]*actor.PID
}

func NewAgent(c *Cluster) actor.Producer {
	kinds := make(map[string]bool)
	localKinds := make(map[string]kind)
	for _, k := range c.kinds {
		kinds[k.name] = true
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
	case actor.Started:
	case actor.Stopped:
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
	case activate:
		pid := a.activate(msg.config, msg.kind)
		ctx.Respond(pid)
	case deactivate:
		a.broadcast(&Deactivation{PID: msg.pid})
	case getActive:
		pid := a.activated[msg.id]
		ctx.Respond(pid)
	case getKinds:
		a.handleGetKinds(ctx)
	case getMembers:
		ctx.Respond(a.members.Slice())
	}
}

func (a *Agent) handleActorTopology(msg *ActorTopology) {
	for _, actorInfo := range msg.Actors {
		a.addActivated(actorInfo.PID)
	}
}

func (a *Agent) handleMembers(members []*Member) {
	joined := NewMemberSet(members...).Except(a.members.Slice())
	left := a.members.Except(members)
	for _, member := range joined {
		a.memberJoin(member)
	}
	for _, member := range left {
		a.memberLeave(member)
	}
}

// A new kind is activated on this cluster.
func (a *Agent) handleActivation(msg *Activation) {
	a.addActivated(msg.PID)
	a.cluster.engine.BroadcastEvent(EventActivation{PID: msg.PID})
}

func (a *Agent) handleDeactivation(msg *Deactivation) {
	a.removeActivated(msg.PID)
	a.cluster.engine.Poison(msg.PID)
	a.cluster.engine.BroadcastEvent(EventDeactivation{PID: msg.PID})
}

func (a *Agent) handleActivationRequest(msg *ActivationRequest) *ActivationResponse {
	if !a.hasLocalKind(msg.Kind) {
		slog.Error("received activation request but kind not registered locally on this node", "kind", msg.Kind)
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
	if cfg.selectMember == nil {
		cfg.selectMember = SelectRandomMember
	}
	memberPID := cfg.selectMember(ActivationDetails{
		Region:  cfg.region,
		Members: members,
	})
	if memberPID == nil {
		slog.Warn("activator did not find a member to activate on")
		return nil
	}
	req := &ActivationRequest{ID: cfg.id, Kind: kind}
	activatorPID := actor.NewPID(memberPID.Host, "cluster"+memberPID.ID)
	var activationResp *ActivationResponse
	// Local activation.
	if memberPID.Host == a.cluster.engine.Address() {
		activationResp = a.handleActivationRequest(req)
	} else {
		resp, err := a.cluster.engine.Request(activatorPID, req, a.cluster.config.requestTimeout).Result()
		if err != nil {
			slog.Error("failed activation request", "err", err)
			return nil
		}
		r, ok := resp.(*ActivationResponse)
		if !ok {
			slog.Error("expected `*ActivationResponse`", "msg", reflect.TypeOf(resp))
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

func (a *Agent) memberJoin(member *Member) {
	a.members.Add(member)
	// Track cluster-wide available kinds.
	for _, k := range member.Kinds {
		if _, ok := a.kinds[k]; !ok {
			a.kinds[k] = true
		}
	}
	var actorInfos []*ActorInfo
	for _, pid := range a.activated {
		actorInfo := &ActorInfo{PID: pid}
		actorInfos = append(actorInfos, actorInfo)
	}
	// Send our ActorTopology to this member
	if len(actorInfos) > 0 {
		a.cluster.engine.Send(member.PID(), &ActorTopology{Actors: actorInfos})
	}
	// Broadcast EventMemberJoin
	a.cluster.engine.BroadcastEvent(EventMemberJoin{Member: member})
	slog.Debug(
		"[CLUSTER] member joined",
		"id", member.ID, "host",
		member.Host, "region",
		member.Region, "kinds",
		member.Kinds,
	)
}

func (a *Agent) memberLeave(member *Member) {
	a.members.Remove(member)
	a.rebuildKinds()
	// Remove all the active kinds that were running on the member that left the cluster.
	for _, pid := range a.activated {
		if pid.Address == member.Host {
			a.removeActivated(pid)
		}
	}
	a.cluster.engine.BroadcastEvent(EventMemberLeave{Member: member})
	slog.Debug(
		"[CLUSTER] member left",
		"id", member.ID, "host",
		member.Host, "region",
		member.Region, "kinds",
		member.Kinds,
	)
}

func (a *Agent) rebuildKinds() {
	clear(a.kinds)
	a.members.ForEach(func(m *Member) bool {
		for _, k := range m.Kinds {
			if _, ok := a.kinds[k]; !ok {
				a.kinds[k] = true
			}
		}
		return true
	})
}

func (a *Agent) hasLocalKind(name string) bool {
	_, ok := a.localKinds[name]
	return ok
}

func (a *Agent) broadcast(msg any) {
	a.members.ForEach(func(member *Member) bool {
		a.cluster.engine.Send(member.PID(), msg)
		return true
	})
}
