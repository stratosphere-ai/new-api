package controller

import (
	"fmt"

	"github.com/QuantumNous/new-api/common"
	mpModel "github.com/QuantumNous/new-api/marketplace/model"
	mpService "github.com/QuantumNous/new-api/marketplace/service"
	"github.com/gin-gonic/gin"
)

// AdminGetWithdrawals returns all pending withdrawals for admin review
func AdminGetWithdrawals(c *gin.Context) {
	status := c.DefaultQuery("status", "pending")
	var withdrawals []*mpModel.Withdrawal
	var err error

	switch status {
	case "pending":
		withdrawals, err = mpModel.GetPendingWithdrawals()
	default:
		withdrawals, err = mpModel.GetPendingWithdrawals()
	}

	if err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, withdrawals)
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

	// Get withdrawal to know the amount and seller
	withdrawals, err := mpModel.GetPendingWithdrawals()
	if err != nil {
		common.ApiError(c, err)
		return
	}

	var target *mpModel.Withdrawal
	for _, w := range withdrawals {
		if w.Id == id {
			target = w
			break
		}
	}
	if target == nil {
		common.ApiErrorMsg(c, "withdrawal not found or not pending")
		return
	}

	// Refund the balance
	if err := mpModel.IncrementSellerBalance(target.SellerId, target.Amount); err != nil {
		common.ApiError(c, fmt.Errorf("failed to refund seller balance: %w", err))
		return
	}

	if err := mpModel.UpdateWithdrawalStatus(id, mpModel.WithdrawalStatusRejected); err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, nil)
}

// AdminGetSellers returns all sellers for admin view
func AdminGetSellers(c *gin.Context) {
	sellers, err := mpModel.GetAllSellers()
	if err != nil {
		common.ApiError(c, err)
		return
	}
	common.ApiSuccess(c, sellers)
}

// AdminTriggerSettlement manually triggers monthly settlement (for testing)
func AdminTriggerSettlement(c *gin.Context) {
	mpService.RunMonthlySettlement()
	common.ApiSuccess(c, nil)
}
