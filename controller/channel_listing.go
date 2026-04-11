package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"

	"github.com/gin-gonic/gin"
)

// SubmitChannelRequest is the request body for submitting a channel to the marketplace.
type SubmitChannelRequest struct {
	Name           string   `json:"name" binding:"required"`        // Display name
	Description    string   `json:"description"`                    // Short description
	Type           int      `json:"type" binding:"required"`        // Channel type (1=OpenAI, 15=Claude, etc.)
	Key            string   `json:"key" binding:"required"`         // The actual API key
	BaseURL        string   `json:"base_url"`                       // Custom base URL (optional)
	Models         []string `json:"models" binding:"required"`      // Models this key supports
	Group          string   `json:"group"`                          // Target group (default: "default")
	MonthlyCapUSD  float64  `json:"monthly_cap_usd"`                // Monthly cap in USD (0=unlimited)
	PricePerMToken float64  `json:"price_per_m_token"`              // Asking price per 1M tokens
}

// SubmitChannelToMarketplace allows a user to submit their API key as a channel.
// The key is stored as a regular channel (leveraging all existing routing),
// and a ChannelListing record tracks ownership and revenue sharing.
//
// POST /api/marketplace/channels
func SubmitChannelToMarketplace(c *gin.Context) {
	var req SubmitChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	userId := c.GetInt("id")

	// Validate
	req.Name = strings.TrimSpace(req.Name)
	if len(req.Name) == 0 || len(req.Name) > 100 {
		common.ApiErrorMsg(c, "Name must be 1-100 characters")
		return
	}
	if len(req.Key) == 0 {
		common.ApiErrorMsg(c, "API key is required")
		return
	}
	if len(req.Models) == 0 {
		common.ApiErrorMsg(c, "At least one model is required")
		return
	}
	if req.Group == "" {
		req.Group = "default"
	}

	// Limit listings per user
	existing, _ := model.GetChannelListingsByUserId(userId)
	if len(existing) >= 20 {
		common.ApiErrorMsg(c, "Maximum 20 channel listings per user")
		return
	}

	// Create the actual channel (this makes it part of the routing pool)
	modelsStr := strings.Join(req.Models, ",")
	weight := uint(50) // Default weight for user-submitted channels
	priority := int64(0)
	autoBan := 1

	var baseURL *string
	if req.BaseURL != "" {
		baseURL = &req.BaseURL
	}

	tag := "marketplace" // Tag all marketplace channels for easy filtering
	channelName := fmt.Sprintf("[MKT] %s (#%d)", req.Name, userId)

	channel := model.Channel{
		Type:        req.Type,
		Key:         req.Key,
		Status:      common.ChannelStatusEnabled,
		Name:        channelName,
		Weight:      &weight,
		CreatedTime: common.GetTimestamp(),
		BaseURL:     baseURL,
		Models:      modelsStr,
		Group:       req.Group,
		Priority:    &priority,
		AutoBan:     &autoBan,
		Tag:         &tag,
	}

	if err := channel.Insert(); err != nil {
		common.SysLog("marketplace: failed to insert channel: " + err.Error())
		common.ApiErrorMsg(c, "Failed to create channel")
		return
	}

	// Calculate monthly cap in quota units
	monthlyCapQuota := int64(0)
	if req.MonthlyCapUSD > 0 {
		monthlyCapQuota = int64(req.MonthlyCapUSD * common.QuotaPerUnit)
	}

	// Create the listing record
	listing := &model.ChannelListing{
		UserId:          userId,
		ChannelId:       channel.Id,
		Status:          model.ListingStatusActive,
		RevenueSharePct: 85, // Seller gets 85%, platform keeps 15%
		MonthlyCapQuota: monthlyCapQuota,
		Models:          modelsStr,
		ChannelType:     req.Type,
		DisplayName:     req.Name,
		Description:     req.Description,
		PricePerMToken:  req.PricePerMToken,
	}

	if err := listing.Insert(); err != nil {
		// Rollback: delete the channel
		channel.Delete()
		common.SysLog("marketplace: failed to insert listing: " + err.Error())
		common.ApiErrorMsg(c, "Failed to create listing")
		return
	}

	// Refresh channel cache so the new channel is immediately available for routing
	model.InitChannelCache()

	common.SysLog(fmt.Sprintf("marketplace: user %d submitted channel #%d with %d models",
		userId, channel.Id, len(req.Models)))

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"listing_id": listing.Id,
			"channel_id": channel.Id,
			"models":     req.Models,
			"status":     "active",
		},
	})
}

