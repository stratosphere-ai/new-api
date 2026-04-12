package controller

import (
	"github.com/QuantumNous/new-api/common"
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
