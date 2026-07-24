package balancer

import (
	"net/http"
	"net/url"
	"sld/internal/config"
	"sync/atomic"
)

type RoundRobinBalancer struct {
	upstreams []*Upstream
	index     atomic.Uint32
}

func NewRoundRobinBalancer(cfgUpstreams []config.UpstreamConfig, transport http.RoundTripper) *RoundRobinBalancer {
	upstreams := make([]*Upstream, 0, len(cfgUpstreams))
	for _, upstream := range cfgUpstreams {
		url, _ := url.Parse(upstream.URL)

		upstreams = append(upstreams, NewUpstream(upstream.Name, url, upstream.Weight, transport))
	}
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