// GetMyListings returns all channel listings for the current user.
//
// GET /api/marketplace/channels
func GetMyListings(c *gin.Context) {
	userId := c.GetInt("id")
	listings, err := model.GetChannelListingsByUserId(userId)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	// Enrich with channel health info
	type ListingWithStats struct {
		*model.ChannelListing
		ChannelStatus    int    `json:"channel_status"`
		MonthlyConsumed  int64  `json:"monthly_consumed"`
		CircuitState     string `json:"circuit_state"`
	}

	result := make([]ListingWithStats, 0, len(listings))
	for _, l := range listings {
		item := ListingWithStats{
			ChannelListing: l,
		}

		// Get channel status
		if ch, err := model.CacheGetChannel(l.ChannelId); err == nil {
			item.ChannelStatus = ch.Status
		}

		// Get monthly consumption
		consumed, _ := model.GetListingMonthlyConsumed(l.Id)
		item.MonthlyConsumed = consumed

		// Get circuit breaker state
		state, _, _ := GetChannelCircuitState(l.ChannelId)
		switch state {
		case CircuitClosed:
			item.CircuitState = "healthy"
		case CircuitOpen:
			item.CircuitState = "down"
		case CircuitHalfOpen:
			item.CircuitState = "recovering"
		}

		result = append(result, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    result,
	})
}

// PauseOrResumeListing allows the seller to pause/resume their listing.
//
// PUT /api/marketplace/channels/:id/status
func PauseOrResumeListing(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	userId := c.GetInt("id")
	listing, err := model.GetChannelListingByIdAndUserId(id, userId)
	if err != nil {
		common.ApiErrorMsg(c, "Listing not found")
		return
	}

	var req struct {
		Status int `json:"status"` // 1=active, 0=paused
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	// Update listing status
	listing.Status = req.Status
	if err := listing.Update(); err != nil {
		common.ApiError(c, err)
		return
	}

	// Enable/disable the actual channel
	channelStatus := common.ChannelStatusEnabled
	if req.Status == model.ListingStatusPaused {
		channelStatus = common.ChannelStatusManuallyDisabled
	}
	model.UpdateChannelStatus(listing.ChannelId, "", channelStatus, "marketplace listing status change")
	model.InitChannelCache()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

// DeleteListing removes a listing and its associated channel.
//
// DELETE /api/marketplace/channels/:id
func DeleteListing(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	userId := c.GetInt("id")
	listing, err := model.GetChannelListingByIdAndUserId(id, userId)
	if err != nil {
		common.ApiErrorMsg(c, "Listing not found")
		return
	}

	// Delete the channel first
	if ch, err := model.GetChannelById(listing.ChannelId, false); err == nil {
		ch.Delete()
	}

	// Delete the listing
	listing.Delete()
	model.InitChannelCache()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
	})
}

// GetMyEarnings returns earning history for the current user.
//
// GET /api/marketplace/earnings?month=2026-04
func GetMyEarnings(c *gin.Context) {
	userId := c.GetInt("id")
	month := c.Query("month")

	earnings, err := model.GetUserEarnings(userId, month)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	unsettled, _ := model.GetUserTotalUnsettled(userId)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"earnings":          earnings,
			"total_unsettled":   unsettled,
			"unsettled_display": logger.FormatQuota(int(unsettled)),
		},
	})
}

// SettleEarnings converts unsettled earnings to user quota balance.
//
// POST /api/marketplace/earnings/settle
func SettleEarnings(c *gin.Context) {
	userId := c.GetInt("id")

	total, err := model.SettleUserEarnings(userId)
	if err != nil {
		common.ApiErrorMsg(c, err.Error())
		return
	}

	if total == 0 {
		common.ApiErrorMsg(c, "No unsettled earnings")
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"settled_amount":  total,
			"settled_display": logger.FormatQuota(int(total)),
		},
	})
}
