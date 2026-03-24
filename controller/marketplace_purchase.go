package controller

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/logger"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/system_setting"

	"github.com/gin-gonic/gin"
)

// MarketplacePurchaseRequest is the request body for purchasing quota and creating a token.
type MarketplacePurchaseRequest struct {
	Name   string   `json:"name" binding:"required"`
	Models []string `json:"models"` // optional: restrict to these models
	Expire string   `json:"expire"` // optional: "1h", "1d", "7d", "30d", "90d", "365d"
	Quota  int      `json:"quota" binding:"required"` // quota amount to purchase
}

// MarketplacePurchaseResponse is the response for a purchase request.
type MarketplacePurchaseResponse struct {
	// Immediate purchase (within budget)
	Token     string   `json:"token,omitempty"`
	Name      string   `json:"name"`
	Models    []string `json:"models,omitempty"`
	Endpoint  string   `json:"endpoint,omitempty"`
	ExpiresAt int64    `json:"expires_at,omitempty"`

	// Pending purchase (over budget)
	Status     string `json:"status"` // "completed" or "pending"
	PurchaseId int    `json:"purchase_id,omitempty"`
	Message    string `json:"message,omitempty"`
}

// PurchaseMarketplaceQuota handles the agent-initiated purchase flow.
// Within monthly budget: auto-approve, deduct quota, create token, send notification email.
// Over monthly budget: create pending purchase, send confirmation email.
//
// POST /api/marketplace/purchase
func PurchaseMarketplaceQuota(c *gin.Context) {
	var req MarketplacePurchaseRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	// --- Validate inputs (reuse existing marketplace validation logic) ---
	req.Name = strings.TrimSpace(req.Name)
	if len(req.Name) == 0 || len(req.Name) > 50 {
		common.ApiErrorI18n(c, i18n.MsgTokenNameTooLong)
		return
	}

	if req.Quota <= 0 {
		common.ApiErrorI18n(c, i18n.MsgMarketplaceInvalidQuota)
		return
	}

	for _, m := range req.Models {
		if len(m) == 0 || len(m) > 200 {
			common.ApiErrorI18n(c, i18n.MsgInvalidParams)
			return
		}
		for _, ch := range m {
			if !isModelNameChar(ch) {
				common.ApiErrorI18n(c, i18n.MsgMarketplaceInvalidModelName)
				return
			}
		}
	}
	if len(req.Models) > 50 {
		common.ApiErrorI18n(c, i18n.MsgMarketplaceModelListTooLong)
		return
	}

	if req.Expire != "" {
		if _, ok := parseExpireDuration(req.Expire); !ok {
			common.ApiErrorI18n(c, i18n.MsgMarketplaceInvalidExpire)
			return
		}
	}

	// Validate models exist in pricing
	if len(req.Models) > 0 {
		pricing := model.GetPricing()
		availableModels := make(map[string]bool, len(pricing))
		for _, p := range pricing {
			availableModels[p.ModelName] = true
		}
		for _, m := range req.Models {
			if !availableModels[m] {
				c.JSON(http.StatusOK, gin.H{
					"success": false,
					"message": fmt.Sprintf("Model not available: %s", m),
				})
				return
			}
		}
	}

	// --- Check user balance ---
	userId := c.GetInt("id")
	user, err := model.GetUserCache(userId)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	if user.Quota < req.Quota {
		common.ApiErrorI18n(c, i18n.MsgMarketplaceInsufficientBalance)
		return
	}

	// --- Check monthly budget ---
	userSetting := user.GetSetting()
	monthlyBudget := userSetting.MarketplaceMonthlyBudget // 0 means unlimited

	monthlySpending, err := model.GetUserMonthlyMarketplaceSpending(userId)
	if err != nil {
		common.SysLog("marketplace: failed to get monthly spending: " + err.Error())
		common.ApiErrorI18n(c, i18n.MsgMarketplacePurchaseFailed)
		return
	}

	withinBudget := monthlyBudget == 0 || (int(monthlySpending)+req.Quota <= monthlyBudget)
	sourceIP := c.ClientIP()

	if withinBudget {
		// --- Auto-approve: deduct + create token + notify ---
		resp, err := executeMarketplacePurchase(c, user, &req, sourceIP, nil)
		if err != nil {
			common.SysLog("marketplace: purchase failed: " + err.Error())
			common.ApiErrorI18n(c, i18n.MsgMarketplacePurchaseFailed)
			return
		}

		// Send notification email (async, non-blocking)
		go sendPurchaseNotificationEmail(user, &req, resp.Token, sourceIP, false)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    resp,
		})
	} else {
		// --- Over budget: create pending purchase + send confirmation email ---
		confirmToken := common.GenerateVerificationCode(0) // 32-char UUID

		purchase := &model.MarketplacePurchase{
			UserId:         userId,
			TokenName:      req.Name,
			Models:         strings.Join(req.Models, ","),
			Quota:          req.Quota,
			ExpireDuration: req.Expire,
			Status:         model.MarketplacePurchaseStatusPending,
			ConfirmToken:   confirmToken,
			SourceIP:       sourceIP,
		}

		if err := purchase.Insert(); err != nil {
			common.SysLog("marketplace: failed to create pending purchase: " + err.Error())
			common.ApiErrorI18n(c, i18n.MsgMarketplacePurchaseFailed)
			return
		}

		// Send confirmation email (async)
		go sendPurchaseConfirmationEmail(user, &req, purchase, sourceIP)

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": MarketplacePurchaseResponse{
				Name:       req.Name,
				Models:     req.Models,
				Status:     "pending",
				PurchaseId: purchase.Id,
				Message:    "Purchase exceeds monthly budget. A confirmation email has been sent to the account owner.",
			},
		})
	}
}

