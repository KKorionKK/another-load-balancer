package internal

import "errors"

// ErrNoUpstreams is returned when there are no upstreams available
var ErrNoUpstreams = errors.New("no upstreams")
