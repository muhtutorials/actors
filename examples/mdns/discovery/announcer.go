package discovery

import "github.com/grandcat/zeroconf"

const (
	serviceName = "_actor_"
	domain      = "local."
	host        = "pc1"
)

type announcer struct {
	id     string
	ip     []string
	port   int
	server *zeroconf.Server
}

func newAnnouncer(opts *opts) *announcer {
	return &announcer{
		id:   opts.id,
		ip:   opts.ip,
		port: opts.port,
	}
}

func (a *announcer) start() {
	server, err := zeroconf.RegisterProxy(
		a.id,
		serviceName,
		domain,
		a.port,
		host,
		a.ip,
		[]string{"txtv=0", "lo=1", "la=2"},
		nil,
	)
	if err != nil {
		panic(err)
	}
	a.server = server
}

func (a *announcer) shutdown() {
	a.server.Shutdown()
}
