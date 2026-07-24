package internal

import (
	"net/http"
	"net/url"
	"sld/internal/balancer"
	"sld/internal/config"
)

func BuildUpstreams(cfgUpstreams []config.UpstreamConfig, transport http.RoundTripper) []*balancer.Upstream {
	upstreams := make([]*balancer.Upstream, 0, len(cfgUpstreams))
	for _, upstream := range cfgUpstreams {
		url, _ := url.Parse(upstream.URL)

		upstreams = append(upstreams, balancer.NewUpstream(upstream.Name, url, upstream.Weight, transport))
	}
	return upstreams
}
