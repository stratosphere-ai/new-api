package model

import (
	"time"

	"gorm.io/gorm"
)

// X402Payment records one per-request x402 settlement. Linked 1:1 to a
// Log row (the relay request that triggered it) via LogID.
type X402Payment struct {
	gorm.Model
	OrgID             uint   `gorm:"index"`
	TokenID           uint   `gorm:"index"`
	RequestID         string `gorm:"type:varchar(64);index"`
	PaymentIntentHash string `gorm:"type:varchar(128);uniqueIndex"`
	Asset             string `gorm:"type:varchar(32)"`
	Chain             string `gorm:"type:varchar(32)"`
	AmountRaw         string `gorm:"type:varchar(64)"`
	AmountUSD         float64
	Facilitator       string `gorm:"type:varchar(128)"`
	Status            string `gorm:"type:varchar(32);index"` // pending | settled | failed
	SettledAt         *time.Time
	SettlementTxID    uint `gorm:"index"` // set when batched into a SettlementTx
	LogID             uint `gorm:"index"`
}

// SettlementTx groups multiple X402Payment rows into one on-chain transaction
// for gas efficiency.
type SettlementTx struct {
	gorm.Model
	Chain     string `gorm:"type:varchar(32)"`
	TxHash    string `gorm:"type:varchar(128);uniqueIndex"`
	AmountUSD float64
	Count     int
	Status    string `gorm:"type:varchar(32);index"`
}