// ConfirmMarketplacePurchase handles the email confirmation link click.
//
// GET /api/marketplace/purchase/confirm?token=<confirm_token>
func ConfirmMarketplacePurchase(c *gin.Context) {
	confirmToken := strings.TrimSpace(c.Query("token"))
	if confirmToken == "" {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	purchase, err := model.GetMarketplacePurchaseByConfirmToken(confirmToken)
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgMarketplacePurchaseNotFound)
		return
	}

	if purchase.Status != model.MarketplacePurchaseStatusPending {
		common.ApiErrorI18n(c, i18n.MsgMarketplacePurchaseNotPending)
		return
	}

	// Check if the pending purchase has expired (24 hours)
	if time.Since(purchase.CreatedAt) > 24*time.Hour {
		_ = purchase.UpdateStatus(model.MarketplacePurchaseStatusExpired)
		common.ApiErrorI18n(c, i18n.MsgMarketplacePurchaseExpired)
		return
	}

	// Load user
	user, err := model.GetUserCache(purchase.UserId)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	// Re-check balance
	if user.Quota < purchase.Quota {
		common.ApiErrorI18n(c, i18n.MsgMarketplaceInsufficientBalance)
		return
	}

	// Execute the purchase
	models := []string{}
	if purchase.Models != "" {
		models = strings.Split(purchase.Models, ",")
	}

	req := &MarketplacePurchaseRequest{
		Name:   purchase.TokenName,
		Models: models,
		Expire: purchase.ExpireDuration,
		Quota:  purchase.Quota,
	}

	resp, err := executeMarketplacePurchase(c, user, req, purchase.SourceIP, purchase)
	if err != nil {
		common.SysLog("marketplace: confirmed purchase execution failed: " + err.Error())
		common.ApiErrorI18n(c, i18n.MsgMarketplacePurchaseFailed)
		return
	}

	// Send notification email about confirmed purchase
	go sendPurchaseNotificationEmail(user, req, resp.Token, purchase.SourceIP, true)

	// Redirect to a user-friendly page or return JSON
	serverAddress := getServerAddress(c)
	c.Redirect(http.StatusFound, serverAddress+"/token?purchase_confirmed=1")
}

