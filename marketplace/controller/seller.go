package controller

import (
	"time"

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

// RequestWithdrawal allows a seller to request a withdrawal of their balance
func RequestWithdrawal(c *gin.Context) {
	userId := c.GetInt("id")
	seller, err := mpModel.GetSellerByUserId(userId)
	if err != nil {
		common.ApiErrorMsg(c, "seller not found")
		return
	}
	if seller.Balance <= 0 {
		common.ApiErrorMsg(c, "no balance to withdraw")
		return
	}

	// Atomically take the balance
	amount, err := mpModel.ResetSellerBalance(seller.Id)
	if err != nil || amount <= 0 {
		common.ApiErrorMsg(c, "no balance to withdraw")
		return
	}

	now := time.Now()
	withdrawal := &mpModel.Withdrawal{
		SellerId:    seller.Id,
		Amount:      amount,
		Status:      mpModel.WithdrawalStatusPending,
		PeriodStart: now.AddDate(0, -1, 0),
		PeriodEnd:   now,
	}
	if err := mpModel.CreateWithdrawal(withdrawal); err != nil {
		// Restore balance on failure
		_ = mpModel.IncrementSellerBalance(seller.Id, amount)
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, withdrawal)
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
