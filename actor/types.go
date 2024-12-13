package actor

import "sync"

type (
	InternalError struct {
		From string
		Err  error
	}

	poisonPill struct {
		wg       *sync.WaitGroup
		graceful bool
	}

	Initialized struct{}

	Started struct{}

	Stopped struct{}
)
