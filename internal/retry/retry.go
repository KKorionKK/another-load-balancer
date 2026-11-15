package retry

import (
	"context"
	"errors"
	"net/http"
	"time"
)

type Policy interface {
	// Next reports whether attempt `attempt` should be retried
	// and how long to wait before the next one.
	Next(attempt int, resp *http.Response, err error) (time.Duration, bool)
}

type BackoffPolicy struct {
	backoff     Backoffer
	maxAttempts int
}

func NewBackoffPolicy(backoff Backoffer, maxAttempts int) *BackoffPolicy {
	return &BackoffPolicy{
		backoff:     backoff,
		maxAttempts: maxAttempts,
	}
}

// Next returns the duration to wait before the next retry and whether to retry
func (p *BackoffPolicy) Next(attempt int, resp *http.Response, err error) (time.Duration, bool) {
	if attempt >= p.maxAttempts {
		return 0, false
	}
	if !isRetryable(resp, err) {
		return 0, false
	}
	return p.backoff.Backoff(attempt), true
}

// isRetryable returns true if the request should be retried
func isRetryable(resp *http.Response, err error) bool {
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			// Context errors are not retryable
			return false
		}
		// EOF, RST, connection refused is not retryable
		return true
	}
	switch resp.StatusCode {
	case http.StatusBadGateway,
		http.StatusInternalServerError,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout:
		return true
	}
	return false
}
