package service

import (
	"fmt"

	"github.com/QuantumNous/new-api/common"
	mpModel "github.com/QuantumNous/new-api/marketplace/model"
	"github.com/QuantumNous/new-api/model"
)

// CreateListingRequest represents the request to create a new listing
type CreateListingRequest struct {
	ProviderType string `json:"provider_type" binding:"required"`
	APIKey       string `json:"api_key" binding:"required"`
	Discount     int    `json:"discount" binding:"required,min=1,max=90"`
	TotalCap     int64  `json:"total_cap"`
	HourlyCap    int64  `json:"hourly_cap"`
	DailyCap     int64  `json:"daily_cap"`
	Label        string `json:"label"`
}

// CreateListing creates a new listing and its associated channel
func CreateListing(sellerId int, req *CreateListingRequest) (*mpModel.Listing, error) {
	// Validate provider
	providerConfig, ok := GetProviderConfig(req.ProviderType)
	if !ok {
		return nil, fmt.Errorf("unsupported provider: %s", req.ProviderType)
	}

	// Encrypt the API key
	encryptedKey, err := common.EncryptAPIKey(req.APIKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt API key: %w", err)
	}

	// Get seller info for naming
	seller, err := mpModel.GetSellerById(sellerId)
	if err != nil {
		return nil, fmt.Errorf("seller not found: %w", err)
	}

	priority := DiscountToPriority(req.Discount)
	weight := DiscountToWeight(req.Discount)

	// Create the channel
	channel := &model.Channel{
		Type:        providerConfig.ChannelType,
		Key:         encryptedKey,
		Status:      common.ChannelStatusEnabled,
		Name:        fmt.Sprintf("mp-seller-%d-%s", seller.Id, req.Label),
		Models:      providerConfig.Models,
		Group:       mpModel.MarketplaceGroup,
		Priority:    common.GetPointer(priority),
		Weight:      common.GetPointer(weight),
		CreatedTime: common.GetTimestamp(),
	}

	if err := channel.Insert(); err != nil {
		return nil, fmt.Errorf("failed to create channel: %w", err)
	}

	// Create the listing
	listing := &mpModel.Listing{
		SellerId:     sellerId,
		ChannelId:    channel.Id,
		ProviderType: req.ProviderType,
		Discount:     req.Discount,
		TotalCap:     req.TotalCap,
		HourlyCap:    req.HourlyCap,
		DailyCap:     req.DailyCap,
		Label:        req.Label,
		Status:       mpModel.ListingStatusActive,
	}

	if err := mpModel.CreateListing(listing); err != nil {
		return nil, fmt.Errorf("failed to create listing: %w", err)
	}

	return listing, nil
}

// PauseListing pauses a listing and disables its channel
func PauseListing(listingId int, sellerId int) error {
	listing, err := mpModel.GetListingById(listingId)
	if err != nil {
		return fmt.Errorf("listing not found: %w", err)
	}
	if listing.SellerId != sellerId {
		return fmt.Errorf("unauthorized: listing does not belong to this seller")
	}
	if err := mpModel.UpdateListingStatus(listingId, mpModel.ListingStatusPaused); err != nil {
		return err
	}
	model.UpdateChannelStatus(listing.ChannelId, "", common.ChannelStatusManuallyDisabled, "marketplace listing paused")
	return nil
}

// ResumeListing resumes a paused listing and enables its channel
func ResumeListing(listingId int, sellerId int) error {
	listing, err := mpModel.GetListingById(listingId)
	if err != nil {
		return fmt.Errorf("listing not found: %w", err)
	}
	if listing.SellerId != sellerId {
		return fmt.Errorf("unauthorized: listing does not belong to this seller")
	}
	if listing.Status == mpModel.ListingStatusExhausted {
		return fmt.Errorf("listing is exhausted, adjust caps to resume")
	}
	if err := mpModel.UpdateListingStatus(listingId, mpModel.ListingStatusActive); err != nil {
		return err
	}
	model.UpdateChannelStatus(listing.ChannelId, "", common.ChannelStatusEnabled, "")
	return nil
}

// DeleteListingForSeller deletes a listing and its channel
func DeleteListingForSeller(listingId int, sellerId int) error {
	listing, err := mpModel.GetListingById(listingId)
	if err != nil {
		return fmt.Errorf("listing not found: %w", err)
	}
	if listing.SellerId != sellerId {
		return fmt.Errorf("unauthorized: listing does not belong to this seller")
	}
	// Disable channel first
	model.UpdateChannelStatus(listing.ChannelId, "", common.ChannelStatusManuallyDisabled, "marketplace listing deleted")
	return mpModel.DeleteListing(listingId)
}
