package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
)

// Circuit breaker states
const (
	CircuitClosed   = 0 // Normal: all requests pass through
	CircuitOpen     = 1 // Tripped: requests are rejected immediately
	CircuitHalfOpen = 2 // Probing: one request allowed to test recovery
)

// CircuitBreakerConfig holds the configuration for the circuit breaker.
type CircuitBreakerConfig struct {
	FailureThreshold int           // Number of consecutive failures to trip the breaker (default: 5)
	CooldownPeriod   time.Duration // How long to wait before probing (default: 30s)
	HalfOpenMaxProbe int           // Max concurrent probe requests in half-open state (default: 1)
}

var defaultCircuitConfig = CircuitBreakerConfig{
	FailureThreshold: 5,
	CooldownPeriod:   30 * time.Second,
	HalfOpenMaxProbe: 1,
}

// channelCircuit tracks the circuit breaker state for a single channel.
type channelCircuit struct {
	state            int
	consecutiveFails int
	lastFailTime     time.Time
	lastSuccessTime  time.Time
	probeInFlight    int // Number of probe requests currently in flight
	mu               sync.Mutex
}

// circuitBreakerStore manages circuit breakers for all channels.
var circuitBreakerStore sync.Map // map[int]*channelCircuit (channelId → circuit)

// getCircuit returns the circuit breaker for a channel, creating one if needed.
func getCircuit(channelId int) *channelCircuit {
	val, _ := circuitBreakerStore.LoadOrStore(channelId, &channelCircuit{
		state: CircuitClosed,
	})
	return val.(*channelCircuit)
}

// IsChannelAvailable checks if a channel is available (not in open state).
// Returns true if the channel can accept requests.
// In half-open state, allows limited probe requests.
func IsChannelAvailable(channelId int) bool {
	cb := getCircuit(channelId)
	cb.mu.Lock()
	defer cb.mu.Unlock()

	switch cb.state {
	case CircuitClosed:
		return true

	case CircuitOpen:
		// Check if cooldown period has elapsed
		if time.Since(cb.lastFailTime) >= defaultCircuitConfig.CooldownPeriod {
			// Transition to half-open: allow a probe request
			cb.state = CircuitHalfOpen
			cb.probeInFlight = 0
			common.SysLog(fmt.Sprintf("circuit breaker: channel #%d transitioning from OPEN to HALF-OPEN (cooldown elapsed)", channelId))
			// Fall through to half-open logic
		} else {
			return false // Still in cooldown
		}
		fallthrough

	case CircuitHalfOpen:
		// Allow limited probe requests
		if cb.probeInFlight < defaultCircuitConfig.HalfOpenMaxProbe {
			cb.probeInFlight++
			return true
		}
		return false // Already has a probe in flight
	}

	return true
}

// RecordChannelSuccess records a successful request to a channel.
// If the channel was in half-open state, it transitions back to closed.
func RecordChannelSuccess(channelId int) {
	cb := getCircuit(channelId)
	cb.mu.Lock()
	defer cb.mu.Unlock()

	prevState := cb.state

	cb.consecutiveFails = 0
	cb.lastSuccessTime = time.Now()

	if cb.state == CircuitHalfOpen {
		cb.probeInFlight--
		if cb.probeInFlight < 0 {
			cb.probeInFlight = 0
		}
	}

	cb.state = CircuitClosed

	if prevState != CircuitClosed {
		common.SysLog(fmt.Sprintf("circuit breaker: channel #%d recovered → CLOSED (was %s)", channelId, stateString(prevState)))
	}
}

// RecordChannelFailure records a failed request to a channel.
// If consecutive failures exceed the threshold, the circuit trips to open state.
func RecordChannelFailure(channelId int) {
	cb := getCircuit(channelId)
	cb.mu.Lock()
	defer cb.mu.Unlock()

	cb.consecutiveFails++
	cb.lastFailTime = time.Now()

	if cb.state == CircuitHalfOpen {
		// Probe failed: go back to open
		cb.probeInFlight--
		if cb.probeInFlight < 0 {
			cb.probeInFlight = 0
		}
		cb.state = CircuitOpen
		common.SysLog(fmt.Sprintf("circuit breaker: channel #%d probe failed → OPEN (will retry in %v)", channelId, defaultCircuitConfig.CooldownPeriod))
		return
	}

	if cb.consecutiveFails >= defaultCircuitConfig.FailureThreshold {
		cb.state = CircuitOpen
		common.SysLog(fmt.Sprintf("circuit breaker: channel #%d tripped → OPEN (consecutive failures: %d, threshold: %d)", channelId, cb.consecutiveFails, defaultCircuitConfig.FailureThreshold))
	}
}

// GetChannelCircuitState returns the current state of a channel's circuit breaker.
// Useful for monitoring and admin dashboards.
func GetChannelCircuitState(channelId int) (state int, consecutiveFails int, lastFailTime time.Time) {
	cb := getCircuit(channelId)
	cb.mu.Lock()
	defer cb.mu.Unlock()
	return cb.state, cb.consecutiveFails, cb.lastFailTime
}

// ResetChannelCircuit manually resets a channel's circuit breaker to closed state.
// Used by admin to force-recover a channel.
func ResetChannelCircuit(channelId int) {
	cb := getCircuit(channelId)
	cb.mu.Lock()
	defer cb.mu.Unlock()
	prevState := cb.state
	cb.state = CircuitClosed
	cb.consecutiveFails = 0
	cb.probeInFlight = 0
	if prevState != CircuitClosed {
		common.SysLog(fmt.Sprintf("circuit breaker: channel #%d manually reset → CLOSED (was %s)", channelId, stateString(prevState)))
	}
}

// GetAllCircuitStates returns the state of all tracked channels.
func GetAllCircuitStates() map[int]map[string]interface{} {
	result := make(map[int]map[string]interface{})
	circuitBreakerStore.Range(func(key, value interface{}) bool {
		channelId := key.(int)
		cb := value.(*channelCircuit)
		cb.mu.Lock()
		result[channelId] = map[string]interface{}{
			"state":             stateString(cb.state),
			"consecutive_fails": cb.consecutiveFails,
			"last_fail_time":    cb.lastFailTime,
			"last_success_time": cb.lastSuccessTime,
		}
		cb.mu.Unlock()
		return true
	})
	return result
}

func stateString(state int) string {
	switch state {
	case CircuitClosed:
		return "CLOSED"
	case CircuitOpen:
		return "OPEN"
	case CircuitHalfOpen:
		return "HALF-OPEN"
	default:
		return "UNKNOWN"
	}
}
