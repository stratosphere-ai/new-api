package controller

import (
	"fmt"

	"github.com/QuantumNous/new-api/common"
	mpModel "github.com/QuantumNous/new-api/marketplace/model"
	mpService "github.com/QuantumNous/new-api/marketplace/service"
	"github.com/gin-gonic/gin"
)

// CreateListing creates a new listing for the current seller
func CreateListing(c *gin.Context) {
	userId := c.GetInt("id")

	// Get seller
	seller, err := mpModel.GetSellerByUserId(userId)
	if err != nil {
		common.ApiErrorMsg(c, "please register as a seller first")
		return
	}

	var req mpService.CreateListingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiError(c, err)
		return
	}

	// Validate provider type
	if !mpService.IsValidProvider(req.ProviderType) {
		common.ApiErrorMsg(c, fmt.Sprintf("unsupported provider: %s", req.ProviderType))
		return
	}

	// Validate API key
	if err := mpService.ValidateAPIKey(req.ProviderType, req.APIKey); err != nil {
		common.ApiError(c, fmt.Errorf("API key validation failed: %w", err))
		return
	}

	listing, err := mpService.CreateListing(seller.Id, &req)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, listing)
}

// PauseListing pauses a listing
func PauseListing(c *gin.Context) {
	userId := c.GetInt("id")
	listingId := common.String2Int(c.Param("id"))
	if listingId == 0 {
		common.ApiErrorMsg(c, "invalid listing id")
		return
	}

	seller, err := mpModel.GetSellerByUserId(userId)
	if err != nil {
		common.ApiErrorMsg(c, "seller not found")
		return
	}

	if err := mpService.PauseListing(listingId, seller.Id); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

// ResumeListing resumes a paused listing
func ResumeListing(c *gin.Context) {
	userId := c.GetInt("id")
	listingId := common.String2Int(c.Param("id"))
	if listingId == 0 {
		common.ApiErrorMsg(c, "invalid listing id")
		return
	}

	seller, err := mpModel.GetSellerByUserId(userId)
	if err != nil {
		common.ApiErrorMsg(c, "seller not found")
		return
	}

	if err := mpService.ResumeListing(listingId, seller.Id); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

// DeleteListing deletes a listing
func DeleteListing(c *gin.Context) {
	userId := c.GetInt("id")
	listingId := common.String2Int(c.Param("id"))
	if listingId == 0 {
		common.ApiErrorMsg(c, "invalid listing id")
		return
	}

	seller, err := mpModel.GetSellerByUserId(userId)
	if err != nil {
		common.ApiErrorMsg(c, "seller not found")
		return
	}

	if err := mpService.DeleteListingForSeller(listingId, seller.Id); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

// GetMarketInfo returns market info for buyers (discount range, etc.)
func GetMarketInfo(c *gin.Context) {
	minDiscount, maxDiscount, err := mpModel.GetMarketDiscountRange()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, gin.H{
		"min_discount": minDiscount,
		"max_discount": maxDiscount,
	})
}
