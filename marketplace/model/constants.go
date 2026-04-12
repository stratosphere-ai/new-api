package model

import (
	"github.com/QuantumNous/new-api/model"
)

// DB alias for the main database connection
var DB = model.DB

const (
	SellerStatusActive    = 1
	SellerStatusSuspended = 2
)

const (
	ListingStatusActive    = 1
	ListingStatusPaused    = 2
	ListingStatusExhausted = 3
)

const (
	WithdrawalStatusPending  = 1
	WithdrawalStatusApproved = 2
	WithdrawalStatusPaid     = 3
	WithdrawalStatusRejected = 4
)

// PlatformFeeRate is the fixed platform commission rate (15%)
const PlatformFeeRate = 0.15

// Marketplace group name used for channel routing
const MarketplaceGroup = "marketplace"

// Provider type constants matching the UI
const (
	ProviderAnthropic = "anthropic"
	ProviderOpenAI    = "openai"
	ProviderGoogle    = "google"
)

// Discount tier thresholds for priority-based routing
const (
	DiscountTierHigh   = 30 // >= 30% off → priority 100
	DiscountTierMedium = 20 // >= 20% off → priority 50
	// < 20% → priority 10
)
