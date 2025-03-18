package cluster

import (
	"fmt"
	"github.com/muhtutorials/actors/actor"
	"github.com/muhtutorials/actors/remote"
	"log/slog"
	"math"
	"math/rand"
	"reflect"
	"time"
)

// Should be a reasonable timeout so long distance nodes could work.
var defaultRequestTimeout = time.Second

// Producer is a function that produces an actor.Producer.
type Producer func(c *Cluster) actor.Producer

// Config holds the cluster configuration.
type Config struct {
	listenAddr     string
	id             string
	region         string
	engine         *actor.Engine
	provider       Producer
	requestTimeout time.Duration
}

// NewConfig returns a Config that is initialized with default values.
func NewConfig() Config {
	return Config{
		listenAddr:     getRandomListenAddr(),
		id:             getRandomID(),
		region:         "default",
		provider:       NewProvider(NewProviderConfig()),
		requestTimeout: defaultRequestTimeout,
	}
}

// WithListenAddr sets the listen address of the underlying remote.
// Defaults to a random port number.
func (cfg Config) WithListenAddr(addr string) Config {
	cfg.listenAddr = addr
	return cfg
}

// WithID sets the ID of this cluster.
// Defaults to a randomly generated ID.
func (cfg Config) WithID(id string) Config {
	cfg.id = id
	return cfg
}

// WithRegion sets the region where the member will be hosted.
// Defaults to "default".
func (cfg Config) WithRegion(region string) Config {
	cfg.region = region
	return cfg
}

// WithEngine sets the internal actor engine that will be used
// to power the actors running on the cluster.
// If no engine is given the cluster will instantiate a new
// engine and remote.
func (cfg Config) WithEngine(e *actor.Engine) Config {
	cfg.engine = e
	return cfg
}

// WithProvider sets the cluster's provider.
// Defaults to the Provider.
func (cfg Config) WithProvider(p Producer) Config {
	cfg.provider = p
	return cfg
}

// WithRequestTimeout sets the maximum amount of time a request
// can take between members of the cluster.
// Defaults to one second to support communication between nodes in
// other regions.
func (cfg Config) WithRequestTimeout(d time.Duration) Config {
	cfg.requestTimeout = d
	return cfg
}

// Cluster allows to write distributed actors. It combines Engine, Remote, and
// Provider which allow members of the cluster to send messages to each other in a
// self discovering environment.
type Cluster struct {
	config      Config
	engine      *actor.Engine
	agentPID    *actor.PID
	providerPID *actor.PID
	isStarted   bool
	kinds       []kind
}

// New returns a new cluster given a Config.
func New(cfg Config) (*Cluster, error) {
	if cfg.engine == nil {
		rem := remote.New(cfg.listenAddr, remote.NewConfig())
		engine, err := actor.NewEngine(actor.NewEngineConfig().WithRemote(rem))
		if err != nil {
			return nil, err
		}
		cfg.engine = engine
	}
	return &Cluster{config: cfg, engine: cfg.engine}, nil
}

// Start the cluster.
func (c *Cluster) Start() {
	c.agentPID = c.engine.Spawn(NewAgent(c), "cluster", actor.WithID(c.config.id))
	c.providerPID = c.engine.Spawn(c.config.provider(c), "provider", actor.WithID(c.config.id))
	c.isStarted = true
}

// Stop will shut down the cluster killing all its actors.
func (c *Cluster) Stop() {
	<-c.engine.Kill(c.agentPID).Done()
	<-c.engine.Kill(c.providerPID).Done()
	// todo: should "c.isStarted" be set to false?
}

// Spawn spawns an actor locally with cluster awareness.
func (c *Cluster) Spawn(p actor.Producer, id string, optFns ...actor.OptFunc) *actor.PID {
	pid := c.engine.Spawn(p, id, optFns...)
	for _, member := range c.Members() {
		c.engine.Send(member.PID(), &Activation{PID: pid})
	}
	return pid
}

