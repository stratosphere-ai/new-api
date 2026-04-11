package service

import (
	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/model"
)

// RecordMarketplaceEarning checks if the channel used for a relay request
// is a marketplace-listed channel, and if so, records the earning for the seller.
// This should be called after UpdateChannelUsedQuota in every post-consume path.
func RecordMarketplaceEarning(channelId int, quotaConsumed int) {
	if quotaConsumed <= 0 {
		return
	}

	// Check if this channel has a marketplace listing
	listing, err := model.GetChannelListingByChannelId(channelId)
	if err != nil {
		return // Not a marketplace channel, nothing to do
	}

	if listing.Status != model.ListingStatusActive {
		return // Listing is paused or under review
	}

	// Check monthly cap
	if listing.MonthlyCapQuota > 0 {
		consumed, _ := model.GetListingMonthlyConsumed(listing.Id)
		if consumed+int64(quotaConsumed) > listing.MonthlyCapQuota {
			common.SysLog("marketplace: listing #%d monthly cap reached, disabling channel")
			// Disable the channel temporarily
			model.UpdateChannelStatus(channelId, "", common.ChannelStatusManuallyDisabled, "marketplace monthly cap reached")
			return
		}
	}

	// Record earning
	_ = model.AddEarning(
		listing.Id,
		listing.UserId,
		listing.ChannelId,
		int64(quotaConsumed),
		listing.RevenueSharePct,
	)
}
