package controller

import (
	"fmt"

	"github.com/QuantumNous/new-api/common"
	mpModel "github.com/QuantumNous/new-api/marketplace/model"
	mpService "github.com/QuantumNous/new-api/marketplace/service"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

// --- Platform overview ---

// AdminGetPlatformStats returns platform-wide marketplace statistics
func AdminGetPlatformStats(c *gin.Context) {
	stats, err := mpModel.GetPlatformStats()
	if err != nil {
		common.ApiError(c, err)
		return
	}

	sellers, err := mpModel.GetAllSellers()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	activeSellers := 0
	for _, s := range sellers {
		if s.Status == mpModel.SellerStatusActive {
			activeSellers++
		}
	}

	listings, err := mpModel.GetActiveListings()
	if err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, gin.H{
		"stats":           stats,
		"total_sellers":   len(sellers),
		"active_sellers":  activeSellers,
		"active_listings": len(listings),
	})
}

// --- Seller management ---

// AdminGetSellers returns all sellers with user info
func AdminGetSellers(c *gin.Context) {
	sellers, err := mpModel.GetAllSellers()
	if err != nil {
		common.ApiError(c, err)
		return
	}

	type SellerWithUser struct {
		mpModel.Seller
		Username     string `json:"username"`
		Email        string `json:"email"`
		ListingCount int    `json:"listing_count"`
	}

	var result []SellerWithUser
	for _, s := range sellers {
		swu := SellerWithUser{Seller: *s}
		user, err := model.GetUserById(s.UserId, false)
		if err == nil && user != nil {
			swu.Username = user.Username
			swu.Email = user.Email
		}
		listings, _ := mpModel.GetListingsBySellerId(s.Id)
		swu.ListingCount = len(listings)
		result = append(result, swu)
	}

	common.ApiSuccess(c, result)
}

// AdminGetSellerDetail returns detailed seller info including listings
func AdminGetSellerDetail(c *gin.Context) {
	sellerId := common.String2Int(c.Param("id"))
	if sellerId == 0 {
		common.ApiErrorMsg(c, "invalid seller id")
		return
	}

	seller, err := mpModel.GetSellerById(sellerId)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	listings, _ := mpModel.GetListingsBySellerId(sellerId)
	withdrawals, _ := mpModel.GetWithdrawalsBySellerId(sellerId)

	user, _ := model.GetUserById(seller.UserId, false)
	var username, email string
	if user != nil {
		username = user.Username
		email = user.Email
	}

	common.ApiSuccess(c, gin.H{
		"seller":      seller,
		"username":    username,
		"email":       email,
		"listings":    listings,
		"withdrawals": withdrawals,
	})
}

// AdminSuspendSeller suspends a seller and disables all their listings
func AdminSuspendSeller(c *gin.Context) {
	sellerId := common.String2Int(c.Param("id"))
	if sellerId == 0 {
		common.ApiErrorMsg(c, "invalid seller id")
		return
	}

	if err := mpModel.UpdateSellerStatus(sellerId, mpModel.SellerStatusSuspended); err != nil {
		common.ApiError(c, err)
		return
	}

	// Disable all listings
	listings, _ := mpModel.GetListingsBySellerId(sellerId)
	for _, listing := range listings {
		if listing.Status == mpModel.ListingStatusActive {
			_ = mpModel.UpdateListingStatus(listing.Id, mpModel.ListingStatusPaused)
			model.UpdateChannelStatus(listing.ChannelId, "", common.ChannelStatusManuallyDisabled, "seller suspended by admin")
		}
	}

	common.ApiSuccess(c, nil)
}

// AdminEnableSeller re-enables a suspended seller
func AdminEnableSeller(c *gin.Context) {
	sellerId := common.String2Int(c.Param("id"))
	if sellerId == 0 {
		common.ApiErrorMsg(c, "invalid seller id")
		return
	}

	if err := mpModel.UpdateSellerStatus(sellerId, mpModel.SellerStatusActive); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

// --- Withdrawal management ---

// AdminGetWithdrawals returns all withdrawals with seller info
func AdminGetWithdrawals(c *gin.Context) {
	statusStr := c.DefaultQuery("status", "0")
	status := common.String2Int(statusStr)

	withdrawals, err := mpModel.GetAllWithdrawals(status)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	type WithdrawalWithSeller struct {
		mpModel.Withdrawal
		Username string `json:"username"`
	}

	var result []WithdrawalWithSeller
	for _, w := range withdrawals {
		wws := WithdrawalWithSeller{Withdrawal: *w}
		seller, err := mpModel.GetSellerById(w.SellerId)
		if err == nil {
			user, err := model.GetUserById(seller.UserId, false)
			if err == nil && user != nil {
				wws.Username = user.Username
			}
		}
		result = append(result, wws)
	}

	common.ApiSuccess(c, result)
}

// AdminApproveWithdrawal marks a withdrawal as approved
func AdminApproveWithdrawal(c *gin.Context) {
	id := common.String2Int(c.Param("id"))
	if id == 0 {
		common.ApiErrorMsg(c, "invalid withdrawal id")
		return
	}
	if err := mpModel.UpdateWithdrawalStatus(id, mpModel.WithdrawalStatusApproved); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

// AdminMarkWithdrawalPaid marks a withdrawal as paid
func AdminMarkWithdrawalPaid(c *gin.Context) {
	id := common.String2Int(c.Param("id"))
	if id == 0 {
		common.ApiErrorMsg(c, "invalid withdrawal id")
		return
	}
	if err := mpModel.UpdateWithdrawalStatus(id, mpModel.WithdrawalStatusPaid); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

// AdminRejectWithdrawal rejects a withdrawal and refunds the seller's balance
func AdminRejectWithdrawal(c *gin.Context) {
	id := common.String2Int(c.Param("id"))
	if id == 0 {
		common.ApiErrorMsg(c, "invalid withdrawal id")
		return
	}

	w, err := mpModel.GetWithdrawalById(id)
	if err != nil {
		common.ApiErrorMsg(c, "withdrawal not found")
		return
	}
	if w.Status != mpModel.WithdrawalStatusPending {
		common.ApiErrorMsg(c, "can only reject pending withdrawals")
		return
	}

	if err := mpModel.IncrementSellerBalance(w.SellerId, w.Amount); err != nil {
		common.ApiError(c, fmt.Errorf("failed to refund seller balance: %w", err))
		return
	}
	if err := mpModel.UpdateWithdrawalStatus(id, mpModel.WithdrawalStatusRejected); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

// --- Settlement ---

// AdminTriggerSettlement manually triggers monthly settlement (for testing)
func AdminTriggerSettlement(c *gin.Context) {
	mpService.RunMonthlySettlement()
	common.ApiSuccess(c, nil)
}
