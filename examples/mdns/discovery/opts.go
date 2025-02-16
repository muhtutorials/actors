package discovery

import (
	"fmt"
	"time"
)

type opts struct {
	// engine global id
	id string
	// engine's ip to listen on
	ip []string
	// engine's port to accept connections
	port int
}

type optFunc func(opts *opts)

func applyOpts(optFns ...optFunc) *opts {
	options := &opts{
		id: fmt.Sprintf("engine_%d", time.Now().UnixNano()),
	}
	for _, fn := range optFns {
		fn(options)
	}
	return options
}

// withAnnounceAddr specifies engine's ip and port information to announce
func withAnnounceAddr(ip string, p int) optFunc {
	return func(opts *opts) {
		opts.ip = append(opts.ip, ip)
		opts.port = p
	}
}
