package integration

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sld/internal/balancer"
	"sld/internal/config"
	"sld/internal/proxy"
	"sync"
	"sync/atomic"
	"testing"
)

func TestLoadBalancer_RoundRobin_EvenDistribution(t *testing.T) {
	const numBackends = 3
	var counts [numBackends]atomic.Int64

	transport := proxy.NewTransport(config.TransportConfig{MaxIdleConnsPerHost: 50})

	upstreams := make([]*balancer.Upstream, numBackends)
	for i := range numBackends {
		backend := newMockBackend(t, func() { counts[i].Add(1) })
		upstreams[i] = newTestUpstream(t, fmt.Sprintf("backend-%d", i), backend.URL, transport)
	}

	bal := balancer.NewRoundRobinBalancer(upstreams, transport)
	lb := newTestLoadBalancer(t, bal)

	const totalRequests = 300
	var wg sync.WaitGroup
	for range totalRequests {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := http.Get(lb.URL)
			if err != nil {
				t.Errorf("request failed: %v", err)
				return
			}
			resp.Body.Close()
		}()
	}
	wg.Wait()

	want := int64(totalRequests / numBackends)
	tolerance := int64(float64(want) * 0.1)
	for i, c := range counts {
		if got := c.Load(); got < want-tolerance || got > want+tolerance {
			t.Errorf("backend %d: got %d requests, want ~%d (±%d)", i, got, want, tolerance)
		}
	}

	// stats
	// for i, c := range counts {
	// 	log.Printf("backend %d: %d requests", i, c.Load())
	// }
}

func TestLoadBalancer_RoundRobin_OneServerDown(t *testing.T) {
	const numBackends = 4
	const backendDown = 1
	var counts [numBackends]atomic.Int64

	backends := make([]*httptest.Server, numBackends)

	transport := proxy.NewTransport(config.TransportConfig{MaxIdleConnsPerHost: 50})

	upstreams := make([]*balancer.Upstream, numBackends)
	for i := range numBackends {
		backend := newMockBackend(t, func() { counts[i].Add(1) })
		upstreams[i] = newTestUpstream(t, fmt.Sprintf("backend-%d", i), backend.URL, transport)
		backends[i] = backend
	}

	backends[backendDown].Close()

	bal := balancer.NewRoundRobinBalancer(upstreams, transport)
	lb := newTestLoadBalancer(t, bal)

	const totalRequests = 400
	var wg sync.WaitGroup
	for range totalRequests {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp, err := http.Get(lb.URL)
			if err != nil {
				t.Errorf("request failed: %v", err)
				return
			}
			resp.Body.Close()
		}()
	}
	wg.Wait()

	want := int64(totalRequests / 4)
	tolerance := int64(float64(want) * 0.1)
	for i, c := range counts {
		if i == backendDown {
			if got := c.Load(); got != 0 {
				t.Errorf("backend %d: got %d requests, want 0", i, got)
			}
			continue
		}

		if got := c.Load(); got < want-tolerance || got > want+tolerance {
			t.Errorf("backend %d: got %d requests, want ~%d (±%d)", i, got, want, tolerance)
		}
	}

	// stats
	// for i, c := range counts {
	// 	log.Printf("backend %d: %d requests", i, c.Load())
	// }
}