// executeMarketplacePurchase performs the actual quota deduction and token creation.
// If existingPurchase is non-nil, it updates that record instead of creating a new one.
func executeMarketplacePurchase(c *gin.Context, user *model.UserBase, req *MarketplacePurchaseRequest, sourceIP string, existingPurchase *model.MarketplacePurchase) (*MarketplacePurchaseResponse, error) {
	userId := user.Id

	// Deduct quota from user
	err := model.DecreaseUserQuota(userId, req.Quota)
	if err != nil {
		return nil, fmt.Errorf("failed to deduct quota: %v", err)
	}

	// Parse expiry
	expiredTime := int64(-1)
	if req.Expire != "" {
		duration, _ := parseExpireDuration(req.Expire)
		expiredTime = common.GetTimestamp() + duration
	}

	// Generate token key
	key, err := common.GenerateKey()
	if err != nil {
		// Refund on failure
		_ = model.IncreaseUserQuota(userId, req.Quota, true)
		return nil, fmt.Errorf("failed to generate key: %v", err)
	}

	// Get user group
	group := ""
	autoGroups := service.GetUserAutoGroup(user.Group)
	if len(autoGroups) > 0 {
		group = autoGroups[0]
	}

	// Build token
	modelLimitsEnabled := len(req.Models) > 0
	modelLimits := ""
	if modelLimitsEnabled {
		modelLimits = strings.Join(req.Models, ",")
	}

	token := model.Token{
		UserId:             userId,
		Name:               req.Name,
		Key:                key,
		CreatedTime:        common.GetTimestamp(),
		AccessedTime:       common.GetTimestamp(),
		ExpiredTime:        expiredTime,
		RemainQuota:        req.Quota,
		UnlimitedQuota:     false,
		ModelLimitsEnabled: modelLimitsEnabled,
		ModelLimits:        modelLimits,
		Group:              group,
	}

	if err := token.Insert(); err != nil {
		// Refund on failure
		_ = model.IncreaseUserQuota(userId, req.Quota, true)
		return nil, fmt.Errorf("failed to create token: %v", err)
	}

	var purchaseId int
	if existingPurchase != nil {
		// Update the existing pending purchase record
		_ = existingPurchase.UpdateStatus(model.MarketplacePurchaseStatusCompleted)
		_ = existingPurchase.SetTokenInfo(token.Id, key)
		purchaseId = existingPurchase.Id
	} else {
		// Create a new completed purchase record (auto-approved)
		confirmToken := common.GenerateVerificationCode(0)
		purchase := &model.MarketplacePurchase{
			UserId:         userId,
			TokenName:      req.Name,
			Models:         modelLimits,
			Quota:          req.Quota,
			ExpireDuration: req.Expire,
			Status:         model.MarketplacePurchaseStatusCompleted,
			ConfirmToken:   confirmToken,
			TokenId:        token.Id,
			TokenKey:       key,
			SourceIP:       sourceIP,
		}
		now := time.Now()
		purchase.CompletedAt = &now
		if err := purchase.Insert(); err != nil {
			common.SysLog("marketplace: failed to record purchase: " + err.Error())
		}
		purchaseId = purchase.Id
	}

	// Log the purchase
	model.RecordLog(userId, model.LogTypeTopup,
		fmt.Sprintf("Marketplace 购买: %s, 额度 %s", req.Name, logger.LogQuota(req.Quota)))

	serverAddress := getServerAddress(c)
	expiresAt := int64(0)
	if expiredTime != -1 {
		expiresAt = expiredTime
	}

	return &MarketplacePurchaseResponse{
		Token:      "sk-" + key,
		Name:       req.Name,
		Models:     req.Models,
		Endpoint:   serverAddress + "/v1/chat/completions",
		ExpiresAt:  expiresAt,
		Status:     "completed",
		PurchaseId: purchaseId,
	}, nil
}

// getServerAddress returns the configured server address or constructs one from the request.
func getServerAddress(c *gin.Context) string {
	serverAddress := system_setting.ServerAddress
	if serverAddress == "" {
		proto := c.Request.Header.Get("X-Forwarded-Proto")
		if proto == "" {
			if c.Request.TLS != nil {
				proto = "https"
			} else {
				proto = "http"
			}
		}
		serverAddress = proto + "://" + c.Request.Host
	}
	return serverAddress
}

