package balancer

import (
	"sync/atomic"
)

type RoundRobinBalancer struct {
	// upstreams is the list of upstream servers
	upstreams []*Upstream
	// index is the current index of the round-robin balancer
	index atomic.Uint32
}

func NewRoundRobinBalancer(upstreams []*Upstream) *RoundRobinBalancer {
	return &RoundRobinBalancer{
		upstreams: upstreams,
	}
}

// Next() Implements logic of round-robin balancing
//
// If next upstream is not ready, skips it and tries the next one.
// If no upstream is ready, returns nil.
func (r *RoundRobinBalancer) Next() *Upstream {
	if len(r.upstreams) == 0 {
		return nil
	}
	for i := 0; i < len(r.upstreams); i++ {
		idx := int(r.index.Add(1)-1) % len(r.upstreams)
		upstream := r.upstreams[idx]
		if upstream.AllowRequest() {
			return upstream
		}
	}
	return nil
}
