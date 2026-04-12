package operation_setting

import (
	"github.com/QuantumNous/new-api/setting/config"
)

type CircuitBreakerSetting struct {
	Enabled                  bool    `json:"enabled"`                     // 是否启用熔断器
	FailureThreshold         int     `json:"failure_threshold"`           // 触发熔断的失败次数
	FailureRateThreshold     float64 `json:"failure_rate_threshold"`      // 触发熔断的失败率
	WindowSeconds            int     `json:"window_seconds"`              // 滑动窗口时长（秒）
	OpenDurationSeconds      int     `json:"open_duration_seconds"`       // 熔断持续时长（秒）
	HalfOpenMaxProbes        int     `json:"half_open_max_probes"`        // 半开状态最大探测数
	HalfOpenSuccessThreshold int     `json:"half_open_success_threshold"` // 半开状态成功阈值
}

var circuitBreakerSetting = CircuitBreakerSetting{
	Enabled:                  false,
	FailureThreshold:         5,
	FailureRateThreshold:     0.5,
	WindowSeconds:            60,
	OpenDurationSeconds:      30,
	HalfOpenMaxProbes:        3,
	HalfOpenSuccessThreshold: 2,
}

func init() {
	config.GlobalConfig.Register("circuit_breaker_setting", &circuitBreakerSetting)
}

func GetCircuitBreakerSetting() *CircuitBreakerSetting {
	return &circuitBreakerSetting
}
