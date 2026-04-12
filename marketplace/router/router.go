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
			authed.POST("/seller/register", middleware.CriticalRateLimit(), mpController.RegisterSeller)
			authed.GET("/seller/dashboard", mpController.GetSellerDashboard)
			authed.GET("/seller/withdrawals", mpController.GetSellerWithdrawals)
			authed.GET("/seller/trades", mpController.GetSellerTrades)
			authed.POST("/seller/withdrawal/request", middleware.CriticalRateLimit(), mpController.RequestWithdrawal)

			// Listing management
			authed.POST("/listings", middleware.CriticalRateLimit(), mpController.CreateListing)
			authed.POST("/listings/:id/pause", mpController.PauseListing)
			authed.POST("/listings/:id/resume", mpController.ResumeListing)
			authed.DELETE("/listings/:id", mpController.DeleteListing)

			// Buyer routes
			authed.GET("/buyer/stats", mpController.GetBuyerStats)
			authed.GET("/buyer/trades", mpController.GetBuyerTrades)
			authed.GET("/buyer/models", mpController.GetBuyerModelUsage)
			authed.GET("/buyer/api-info", mpController.GetBuyerAPIInfo)
			authed.POST("/buyer/api-token", middleware.CriticalRateLimit(), mpController.CreateBuyerAPIToken)
		}

		// Admin routes (require admin role)
		admin := mp.Group("/admin")
		admin.Use(middleware.AdminAuth())
		{
			// Platform overview
			admin.GET("/stats", mpController.AdminGetPlatformStats)

			// Seller management
			admin.GET("/sellers", mpController.AdminGetSellers)
			admin.GET("/sellers/:id", mpController.AdminGetSellerDetail)
			admin.POST("/sellers/:id/suspend", mpController.AdminSuspendSeller)
			admin.POST("/sellers/:id/enable", mpController.AdminEnableSeller)

			// Withdrawal management
			admin.GET("/withdrawals", mpController.AdminGetWithdrawals)
			admin.POST("/withdrawals/:id/approve", mpController.AdminApproveWithdrawal)
			admin.POST("/withdrawals/:id/paid", mpController.AdminMarkWithdrawalPaid)
			admin.POST("/withdrawals/:id/reject", mpController.AdminRejectWithdrawal)

			// Settlement
			admin.POST("/settlement/trigger", mpController.AdminTriggerSettlement)
		}
	}
}
