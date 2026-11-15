package internal

import (
	"net/url"
	"sld/internal/balancer"
	"sld/internal/config"
)

// BuildUpstreams builds a slice of Upstream from the given UpstreamConfig slice
func BuildUpstreams(cfgUpstreams []config.UpstreamConfig) []*balancer.Upstream {
	upstreams := make([]*balancer.Upstream, 0, len(cfgUpstreams))
	for _, upstream := range cfgUpstreams {
		url, _ := url.Parse(upstream.URL)

		upstreams = append(upstreams, balancer.NewUpstream(upstream.Name, url, upstream.Weight))
	}
	return upstreams
}
