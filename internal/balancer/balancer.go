package balancer

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sld/internal/proxy"
	"sync/atomic"
)

// Upstream represents a single upstream server
type Upstream struct {
	// Name is the name of the upstream server
	Name string
	// URL is the URL of the upstream server
	URL *url.URL
	// Alive indicates whether the upstream server is alive
	Alive atomic.Bool
	// ActiveConns is the number of active connections to the upstream server
	ActiveConns atomic.Int64
	// Weight is the weight of the upstream server, need for weighted round-robin balancing
	Weight int
	// Proxy is the reverse proxy for the upstream server
	Proxy *httputil.ReverseProxy
}

func NewUpstream(name string, url *url.URL, weight int, transport http.RoundTripper) *Upstream {
	up := Upstream{
		Name:   name,
		URL:    url,
		Weight: weight,
	}

	up.Proxy = proxy.New(url, transport, func(err error) {
		log.Printf("[%s] proxy error: %s", up.Name, err)
		up.Alive.Store(false)
	})

	return &up
}

// Balancer is the interface for a load balancer
type Balancer interface {
	// Next returns the next upstream server to use
	Next() *Upstream
	ServeHTTP(http.ResponseWriter, *http.Request)
}
