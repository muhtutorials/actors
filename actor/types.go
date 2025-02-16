package actor

import "context"

type (
	InternalError struct {
		From string
		Err  error
	}

	killProcess struct {
		cancel   context.CancelFunc
		graceful bool
	}

	Initialized struct{}

	Started struct{}

	Stopped struct{}
)
