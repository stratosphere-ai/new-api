package payment

import (
	"context"
	"time"
)

// SettleRequest describes a single settlement operation.
type SettleRequest struct {
	OrgID     uint
	UserID    uint
	TokenID   uint
	RequestID string
	AmountUSD float64
	Asset     string // "USDT", "USDC", "stripe", "epay", ...
	Metadata  map[string]string
}

// Receipt is the settlement outcome returned by Settle.
type Receipt struct {
	ID        string
	Status    string
	SettledAt time.Time
	TxHash    string // on-chain hash when applicable
	Raw       map[string]any
}

// PaymentSettler abstracts any settlement provider: Stripe, epay, CreEM,
// crypto deposit, or x402 per-request settlement.
type PaymentSettler interface {
	Settle(ctx context.Context, req SettleRequest) (Receipt, error)
	Refund(ctx context.Context, receiptID string) error
}
