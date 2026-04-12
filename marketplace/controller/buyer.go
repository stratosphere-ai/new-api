package controller

import (
	"github.com/QuantumNous/new-api/common"
	mpModel "github.com/QuantumNous/new-api/marketplace/model"
	"github.com/QuantumNous/new-api/model"
	"github.com/gin-gonic/gin"
)

// GetBuyerStats returns aggregated usage stats for the current buyer
func GetBuyerStats(c *gin.Context) {
	userId := c.GetInt("id")

	stats, err := mpModel.GetBuyerStats(userId)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	// Also get user's current balance
	user, err := model.GetUserById(userId, false)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, gin.H{
		"stats":   stats,
		"balance": user.Quota,
	})
}

// GetBuyerTrades returns trade history for the current buyer
func GetBuyerTrades(c *gin.Context) {
	userId := c.GetInt("id")
	limit := common.String2Int(c.DefaultQuery("limit", "50"))
	offset := common.String2Int(c.DefaultQuery("offset", "0"))

	if limit <= 0 || limit > 100 {
		limit = 50
	}

	trades, err := mpModel.GetTradesByBuyerUserId(userId, limit, offset)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, trades)
}

// GetBuyerModelUsage returns per-model usage breakdown
func GetBuyerModelUsage(c *gin.Context) {
	userId := c.GetInt("id")

	usage, err := mpModel.GetBuyerModelUsage(userId)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, usage)
}

// GetBuyerAPIInfo returns API usage info for the buyer
func GetBuyerAPIInfo(c *gin.Context) {
	userId := c.GetInt("id")

	// Get user's tokens that have marketplace group access
	tokens, err := model.GetAllUserTokens(userId, 0, 100)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	// Filter tokens with marketplace group
	var mpTokens []gin.H
	for _, t := range tokens {
		if t.Group == mpModel.MarketplaceGroup {
			mpTokens = append(mpTokens, gin.H{
				"id":     t.Id,
				"name":   t.Name,
				"key":    t.Key,
				"status": t.Status,
			})
		}
	}

	common.ApiSuccess(c, gin.H{
		"tokens":   mpTokens,
		"base_url": "/v1",
		"group":    mpModel.MarketplaceGroup,
	})
}

// CreateBuyerAPIToken creates a token with marketplace group for the buyer
func CreateBuyerAPIToken(c *gin.Context) {
	userId := c.GetInt("id")

	key, err := common.GenerateKey()
	if err != nil {
		common.ApiError(c, err)
		return
	}

	token := &model.Token{
		UserId:         userId,
		Name:           "Marketplace Token",
		Key:            key,
		Status:         1,
		UnlimitedQuota: true,
		Group:          mpModel.MarketplaceGroup,
		CreatedTime:    common.GetTimestamp(),
		ExpiredTime:    -1,
	}

	if err := token.Insert(); err != nil {
		common.ApiError(c, err)
		return
	}

	common.ApiSuccess(c, gin.H{
		"id":   token.Id,
		"name": token.Name,
		"key":  token.Key,
	})
}