// sendPurchaseNotificationEmail sends a notification email about a completed purchase.
func sendPurchaseNotificationEmail(user *model.UserBase, req *MarketplacePurchaseRequest, tokenDisplay string, sourceIP string, wasConfirmed bool) {
	userSetting := user.GetSetting()
	email := userSetting.NotificationEmail
	if email == "" {
		email = user.Email
	}
	if email == "" {
		return
	}

	serverAddress := system_setting.ServerAddress
	if serverAddress == "" {
		serverAddress = "http://localhost:3000"
	}

	quotaDisplay := logger.FormatQuota(req.Quota)
	balanceDisplay := logger.FormatQuota(user.Quota - req.Quota) // after deduction

	modelsDisplay := "All models"
	if len(req.Models) > 0 {
		modelsDisplay = strings.Join(req.Models, ", ")
	}

	expireDisplay := "Never"
	if req.Expire != "" {
		expireDisplay = req.Expire
	}

	// Mask token for display
	tokenMasked := ""
	if len(tokenDisplay) > 10 {
		tokenMasked = tokenDisplay[:7] + "***"
	}

	triggerType := "Auto-approved (within monthly budget)"
	if wasConfirmed {
		triggerType = "Manually confirmed via email"
	}

	subject := fmt.Sprintf("[%s] Marketplace Purchase Notification", common.SystemName)

	content := fmt.Sprintf(`
<div style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
  <h2 style="color: #1a1a1a; border-bottom: 2px solid #4CAF50; padding-bottom: 10px;">Marketplace Purchase Notification</h2>

  <p>Hello <strong>%s</strong>,</p>
  <p>A quota purchase was completed via the Marketplace API:</p>

  <table style="width: 100%%; border-collapse: collapse; margin: 16px 0;">
    <tr><td style="padding: 8px; color: #666;">Purchase Amount</td><td style="padding: 8px; font-weight: bold;">%s</td></tr>
    <tr style="background: #f9f9f9;"><td style="padding: 8px; color: #666;">Token Name</td><td style="padding: 8px;">%s</td></tr>
    <tr><td style="padding: 8px; color: #666;">Models</td><td style="padding: 8px;">%s</td></tr>
    <tr style="background: #f9f9f9;"><td style="padding: 8px; color: #666;">Expiry</td><td style="padding: 8px;">%s</td></tr>
    <tr><td style="padding: 8px; color: #666;">Token</td><td style="padding: 8px; font-family: monospace;">%s</td></tr>
    <tr style="background: #f9f9f9;"><td style="padding: 8px; color: #666;">Source IP</td><td style="padding: 8px;">%s</td></tr>
    <tr><td style="padding: 8px; color: #666;">Trigger</td><td style="padding: 8px;">%s</td></tr>
    <tr style="background: #f9f9f9;"><td style="padding: 8px; color: #666;">Account Balance</td><td style="padding: 8px;">%s</td></tr>
  </table>

  <p style="color: #666; font-size: 14px;">If this was not authorized by you, please disable the token immediately:</p>
  <p><a href="%s/token" style="display: inline-block; padding: 10px 20px; background: #f44336; color: white; text-decoration: none; border-radius: 4px;">Manage Tokens</a></p>

  <hr style="border: none; border-top: 1px solid #eee; margin: 20px 0;">
  <p style="color: #999; font-size: 12px;">This email was sent by %s Marketplace API.</p>
</div>`,
		user.Username,
		quotaDisplay,
		req.Name,
		modelsDisplay,
		expireDisplay,
		tokenMasked,
		sourceIP,
		triggerType,
		balanceDisplay,
		serverAddress,
		common.SystemName,
	)

	if err := common.SendEmail(subject, email, content); err != nil {
		common.SysLog(fmt.Sprintf("marketplace: failed to send notification email to %s: %s", email, err.Error()))
	}
}

