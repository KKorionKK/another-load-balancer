package integration

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"sld/internal/balancer"
	"testing"
)

func newMockBackend(t *testing.T, onRequest func()) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if onRequest != nil {
			onRequest()
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func newTestUpstream(t *testing.T, name, backendURL string, transport http.RoundTripper) *balancer.Upstream {
	t.Helper()
	u, err := url.Parse(backendURL)
	if err != nil {
		t.Fatalf("parse backend url %q: %v", backendURL, err)
	}
	return balancer.NewUpstream(name, u, 0, transport)
}

func newTestLoadBalancer(t *testing.T, bal balancer.Balancer) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upstream := bal.Next()
		if upstream == nil {
			http.Error(w, "no healthy upstreams", http.StatusServiceUnavailable)
			return
		}
		upstream.Proxy.ServeHTTP(w, r)
	}))
	t.Cleanup(srv.Close)
	return srv
}
