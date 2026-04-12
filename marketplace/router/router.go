package router

import (
	mpController "github.com/QuantumNous/new-api/marketplace/controller"
	"github.com/QuantumNous/new-api/middleware"
	"github.com/gin-gonic/gin"
)

// SetMarketplaceRouter registers marketplace API routes
func SetMarketplaceRouter(router *gin.Engine) {
	mp := router.Group("/api/marketplace")
	mp.Use(middleware.GlobalAPIRateLimit())
	{
		// Public routes (no auth needed)
		mp.GET("/info", mpController.GetMarketInfo)

		// Authenticated routes
		authed := mp.Group("")
		authed.Use(middleware.UserAuth())
		{
			// Seller routes
			authed.POST("/seller/register", mpController.RegisterSeller)
			authed.GET("/seller/dashboard", mpController.GetSellerDashboard)

			// Listing management
			authed.POST("/listings", mpController.CreateListing)
			authed.POST("/listings/:id/pause", mpController.PauseListing)
			authed.POST("/listings/:id/resume", mpController.ResumeListing)
			authed.DELETE("/listings/:id", mpController.DeleteListing)
		}
	}
}
