package breaker

import (
	"sync/atomic"
	"time"
)

type CircuitBreaker interface {
	// ReportFailure reports a failure to the circuit breaker, incrementing the consequent error count and penalizing the upstream server if necessary
	ReportFailure()
	// AllowRequest returns true if a request should be allowed, false otherwise
	AllowRequest() bool
	// ReportSuccess reports a success to the circuit breaker, resetting the consequent error count
	ReportSuccess()
	// Penalty returns the duration of the penalty for the upstream server
	Penalty()
	// ReportCanceled reports a canceled request to the circuit breaker
	ReportCanceled()
}

type DefaultCircuitBreaker struct {
	// ConsequentErrorCount is the number of consequent errors to the upstream server
	ConsequentErrorCount atomic.Int32
	// MaxConsequentErrorCount is the maximum number of consequent errors allowed before the upstream server is penalized
	MaxConsequentErrorCount int
	// penaltedUntil is the timestamp until which the upstream server is penalized
	penaltedUntil atomic.Int64
	// penaltiesCount is the number of times the upstream server has been penalized
	penaltiesCount atomic.Int32
	// lastPenaltyTime is the timestamp of the last penalty
	lastPenaltyTime atomic.Int64
	// BasePenaltyDuration is the base duration of the penalty
	BasePenaltyDuration time.Duration
	// MaxPenaltyDuration is the maximum duration of the penalty
	MaxPenaltyDuration time.Duration
	// probeInFlight is the flag indicating whether a probe request is in flight
	probeInFlight atomic.Bool
}

func NewDefaultCircuitBreaker() *DefaultCircuitBreaker {
	return &DefaultCircuitBreaker{
		MaxConsequentErrorCount: 5,
		BasePenaltyDuration:     time.Second,
		MaxPenaltyDuration:      time.Minute,
	}
}

// ReportFailure reports a failure to the circuit breaker, incrementing the consequent error count and penalizing the upstream server if necessary
func (dcb *DefaultCircuitBreaker) ReportFailure() {
	until := dcb.penaltedUntil.Load()
	if until != 0 && time.Now().UnixNano() >= until {
		// Failed probe in Half-Open causes the circuit breaker to go back to Open state
		dcb.Penalty()
		dcb.probeInFlight.Store(false)
		return
	}
	if dcb.ConsequentErrorCount.Add(1) >= int32(dcb.MaxConsequentErrorCount) {
		dcb.ConsequentErrorCount.Store(0)
		dcb.Penalty()
	}
}

// ReportSuccess reports a success to the circuit breaker, resetting the consequent error count and closing the circuit if necessary
func (dcb *DefaultCircuitBreaker) ReportSuccess() {
	dcb.ConsequentErrorCount.Store(0)

	if dcb.penaltedUntil.Load() != 0 {
		// Successful probe in Half-Open -> Closed
		dcb.penaltedUntil.Store(0)
		dcb.probeInFlight.Store(false)
	}

	last := dcb.lastPenaltyTime.Load()
	if last != 0 && time.Since(time.Unix(0, last)) > dcb.BasePenaltyDuration*time.Duration(3) { // penaltyAmnesty is 3 times the base penalty duration
		dcb.penaltiesCount.Store(0)
		dcb.lastPenaltyTime.Store(0)
	}
}

// AllowRequest returns true if the circuit breaker allows a request, false otherwise
func (dcb *DefaultCircuitBreaker) AllowRequest() bool {
	until := dcb.penaltedUntil.Load()
	if until == 0 {
		// closed
		return true
	}
	if time.Now().UnixNano() < until {
		// open
		return false
	}
	// half-open
	return dcb.probeInFlight.CompareAndSwap(false, true)
}

// Penalty applies a penalty to the circuit breaker, increasing the penalty duration and resetting the consequent error count
func (dcb *DefaultCircuitBreaker) Penalty() {
	n := dcb.penaltiesCount.Add(1)
	d := min(dcb.BasePenaltyDuration*time.Duration(n), dcb.MaxPenaltyDuration)
	now := time.Now()
	dcb.lastPenaltyTime.Store(now.UnixNano())
	dcb.penaltedUntil.Store(now.Add(d).UnixNano())
}

// ReportCanceled releases the half-open probe slot without judging
// the upstream: the client canceled the request, so the probe result
// is unknown — the upstream stays in Half-Open awaiting the next probe.
func (dcb *DefaultCircuitBreaker) ReportCanceled() {
	dcb.probeInFlight.Store(false)
}
