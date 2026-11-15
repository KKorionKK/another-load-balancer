package retry

import (
	"math"
	"math/rand/v2"
	"time"
)

type Backoffer interface {
	// Backoff returns the duration to wait before retry number `attempt` (1-based).
	Backoff(attempt int) time.Duration
}

// ExponentialBackoff applies an exponential backoff strategy without jitter.
type ExponentialBackoff struct {
	startTimeout float64
	maxTimeout   float64
}

func NewExponentialBackoff(startTimeout, maxTimeout float64) *ExponentialBackoff {
	return &ExponentialBackoff{
		startTimeout: startTimeout,
		maxTimeout:   maxTimeout,
	}
}

// Backoff returns the duration to wait before the next retry.
func (eb *ExponentialBackoff) Backoff(attempt int) time.Duration {
	return time.Duration(min(eb.maxTimeout, eb.startTimeout*math.Pow(2, float64(attempt-1))) * float64(time.Second))
}

// ExponentialBackoffWithJitter applies jitter to the exponential backoff duration.
type ExponentialBackoffWithJitter struct {
	startTimeout float64
	maxTimeout   float64
	jitterFactor float64
}

func NewExponentialBackoffWithJitter(startTimeout, maxTimeout, jitterFactor float64) *ExponentialBackoffWithJitter {
	return &ExponentialBackoffWithJitter{
		startTimeout: startTimeout,
		maxTimeout:   maxTimeout,
		jitterFactor: jitterFactor,
	}
}

// Backoff returns the duration to wait before the next retry, with jitter applied.
func (eb *ExponentialBackoffWithJitter) Backoff(attempt int) time.Duration {
	backoff := min(
		eb.maxTimeout,
		eb.startTimeout*math.Pow(2, float64(attempt-1)),
	) * float64(time.Second)

	randomShift := (rand.Float64()*2 - 1) * eb.jitterFactor

	return time.Duration(backoff * (1 + randomShift))
}
