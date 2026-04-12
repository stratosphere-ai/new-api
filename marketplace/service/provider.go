package service

import (
	"github.com/QuantumNous/new-api/constant"
	mpModel "github.com/QuantumNous/new-api/marketplace/model"
)

// ProviderConfig holds the channel type and default models for a provider
type ProviderConfig struct {
	ChannelType int
	Models      string // comma-separated
}

var providerConfigs = map[string]ProviderConfig{
	mpModel.ProviderAnthropic: {
		ChannelType: constant.ChannelTypeAnthropic,
		Models:      "claude-sonnet-4-20250514,claude-opus-4-20250514,claude-haiku-4-5-20251001,claude-3-5-sonnet-20241022,claude-3-5-haiku-20241022",
	},
	mpModel.ProviderOpenAI: {
		ChannelType: constant.ChannelTypeOpenAI,
		Models:      "gpt-4o,gpt-4o-mini,gpt-4-turbo,gpt-4,gpt-3.5-turbo,o1,o1-mini,o3-mini",
	},
	mpModel.ProviderGoogle: {
		ChannelType: constant.ChannelTypeGemini,
		Models:      "gemini-2.5-pro,gemini-2.5-flash,gemini-2.0-flash,gemini-1.5-pro,gemini-1.5-flash",
	},
}

// GetProviderConfig returns the config for a given provider type
func GetProviderConfig(providerType string) (ProviderConfig, bool) {
	config, ok := providerConfigs[providerType]
	return config, ok
}

// IsValidProvider checks if the provider type is supported
func IsValidProvider(providerType string) bool {
	_, ok := providerConfigs[providerType]
	return ok
}

// DiscountToPriority maps discount percentage to channel priority using tiered strategy
func DiscountToPriority(discount int) int64 {
	if discount >= mpModel.DiscountTierHigh {
		return 100
	}
	if discount >= mpModel.DiscountTierMedium {
		return 50
	}
	return 10
}

// DiscountToWeight maps discount to channel weight within same priority tier
func DiscountToWeight(discount int) uint {
	if discount <= 0 {
		return 1
	}
	return uint(discount)
}
