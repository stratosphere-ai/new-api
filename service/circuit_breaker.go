package service

import (
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/operation_setting"
)

type CircuitState int

const (
	CircuitStateClosed   CircuitState = 0
	CircuitStateOpen     CircuitState = 1
	CircuitStateHalfOpen CircuitState = 2
)

type CircuitBreaker struct {
	mu              sync.Mutex
	channelId       int
	state           CircuitState
	failures        int64
	total           int64
	halfOpenSuccess int64
	halfOpenProbes  int64
	windowStart     time.Time
	lastStateChange time.Time
}

var circuitBreakers sync.Map // channelId -> *CircuitBreaker

func getOrCreateCircuitBreaker(channelId int) *CircuitBreaker {
	if v, ok := circuitBreakers.Load(channelId); ok {
		return v.(*CircuitBreaker)
	}
	cb := &CircuitBreaker{
		channelId:       channelId,
		state:           CircuitStateClosed,
		windowStart:     time.Now(),
		lastStateChange: time.Now(),
	}
	actual, _ := circuitBreakers.LoadOrStore(channelId, cb)
	return actual.(*CircuitBreaker)
}

func (cb *CircuitBreaker) getConfig() *operation_setting.CircuitBreakerSetting {
	return operation_setting.GetCircuitBreakerSetting()
}

// Allow checks whether a request to this channel should be permitted.
func (cb *CircuitBreaker) Allow() bool {
	cfg := cb.getConfig()
	if !cfg.Enabled {
		return true
	}

	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitStateClosed:
		return true
	case CircuitStateOpen:
		if time.Since(cb.lastStateChange) > time.Duration(cfg.OpenDurationSeconds)*time.Second {
			cb.transitionTo(CircuitStateHalfOpen)
			cb.halfOpenSuccess = 0
			cb.halfOpenProbes = 0
			return true // first probe
		}
		return false
	case CircuitStateHalfOpen:
		if cb.halfOpenProbes < int64(cfg.HalfOpenMaxProbes) {
			cb.halfOpenProbes++
			return true
		}
		return false
	}
	return true
}

// RecordSuccess records a successful request.
func (cb *CircuitBreaker) RecordSuccess() {
	cfg := cb.getConfig()
	if !cfg.Enabled {
		return
	}

	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitStateClosed:
		cb.total++
		cb.maybeResetWindow(cfg)
	case CircuitStateHalfOpen:
		cb.halfOpenSuccess++
		if cb.halfOpenSuccess >= int64(cfg.HalfOpenSuccessThreshold) {
			cb.transitionTo(CircuitStateClosed)
			cb.resetCounters()
		}
	case CircuitStateOpen:
		// should not happen, but reset to closed if it does
		cb.transitionTo(CircuitStateClosed)
		cb.resetCounters()
	}
}

// RecordFailure records a failed request.
func (cb *CircuitBreaker) RecordFailure() {
	cfg := cb.getConfig()
	if !cfg.Enabled {
		return
	}

	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitStateClosed:
		cb.maybeResetWindow(cfg)
		cb.failures++
		cb.total++
		if cb.shouldTrip(cfg) {
			cb.transitionTo(CircuitStateOpen)
			common.SysLog("circuit breaker opened for channel, will recover after " +
				time.Duration(cfg.OpenDurationSeconds).String() + "s")
		}
	case CircuitStateHalfOpen:
		// probe failed, back to open
		cb.transitionTo(CircuitStateOpen)
	case CircuitStateOpen:
		// already open, ignore
	}
}

func (cb *CircuitBreaker) shouldTrip(cfg *operation_setting.CircuitBreakerSetting) bool {
	if cb.failures >= int64(cfg.FailureThreshold) {
		return true
	}
	if cb.total > 0 && float64(cb.failures)/float64(cb.total) >= cfg.FailureRateThreshold {
		return cb.failures >= 3 // require at least 3 failures to avoid tripping on low volume
	}
	return false
}

func (cb *CircuitBreaker) maybeResetWindow(cfg *operation_setting.CircuitBreakerSetting) {
	if time.Since(cb.windowStart) > time.Duration(cfg.WindowSeconds)*time.Second {
		cb.resetCounters()
	}
}

func (cb *CircuitBreaker) resetCounters() {
	cb.failures = 0
	cb.total = 0
	cb.halfOpenSuccess = 0
	cb.halfOpenProbes = 0
	cb.windowStart = time.Now()
}

func (cb *CircuitBreaker) transitionTo(state CircuitState) {
	cb.state = state
	cb.lastStateChange = time.Now()
	SetCircuitBreakerMetricState(cb.channelId, state)
}

// State returns the current circuit breaker state.
func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state
}

// IsCircuitBreakerTripped checks if the circuit breaker for a channel is open (rejecting requests).
func IsCircuitBreakerTripped(channelId int) bool {
	cb := getOrCreateCircuitBreaker(channelId)
	return !cb.Allow()
}

// RecordCircuitBreakerSuccess records a successful request for a channel.
func RecordCircuitBreakerSuccess(channelId int) {
	cb := getOrCreateCircuitBreaker(channelId)
	cb.RecordSuccess()
}

// RecordCircuitBreakerFailure records a failed request for a channel (transient errors only).
func RecordCircuitBreakerFailure(channelId int) {
	cb := getOrCreateCircuitBreaker(channelId)
	cb.RecordFailure()
}

// PruneCircuitBreakers removes circuit breakers for channels that no longer exist.
func PruneCircuitBreakers(activeChannelIds map[int]bool) {
	circuitBreakers.Range(func(key, value any) bool {
		channelId := key.(int)
		if !activeChannelIds[channelId] {
			circuitBreakers.Delete(channelId)
		}
		return true
	})
}

// GetCircuitBreakerState returns the state for a channel (for metrics).
func GetCircuitBreakerState(channelId int) CircuitState {
	if v, ok := circuitBreakers.Load(channelId); ok {
		return v.(*CircuitBreaker).State()
	}
	return CircuitStateClosed
}

// RegisterChannelCacheHooks sets up the function variables in the model package
// for circuit breaker filtering, rate limiting, and cache sync callbacks.
func RegisterChannelCacheHooks() {
	model.ChannelShouldSkip = func(channelId int, rpmLimit int, tpmLimit int) bool {
		if IsCircuitBreakerTripped(channelId) {
			return true
		}
		if IsChannelRateLimited(channelId, rpmLimit, tpmLimit) {
			return true
		}
		return false
	}
	model.ChannelCacheSyncCallback = func(activeChannelIds map[int]bool) {
		PruneCircuitBreakers(activeChannelIds)
	}
}
