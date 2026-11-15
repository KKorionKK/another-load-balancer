package balancer

import (
	"net/url"
	"sld/internal/breaker"
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
	CB     breaker.CircuitBreaker
}

func NewUpstream(name string, url *url.URL, weight int) *Upstream {
	up := Upstream{
		Name:   name,
		URL:    url,
		Weight: weight,
		CB:     breaker.NewDefaultCircuitBreaker(), // TODO: configure circuit breaker
	}
	return &up
}

// ReportSuccess reports a success to the circuit breaker, resetting the consequent error count
func (u *Upstream) ReportSuccess() {
	u.CB.ReportSuccess()
}

// ReportFailure reports a failure to the circuit breaker, incrementing the consequent error count and penalizing the upstream server if necessary
func (u *Upstream) ReportFailure() {
	u.CB.ReportFailure()
}

// ReportCanceled reports a canceled request to the circuit breaker
func (u *Upstream) ReportCanceled() {
	u.CB.ReportCanceled()
}

// AllowRequest returns true if a request should be allowed, false otherwise
func (u *Upstream) AllowRequest() bool {
	return u.CB.AllowRequest()
}

// Balancer is the interface for a load balancer
type Balancer interface {
	// Next returns the next upstream server to use
	Next() *Upstream
}
