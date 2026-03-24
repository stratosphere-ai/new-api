package model

import (
	"errors"
	"time"

	"github.com/QuantumNous/new-api/common"
	"gorm.io/gorm"
)

const (
	MarketplacePurchaseStatusPending   = 0 // Awaiting user confirmation (over budget)
	MarketplacePurchaseStatusCompleted = 1 // Purchase completed (auto-approved or confirmed)
	MarketplacePurchaseStatusExpired   = 2 // Confirmation link expired
	MarketplacePurchaseStatusCancelled = 3 // User cancelled
)

// MarketplacePurchase tracks agent-initiated quota purchases.
type MarketplacePurchase struct {
	Id               int            `json:"id" gorm:"primaryKey"`
	UserId           int            `json:"user_id" gorm:"index"`
	TokenName        string         `json:"token_name" gorm:"type:varchar(50)"`
	Models           string         `json:"models" gorm:"type:varchar(1024);default:''"`
	Quota            int            `json:"quota"`                                         // Quota amount purchased
	ExpireDuration   string         `json:"expire_duration" gorm:"type:varchar(10)"`       // e.g. "30d"
	Status           int            `json:"status" gorm:"default:0;index"`                 // 0=pending, 1=completed, 2=expired, 3=cancelled
	ConfirmToken     string         `json:"confirm_token" gorm:"type:char(32);uniqueIndex"` // For email confirmation link
	TokenId          int            `json:"token_id"`                                      // Created token ID (set after completion)
	TokenKey         string         `json:"token_key" gorm:"type:char(48)"`                // Created token key (set after completion)
	SourceIP         string         `json:"source_ip" gorm:"type:varchar(45)"`             // Request source IP
	CreatedAt        time.Time      `json:"created_at" gorm:"autoCreateTime"`
	CompletedAt      *time.Time     `json:"completed_at"`
	DeletedAt        gorm.DeletedAt `gorm:"index"`
}

func (p *MarketplacePurchase) Insert() error {
	return DB.Create(p).Error
}

func GetMarketplacePurchaseByConfirmToken(token string) (*MarketplacePurchase, error) {
	if token == "" {
		return nil, errors.New("confirmation token is empty")
	}
	var purchase MarketplacePurchase
	err := DB.Where("confirm_token = ?", token).First(&purchase).Error
	return &purchase, err
}

func (p *MarketplacePurchase) UpdateStatus(status int) error {
	updates := map[string]interface{}{
		"status": status,
	}
	if status == MarketplacePurchaseStatusCompleted {
		now := time.Now()
		updates["completed_at"] = &now
	}
	return DB.Model(p).Updates(updates).Error
}

func (p *MarketplacePurchase) SetTokenInfo(tokenId int, tokenKey string) error {
	return DB.Model(p).Updates(map[string]interface{}{
		"token_id":  tokenId,
		"token_key": tokenKey,
	}).Error
}

// GetUserMonthlyMarketplaceSpending returns total quota spent via marketplace this month.
func GetUserMonthlyMarketplaceSpending(userId int) (int64, error) {
	now := time.Now()
	monthStart := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())

	var total int64
	err := DB.Model(&MarketplacePurchase{}).
		Where("user_id = ? AND status = ? AND created_at >= ?",
			userId, MarketplacePurchaseStatusCompleted, monthStart).
		Select("COALESCE(SUM(quota), 0)").
		Scan(&total).Error
	return total, err
}

// ExpirePendingPurchases marks old pending purchases as expired.
// Called periodically or on-demand. Expires purchases older than the given duration.
func ExpirePendingPurchases(maxAge time.Duration) (int64, error) {
	cutoff := time.Now().Add(-maxAge)
	result := DB.Model(&MarketplacePurchase{}).
		Where("status = ? AND created_at < ?", MarketplacePurchaseStatusPending, cutoff).
		Update("status", MarketplacePurchaseStatusExpired)
	return result.RowsAffected, result.Error
}
