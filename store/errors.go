package store

import "errors"

var (
	ErrCapacityLimitReached = errors.New(`store capacity limit reached`)
	ErrClosed               = errors.New(`store already closed`)
)
