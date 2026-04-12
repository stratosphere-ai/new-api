package service

import (
	"sync"

	mpModel "github.com/QuantumNous/new-api/marketplace/model"
)

// listingCache provides a fast channelId → listing lookup for the hot path.
// Rebuilt on cache miss from DB.
var listingCache sync.Map // channelId (int) → *mpModel.Listing

// GetListingByChannelIdCached returns listing from cache, falling back to DB.
func GetListingByChannelIdCached(channelId int) *mpModel.Listing {
	if v, ok := listingCache.Load(channelId); ok {
		return v.(*mpModel.Listing)
	}
	listing, err := mpModel.GetListingByChannelId(channelId)
	if err != nil {
		return nil
	}
	listingCache.Store(channelId, listing)
	return listing
}

// InvalidateListingCache removes a listing from cache (call after status change)
func InvalidateListingCache(channelId int) {
	listingCache.Delete(channelId)
}

// IsListingCapExceeded checks if a marketplace channel should be skipped
// due to hourly/daily cap limits. Returns true if should skip.
func IsListingCapExceeded(channelId int) bool {
	listing := GetListingByChannelIdCached(channelId)
	if listing == nil {
		return false // not a marketplace channel
	}
	if listing.Status != mpModel.ListingStatusActive {
		return true // paused or exhausted
	}
	return CheckHourlyDailyCaps(listing.Id, listing.HourlyCap, listing.DailyCap)
}
