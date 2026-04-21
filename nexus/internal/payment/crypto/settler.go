// Package crypto implements stablecoin top-up settlement (USDT/USDC) across
// Ethereum, Tron, BSC, and Solana. It provides per-user deposit addresses
// and an on-chain watcher that credits the user's quota on confirmation.
package crypto

import (
	"context"

	"github.com/stratosphere-ai/nexus-api/internal/payment"
)

// Settler implements payment.PaymentSettler for crypto top-ups. A "Settle"
// call corresponds to an on-chain deposit the Watcher has observed.
//
// TODO: implement. Settle looks up CryptoPayment by tx_hash, converts
// raw amount to USD via oracle, credits the user's quota, writes a Log row.
type Settler struct{}

func (s *Settler) Settle(ctx context.Context, req payment.SettleRequest) (payment.Receipt, error) {
	_ = ctx
	_ = req
	return payment.Receipt{}, nil
}

func (s *Settler) Refund(ctx context.Context, receiptID string) error {
	_ = ctx
	_ = receiptID
	return nil
}
