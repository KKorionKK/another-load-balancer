package balancer

import (
	"net/http"
	"sync/atomic"
)

type RoundRobinBalancer struct {
	upstreams []*Upstream
	index     atomic.Uint32
}

func NewRoundRobinBalancer(upstreams []*Upstream, transport http.RoundTripper) *RoundRobinBalancer {
	return &RoundRobinBalancer{
		upstreams: upstreams,
	}
}

// Next() Implements logic of round-robin balancing
func (r *RoundRobinBalancer) Next() *Upstream {
	indx := r.index.Add(1)
	return r.upstreams[indx%uint32(len(r.upstreams))]
}

// ServeHTTP dispatches the request to the next upstream
func (rrb *RoundRobinBalancer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	upstream := rrb.Next()
	if upstream == nil {
		http.Error(w, "no healthy upstreams", http.StatusServiceUnavailable)
		return
	}
	upstream.ActiveConns.Add(1)
	defer upstream.ActiveConns.Add(-1)
	upstream.Proxy.ServeHTTP(w, r)
}