// sendPurchaseConfirmationEmail sends an email with a confirmation link for over-budget purchases.
func sendPurchaseConfirmationEmail(user *model.UserBase, req *MarketplacePurchaseRequest, purchase *model.MarketplacePurchase, sourceIP string) {
	userSetting := user.GetSetting()
	email := userSetting.NotificationEmail
	if email == "" {
		email = user.Email
	}
	if email == "" {
		common.SysLog(fmt.Sprintf("marketplace: user %d has no email, cannot send confirmation", user.Id))
		return
	}

	serverAddress := system_setting.ServerAddress
	if serverAddress == "" {
		serverAddress = "http://localhost:3000"
	}

	quotaDisplay := logger.FormatQuota(req.Quota)
	balanceDisplay := logger.FormatQuota(user.Quota)

	// Get monthly spending for context
	monthlySpending, _ := model.GetUserMonthlyMarketplaceSpending(user.Id)
	monthlySpendingDisplay := logger.FormatQuota(int(monthlySpending))

	budgetDisplay := "Unlimited"
	userSettingBudget := user.GetSetting().MarketplaceMonthlyBudget
	if userSettingBudget > 0 {
		budgetDisplay = logger.FormatQuota(userSettingBudget)
	}

	modelsDisplay := "All models"
	if len(req.Models) > 0 {
		modelsDisplay = strings.Join(req.Models, ", ")
	}

	confirmURL := fmt.Sprintf("%s/api/marketplace/purchase/confirm?token=%s", serverAddress, purchase.ConfirmToken)

	subject := fmt.Sprintf("[%s] Marketplace Purchase Confirmation Required", common.SystemName)

	content := fmt.Sprintf(`
<div style="font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif; max-width: 600px; margin: 0 auto; padding: 20px;">
  <h2 style="color: #1a1a1a; border-bottom: 2px solid #FF9800; padding-bottom: 10px;">Purchase Confirmation Required</h2>

  <p>Hello <strong>%s</strong>,</p>
  <p>Your AI agent attempted a Marketplace purchase that <strong>exceeds your monthly budget</strong>. Please review and confirm:</p>

  <table style="width: 100%%; border-collapse: collapse; margin: 16px 0;">
    <tr><td style="padding: 8px; color: #666;">Purchase Amount</td><td style="padding: 8px; font-weight: bold; color: #e65100;">%s</td></tr>
    <tr style="background: #f9f9f9;"><td style="padding: 8px; color: #666;">Token Name</td><td style="padding: 8px;">%s</td></tr>
    <tr><td style="padding: 8px; color: #666;">Models</td><td style="padding: 8px;">%s</td></tr>
    <tr style="background: #f9f9f9;"><td style="padding: 8px; color: #666;">Source IP</td><td style="padding: 8px;">%s</td></tr>
    <tr><td style="padding: 8px; color: #666;">Monthly Budget</td><td style="padding: 8px;">%s</td></tr>
    <tr style="background: #f9f9f9;"><td style="padding: 8px; color: #666;">Already Spent This Month</td><td style="padding: 8px;">%s</td></tr>
    <tr><td style="padding: 8px; color: #666;">Account Balance</td><td style="padding: 8px;">%s</td></tr>
  </table>

  <div style="text-align: center; margin: 24px 0;">
    <a href="%s" style="display: inline-block; padding: 14px 28px; background: #4CAF50; color: white; text-decoration: none; border-radius: 6px; font-size: 16px; font-weight: bold;">Confirm Purchase</a>
  </div>

  <p style="color: #999; font-size: 13px;">This confirmation link will expire in <strong>24 hours</strong>. If you did not authorize this purchase, you can safely ignore this email.</p>

  <p style="color: #999; font-size: 13px;">To increase your monthly budget and allow future auto-approvals, go to <a href="%s/setting">Account Settings</a>.</p>

  <hr style="border: none; border-top: 1px solid #eee; margin: 20px 0;">
  <p style="color: #999; font-size: 12px;">This email was sent by %s Marketplace API.</p>
</div>`,
		user.Username,
		quotaDisplay,
		req.Name,
		modelsDisplay,
		sourceIP,
		budgetDisplay,
		monthlySpendingDisplay,
		balanceDisplay,
		confirmURL,
		serverAddress,
		common.SystemName,
	)

	if err := common.SendEmail(subject, email, content); err != nil {
		common.SysLog(fmt.Sprintf("marketplace: failed to send confirmation email to %s: %s", email, err.Error()))
	}
}
