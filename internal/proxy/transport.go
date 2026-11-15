package proxy

import (
	"net"
	"net/http"
	"sld/internal"
	"sld/internal/balancer"
	"sld/internal/config"
	"sld/internal/retry"
	"time"
)

// NewTransport defines transport for proxy server
func NewTransport(cfg config.TransportConfig) *http.Transport {
	return &http.Transport{
		MaxIdleConns:          cfg.MaxIdleConns,
		MaxIdleConnsPerHost:   cfg.MaxIdleConnsPerHost,
		IdleConnTimeout:       cfg.IdleConnTimeout,
		TLSHandshakeTimeout:   cfg.TLSHandshakeTimeout,
		ResponseHeaderTimeout: cfg.ResponseHeaderTimeout,
		DialContext: (&net.Dialer{
			Timeout: cfg.DialTimeout,
		}).DialContext,
	}
}

// RetryingTransport wraps http.RoundTripper and implements retry logic for proxy requests
type RetryingTransport struct {
	// base is the underlying http.RoundTripper to use for requests
	base http.RoundTripper
	// balancer is the balancer to use for selecting upstream servers
	balancer balancer.Balancer
	// policy is the retry policy to use for retrying requests
	policy retry.Policy
}

// NewRetryingTransport creates a new RetryingTransport with the given base, balancer, and policy
func NewRetryingTransport(base http.RoundTripper, balancer balancer.Balancer, policy retry.Policy) *RetryingTransport {
	return &RetryingTransport{
		base:     base,
		balancer: balancer,
		policy:   policy,
	}
}

// isUpstreamFailure returns true if the request failed due to an upstream error
func isUpstreamFailure(req *http.Request, resp *http.Response, err error) bool {
	if err != nil {
		return req.Context().Err() == nil
	}
	return resp.StatusCode >= 500
}

// RoundTrip implements http.RoundTripper interface for RetryingTransport
func (t *RetryingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for attempt := 1; ; attempt++ {
		up := t.balancer.Next()
		if up == nil {
			return nil, internal.ErrNoUpstreams
		}

		out := req.Clone(req.Context())
		out.URL.Scheme = up.URL.Scheme
		out.URL.Host = up.URL.Host

		resp, err := t.base.RoundTrip(out)

		switch {
		case isUpstreamFailure(req, resp, err):
			up.ReportFailure()
		case err != nil && req.Context().Err() != nil:
			up.ReportCanceled()
		default:
			up.ReportSuccess()
		}

		delay, retryable := t.policy.Next(attempt, resp, err)

		if !retryable {
			return resp, err
		}
		if resp != nil {
			resp.Body.Close()
		}

		select {
		case <-time.After(delay):
		case <-req.Context().Done():
			return nil, req.Context().Err()
		}
	}
}
