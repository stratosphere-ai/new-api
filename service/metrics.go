package service

import (
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// MetricsEnabled is set to true when ENABLE_PROMETHEUS=true.
var MetricsEnabled = false

var (
	requestDuration = promauto.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "newapi_request_duration_seconds",
		Help:    "Request duration by channel, model, and status code",
		Buckets: []float64{0.1, 0.25, 0.5, 1, 2.5, 5, 10, 30, 60},
	}, []string{"channel_id", "model", "status_code"})

	requestErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "newapi_request_errors_total",
		Help: "Request errors by channel and error type",
	}, []string{"channel_id", "error_type"})

	tokensTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "newapi_tokens_total",
		Help: "Tokens processed by channel, model, and direction",
	}, []string{"channel_id", "model", "type"})

	circuitBreakerState = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "newapi_circuit_breaker_state",
		Help: "Circuit breaker state (0=closed, 1=open, 2=half_open)",
	}, []string{"channel_id"})

	retryTotal = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "newapi_retry_total",
		Help: "Retry attempts by channel",
	}, []string{"channel_id"})

	activeRequests = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "newapi_active_requests",
		Help: "Currently active relay requests",
	})

	channelRateLimitRejected = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "newapi_channel_rate_limit_rejected_total",
		Help: "Channel rate limit rejections by channel and limit type",
	}, []string{"channel_id", "limit_type"})
)

// ObserveRequestDuration records request latency.
func ObserveRequestDuration(channelId int, model string, statusCode int, durationSeconds float64) {
	if !MetricsEnabled {
		return
	}
	requestDuration.WithLabelValues(
		strconv.Itoa(channelId), model, strconv.Itoa(statusCode),
	).Observe(durationSeconds)
}

// RecordRequestError increments the error counter.
func RecordRequestError(channelId int, errorType string) {
	if !MetricsEnabled {
		return
	}
	requestErrors.WithLabelValues(
		strconv.Itoa(channelId), errorType,
	).Inc()
}

// RecordTokenUsage records prompt and completion token usage.
func RecordTokenUsage(channelId int, model string, promptTokens int, completionTokens int) {
	if !MetricsEnabled {
		return
	}
	chId := strconv.Itoa(channelId)
	if promptTokens > 0 {
		tokensTotal.WithLabelValues(chId, model, "prompt").Add(float64(promptTokens))
	}
	if completionTokens > 0 {
		tokensTotal.WithLabelValues(chId, model, "completion").Add(float64(completionTokens))
	}
}

// SetCircuitBreakerMetricState updates the circuit breaker gauge for a channel.
func SetCircuitBreakerMetricState(channelId int, state CircuitState) {
	if !MetricsEnabled {
		return
	}
	circuitBreakerState.WithLabelValues(strconv.Itoa(channelId)).Set(float64(state))
}

// RecordRetry increments the retry counter for a channel.
func RecordRetry(channelId int) {
	if !MetricsEnabled {
		return
	}
	retryTotal.WithLabelValues(strconv.Itoa(channelId)).Inc()
}

// IncActiveRequests increments active request gauge.
func IncActiveRequests() {
	if !MetricsEnabled {
		return
	}
	activeRequests.Inc()
}

// DecActiveRequests decrements active request gauge.
func DecActiveRequests() {
	if !MetricsEnabled {
		return
	}
	activeRequests.Dec()
}

// RecordChannelRateLimitRejection records a rate limit rejection.
func RecordChannelRateLimitRejection(channelId int, limitType string) {
	if !MetricsEnabled {
		return
	}
	channelRateLimitRejected.WithLabelValues(strconv.Itoa(channelId), limitType).Inc()
}