// Members returns all the members that are part of the cluster.
func (c *Cluster) Members() []*Member {
	resp, err := c.engine.Request(c.agentPID, GetMembers{}, c.config.requestTimeout).Result()
	if err != nil {
		return nil
	}
	if members, ok := resp.([]*Member); ok {
		return members
	}
	return nil
}

// Member returns the member info of the cluster.
func (c *Cluster) Member() *Member {
	kinds := make([]string, len(c.kinds))
	for i := 0; i < len(c.kinds); i++ {
		kinds[i] = c.kinds[i].name
	}
	return &Member{
		Address: c.engine.Address(),
		ID:      c.config.id,
		Region:  c.config.region,
		Kinds:   kinds,
	}
}

// Activate activates the registered kind in the cluster based on the given config.
// The actor does not need to be registered locally on the member if at least one
// member has that kind registered.
func (c *Cluster) Activate(cfg ActivationConfig, kind string) *actor.PID {
	msg := Activate{
		Config: cfg,
		Kind:   kind,
	}
	resp, err := c.engine.Request(c.agentPID, msg, c.config.requestTimeout).Result()
	if err != nil {
		slog.Error("activation failed", "error", err)
		return nil
	}
	pid, ok := resp.(*actor.PID)
	if !ok {
		slog.Warn("expected activation response of '*actor.PID'", "got", reflect.TypeOf(resp))
		return nil
	}
	return pid
}

// Deactivate deactivates the given PID.
func (c *Cluster) Deactivate(pid *actor.PID) {
	c.engine.Send(c.agentPID, Deactivate{PID: pid})
}

// GetActivated gets the activated actor
func (c *Cluster) GetActivated(id string) *actor.PID {
	resp, err := c.engine.Request(c.agentPID, GetActivated{ID: id}, c.config.requestTimeout).Result()
	if err != nil {
		return nil
	}
	if pid, ok := resp.(*actor.PID); ok {
		return pid
	}
	return nil
}

// RegisterKind registers a new actor that can be activated from
// any member in the cluster.
// NOTE: Kinds can only be registered before the cluster is started.
func (c *Cluster) RegisterKind(cfg KindConfig, kind string, p actor.Producer) {
	if c.isStarted {
		slog.Warn("failed to register kind", "reason", "cluster already started", "kind", kind)
		return
	}
	c.kinds = append(c.kinds, newKind(cfg, kind, p))
}

// HasLocalKind returns true whether the cluster has the kind locally registered.
func (c *Cluster) HasLocalKind(name string) bool {
	for _, k := range c.kinds {
		if k.name == name {
			return true
		}
	}
	return false
}

// HasKind returns true whether the given kind is available for activation on
// the cluster.
func (c *Cluster) HasKind(name string) bool {
	resp, err := c.engine.Request(c.agentPID, GetKinds{}, c.config.requestTimeout).Result()
	if err != nil {
		return false
	}
	if kinds, ok := resp.([]string); ok {
		for _, k := range kinds {
			if k == name {
				return true
			}
		}
	}
	return false
}

// ID returns the ID of the cluster.
func (c *Cluster) ID() string {
	return c.config.id
}

// Region return the region of the cluster.
func (c *Cluster) Region() string {
	return c.config.region
}

// Engine returns the actor engine.
func (c *Cluster) Engine() *actor.Engine {
	return c.engine
}

// PID returns the reachable actor process id, which is the Agent actor.
func (c *Cluster) PID() *actor.PID {
	return c.agentPID
}

// Address returns the address of the cluster.
func (c *Cluster) Address() string {
	return c.agentPID.Address
}

func getRandomListenAddr() string {
	return fmt.Sprintf("127.0.0.1:%d", rand.Intn(50000)+10000)
}

func getRandomID() string {
	return fmt.Sprintf("%d", rand.Intn(math.MaxInt))
}
