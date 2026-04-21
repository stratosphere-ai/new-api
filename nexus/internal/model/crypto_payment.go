package model

import "gorm.io/gorm"

// CryptoPayment is a recorded on-chain deposit that tops up a user's quota.
// Produced by payment/crypto.Watcher when a transfer is confirmed.
type CryptoPayment struct {
	gorm.Model
	OrgID         uint   `gorm:"index"`
	UserID        uint   `gorm:"index"`
	Asset         string `gorm:"type:varchar(32)"`  // USDT / USDC
	Chain         string `gorm:"type:varchar(32)"`  // ethereum / tron / bsc / solana
	TxHash        string `gorm:"type:varchar(128);uniqueIndex"`
	FromAddr      string `gorm:"type:varchar(128);index"`
	ToAddr        string `gorm:"type:varchar(128);index"`
	AmountRaw     string `gorm:"type:varchar(64)"`  // base units as decimal string
	AmountUSD     float64
	Confirmations int
	Status        string `gorm:"type:varchar(32);index"` // pending | confirmed | credited | failed
}
