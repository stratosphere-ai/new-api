package service

import (
	"fmt"
	"sync"
	"time"

	"github.com/QuantumNous/new-api/common"
)

var channelRateLimiter common.InMemoryRateLimiter
var channelRateLimiterOnce sync.Once

func initChannelRateLimiter() {
	channelRateLimiterOnce.Do(func() {
		channelRateLimiter.Init(5 * time.Minute)
	})
}

// CheckChannelRPM checks if the channel has exceeded its RPM limit.
// Returns true if the request is allowed, false if rate limited.
func CheckChannelRPM(channelId int, rpmLimit int) bool {
	if rpmLimit <= 0 {
		return true
	}
	initChannelRateLimiter()
	key := fmt.Sprintf("channel_rpm:%d", channelId)
	return channelRateLimiter.Request(key, rpmLimit, 60)
}

// CheckChannelTPM checks if the channel has exceeded its TPM limit using estimated tokens.
// Returns true if the request is allowed, false if rate limited.
func CheckChannelTPM(channelId int, tpmLimit int, estimatedTokens int) bool {
	if tpmLimit <= 0 || estimatedTokens <= 0 {
		return true
	}
	initChannelRateLimiter()
	key := fmt.Sprintf("channel_tpm:%d", channelId)
	// Use token count as the number of "requests" against the limit
	// Each token counts as one unit against the TPM limit
	return channelRateLimiter.Request(key, tpmLimit, 60)
}

// IsChannelRateLimited checks both RPM and TPM limits for a channel.
// Returns true if the channel should be skipped (rate limited).
func IsChannelRateLimited(channelId int, rpmLimit int, tpmLimit int) bool {
	if rpmLimit > 0 && !CheckChannelRPM(channelId, rpmLimit) {
		return true
	}
	return false
}
