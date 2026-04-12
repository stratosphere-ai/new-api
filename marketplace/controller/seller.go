package controller

import (
	"github.com/QuantumNous/new-api/common"
	mpModel "github.com/QuantumNous/new-api/marketplace/model"
	mpService "github.com/QuantumNous/new-api/marketplace/service"
	"github.com/gin-gonic/gin"
)

// RegisterSeller registers the current user as a seller
func RegisterSeller(c *gin.Context) {
	userId := c.GetInt("id")
	seller, err := mpService.RegisterSeller(userId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, seller)
}

// GetSellerDashboard returns seller info and listings
func GetSellerDashboard(c *gin.Context) {
	userId := c.GetInt("id")
	dashboard, err := mpService.GetSellerDashboard(userId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, dashboard)
}

// GetSellerWithdrawals returns the seller's withdrawal history
func GetSellerWithdrawals(c *gin.Context) {
	userId := c.GetInt("id")
	seller, err := mpModel.GetSellerByUserId(userId)
	if err != nil {
		common.ApiErrorMsg(c, "seller not found")
		return
	}

	withdrawals, err := mpModel.GetWithdrawalsBySellerId(seller.Id)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, gin.H{
		"withdrawals": withdrawals,
		"balance":     seller.Balance,
		"total_earned": seller.TotalEarned,
	})
}

// GetSellerTrades returns the seller's trade history
func GetSellerTrades(c *gin.Context) {
	userId := c.GetInt("id")
	seller, err := mpModel.GetSellerByUserId(userId)
	if err != nil {
		common.ApiErrorMsg(c, "seller not found")
		return
	}

	trades, err := mpModel.GetSellerRecentTrades(seller.Id, 100)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, trades)
}
