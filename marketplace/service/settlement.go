package service

import (
	"context"
	"fmt"
	"time"

	"github.com/QuantumNous/new-api/common"
	mpModel "github.com/QuantumNous/new-api/marketplace/model"
	"github.com/QuantumNous/new-api/model"
)

// RecordTrade records a trade and distributes funds between seller and platform.
// Called from the relay post-consume hook.
func RecordTrade(channelId int, buyerUserId int, modelName string, promptTokens int, completionTokens int, retailQuota int64) error {
	listing, err := mpModel.GetListingByChannelId(channelId)
	if err != nil {
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

	// Record the trade
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

	if err := mpModel.CreateTrade(trade); err != nil {
		common.SysError(fmt.Sprintf("failed to record marketplace trade: %v", err))
		return err
	}

	// Credit seller balance
	if sellerAmount > 0 {
		if err := mpModel.IncrementSellerBalance(listing.SellerId, sellerAmount); err != nil {
			common.SysError(fmt.Sprintf("failed to credit seller %d: %v", listing.SellerId, err))
		}
	}

	// Update token usage and check caps
	if err := CheckAndUpdateCaps(listing.Id, int64(totalTokens)); err != nil {
		common.SysError(fmt.Sprintf("failed to update listing caps: %v", err))
	}

	return nil
}

// CheckAndUpdateCaps increments token usage and auto-pauses listing if cap is reached
func CheckAndUpdateCaps(listingId int, tokensUsed int64) error {
	newTotal, err := mpModel.IncrementListingTokensUsed(listingId, tokensUsed)
	if err != nil {
		return err
	}

	listing, err := mpModel.GetListingById(listingId)
	if err != nil {
		return err
	}

	// Check total cap
	if listing.TotalCap > 0 && newTotal >= listing.TotalCap {
		common.SysLog(fmt.Sprintf("listing %d reached total cap (%d/%d), auto-pausing", listingId, newTotal, listing.TotalCap))
		_ = mpModel.UpdateListingStatus(listingId, mpModel.ListingStatusExhausted)
		model.UpdateChannelStatus(listing.ChannelId, "", common.ChannelStatusManuallyDisabled, "marketplace listing exhausted")
	}

	// Hourly and daily caps are checked via Redis (see CheckHourlyDailyCaps)

	return nil
}

// CheckHourlyDailyCaps checks hourly/daily token caps using Redis counters.
// Returns true if the listing should be skipped (cap exceeded).
func CheckHourlyDailyCaps(listingId int, hourlyCap int64, dailyCap int64) bool {
	if hourlyCap <= 0 && dailyCap <= 0 {
		return false
	}

	if !common.RedisEnabled {
		// Without Redis, we can't efficiently track hourly/daily caps
		return false
	}

	ctx := context.Background()
	now := time.Now()

	// Check hourly cap
	if hourlyCap > 0 {
		hourKey := fmt.Sprintf("mp:cap:hourly:%d:%s", listingId, now.Format("2006010215"))
		val, err := common.RDB.Get(ctx, hourKey).Int64()
		if err == nil && val >= hourlyCap {
			return true
		}
	}

	// Check daily cap
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

	// Increment hourly counter
	hourKey := fmt.Sprintf("mp:cap:hourly:%d:%s", listingId, now.Format("2006010215"))
	common.RDB.IncrBy(ctx, hourKey, tokens)
	common.RDB.Expire(ctx, hourKey, 2*time.Hour)

	// Increment daily counter
	dayKey := fmt.Sprintf("mp:cap:daily:%d:%s", listingId, now.Format("20060102"))
	common.RDB.IncrBy(ctx, dayKey, tokens)
	common.RDB.Expire(ctx, dayKey, 48*time.Hour)
}
