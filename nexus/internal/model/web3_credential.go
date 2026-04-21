package model

import (
	"time"

	"gorm.io/gorm"
)

// Web3Credential links a wallet address to a User. A user may link multiple
// wallets across different chains.
type Web3Credential struct {
	gorm.Model
	UserID      uint   `gorm:"index"`
	Chain       string `gorm:"type:varchar(32);index:ix_chain_addr,priority:1"`
	Address     string `gorm:"type:varchar(128);index:ix_chain_addr,priority:2,unique"`
	Nonce       string `gorm:"type:varchar(128)"`
	LastSignAt  *time.Time
	Verified    bool  `gorm:"default:false"`
	IsPrimary   bool  `gorm:"default:false"`
}
