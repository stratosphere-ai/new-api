package service

import (
	"context"
	"fmt"
	"time"

	"github.com/QuantumNous/new-api/common"
	mpModel "github.com/QuantumNous/new-api/marketplace/model"
	"github.com/QuantumNous/new-api/model"
	"gorm.io/gorm"
)

// RecordTrade records a trade and distributes funds between seller and platform.
// All DB operations are wrapped in a transaction for atomicity.
// Called from the relay post-consume hook.
func RecordTrade(channelId int, buyerUserId int, modelName string, promptTokens int, completionTokens int, retailQuota int64) error {
	listing := GetListingByChannelIdCached(channelId)
	if listing == nil {
		return nil // not a marketplace channel, skip silently
	}

	totalTokens := promptTokens + completionTokens
	discount := listing.Discount

	// Calculate billing split
	buyerAmount := int64(float64(retailQuota) * (1 - float64(discount)/100.0))
	platformAmount := int64(float64(retailQuota) * mpModel.PlatformFeeRate)
	sellerAmount := buyerAmount - platformAmount
	if sellerAmount < 0 {
		sellerAmount = 0
	}

	trade := &mpModel.Trade{
		ListingId:        listing.Id,
		BuyerUserId:      buyerUserId,
		SellerId:         listing.SellerId,
		ChannelId:        channelId,
		ModelName:        modelName,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      totalTokens,
		Discount:         discount,
		RetailAmount:     retailQuota,
		BuyerAmount:      buyerAmount,
		SellerAmount:     sellerAmount,
		PlatformAmount:   platformAmount,
	}

	// Wrap all DB operations in a transaction
	err := model.DB.Transaction(func(tx *gorm.DB) error {
		// 1. Record the trade
		if err := tx.Create(trade).Error; err != nil {
			return fmt.Errorf("create trade: %w", err)
		}
		// 2. Credit seller balance
		if sellerAmount > 0 {
			result := tx.Model(&mpModel.Seller{}).Where("id = ?", listing.SellerId).
				UpdateColumns(map[string]interface{}{
					"balance":      gorm.Expr("balance + ?", sellerAmount),
					"total_earned": gorm.Expr("total_earned + ?", sellerAmount),
				})
			if result.Error != nil {
				return fmt.Errorf("credit seller: %w", result.Error)
			}
		}
		// 3. Increment listing tokens_used
		result := tx.Model(&mpModel.Listing{}).Where("id = ?", listing.Id).
			Update("tokens_used", gorm.Expr("tokens_used + ?", int64(totalTokens)))
		if result.Error != nil {
			return fmt.Errorf("update tokens_used: %w", result.Error)
		}
		return nil
	})

	if err != nil {
		common.SysError(fmt.Sprintf("marketplace trade transaction failed (channel=%d, buyer=%d): %v", channelId, buyerUserId, err))
		return err
	}

	// Post-transaction: update Redis hourly/daily caps (non-critical, best-effort)
	IncrementHourlyDailyCaps(listing.Id, int64(totalTokens))

	// Check total cap and auto-pause if exceeded
	checkTotalCap(listing, int64(totalTokens))

	return nil
}

// checkTotalCap checks if listing total cap is exceeded and auto-pauses
func checkTotalCap(listing *mpModel.Listing, newTokens int64) {
	if listing.TotalCap <= 0 {
		return
	}
	// Re-read from DB to get accurate tokens_used after transaction
	updated, err := mpModel.GetListingById(listing.Id)
	if err != nil {
		return
	}
	if updated.TokensUsed >= listing.TotalCap {
		common.SysLog(fmt.Sprintf("listing %d reached total cap (%d/%d), auto-pausing", listing.Id, updated.TokensUsed, listing.TotalCap))
		_ = mpModel.UpdateListingStatus(listing.Id, mpModel.ListingStatusExhausted)
		model.UpdateChannelStatus(listing.ChannelId, "", common.ChannelStatusManuallyDisabled, "marketplace listing exhausted")
		InvalidateListingCache(listing.ChannelId)
	}
}

// CheckHourlyDailyCaps checks hourly/daily token caps using Redis counters.
// Returns true if the listing should be skipped (cap exceeded).
func CheckHourlyDailyCaps(listingId int, hourlyCap int64, dailyCap int64) bool {
	if hourlyCap <= 0 && dailyCap <= 0 {
		return false
	}
	if !common.RedisEnabled {
		return false
	}

	ctx := context.Background()
	now := time.Now()

	if hourlyCap > 0 {
		hourKey := fmt.Sprintf("mp:cap:hourly:%d:%s", listingId, now.Format("2006010215"))
		val, err := common.RDB.Get(ctx, hourKey).Int64()
		if err == nil && val >= hourlyCap {
			return true
		}
	}

	if dailyCap > 0 {
		dayKey := fmt.Sprintf("mp:cap:daily:%d:%s", listingId, now.Format("20060102"))
		val, err := common.RDB.Get(ctx, dayKey).Int64()
		if err == nil && val >= dailyCap {
			return true
		}
	}

	return false
}

// IncrementHourlyDailyCaps updates hourly/daily Redis counters after a request completes
func IncrementHourlyDailyCaps(listingId int, tokens int64) {
	if !common.RedisEnabled {
		return
	}

	ctx := context.Background()
	now := time.Now()

	hourKey := fmt.Sprintf("mp:cap:hourly:%d:%s", listingId, now.Format("2006010215"))
	common.RDB.IncrBy(ctx, hourKey, tokens)
	common.RDB.Expire(ctx, hourKey, 2*time.Hour)

	dayKey := fmt.Sprintf("mp:cap:daily:%d:%s", listingId, now.Format("20060102"))
	common.RDB.IncrBy(ctx, dayKey, tokens)
	common.RDB.Expire(ctx, dayKey, 48*time.Hour)
}
