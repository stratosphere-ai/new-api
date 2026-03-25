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

// AgentProvisionRequest is the minimal request an agent needs to send.
// All fields are optional — the agent key's pre-configured defaults are used as fallback.
type AgentProvisionRequest struct {
	// Optional overrides (if omitted, uses agent key defaults)
	Models []string `json:"models,omitempty"` // Override model restriction
	Quota  int      `json:"quota,omitempty"`  // Override quota amount (0 = use sensible default)
	Expire string   `json:"expire,omitempty"` // Override expiry
	Name   string   `json:"name,omitempty"`   // Token name (auto-generated if empty)
	Reason string   `json:"reason,omitempty"` // Why the agent needs this (logged for user visibility)
}

// AgentProvisionResponse is what the agent gets back.
type AgentProvisionResponse struct {
	Token     string   `json:"token"`               // Ready-to-use sk-xxx
	Endpoint  string   `json:"endpoint"`             // API base URL
	Models    []string `json:"models,omitempty"`     // Which models this token can access
	ExpiresAt int64    `json:"expires_at,omitempty"` // Unix timestamp, 0 = never
	Quota     int      `json:"quota"`                // How much quota was allocated
	Status    string   `json:"status"`               // "ready" or "pending_confirmation"
	Message   string   `json:"message,omitempty"`    // Human-readable status for agent to relay to user
}

const defaultProvisionQuota = 500000 // ~$1 USD, safe default

// AgentProvision is the one-stop endpoint for AI agents.
// The agent authenticates with an Agent Key (ak-xxx) and gets back a ready-to-use API token.
//
// POST /api/agent/provision
// Authorization: Bearer ak-xxxxxxxx
//
// The platform handles everything: auth, model selection, quota allocation, token creation, notification.
func AgentProvision(c *gin.Context) {
	// --- Step 1: Authenticate via Agent Key ---
	authHeader := c.Request.Header.Get("Authorization")
	rawKey := strings.TrimPrefix(authHeader, "Bearer ")
	rawKey = strings.TrimPrefix(rawKey, "ak-")

	if rawKey == "" || rawKey == authHeader {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Missing or invalid Agent Key. Expected: Authorization: Bearer ak-xxx",
		})
		return
	}

	agentKey, err := model.GetAgentKeyByKey(rawKey)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"message": "Invalid or disabled Agent Key.",
		})
		return
	}

	// IP whitelist check
	clientIP := c.ClientIP()
	if whitelist := agentKey.GetIpWhitelist(); len(whitelist) > 0 {
		allowed := false
		for _, ip := range whitelist {
			if ip == clientIP {
				allowed = true
				break
			}
		}
		if !allowed {
			c.JSON(http.StatusForbidden, gin.H{
				"success": false,
				"message": "This IP is not allowed for this Agent Key.",
			})
			return
		}
	}

	// --- Step 2: Parse request (merge with agent key defaults) ---
	var req AgentProvisionRequest
	// Allow empty body (all defaults)
	_ = c.ShouldBindJSON(&req)

	// Apply defaults from agent key config
	models := req.Models
	if len(models) == 0 {
		models = agentKey.GetAllowedModels()
	}

	quota := req.Quota
	if quota <= 0 {
		quota = defaultProvisionQuota
	}

	// Enforce per-request limit from agent key
	if agentKey.MaxQuotaPerRequest > 0 && quota > agentKey.MaxQuotaPerRequest {
		quota = agentKey.MaxQuotaPerRequest
	}

	expire := req.Expire
	if expire == "" {
		expire = agentKey.DefaultExpire
	}
	if expire != "" {
		if _, ok := parseExpireDuration(expire); !ok {
			expire = "30d" // fallback to safe default
		}
	}

	tokenName := req.Name
	if tokenName == "" {
		tokenName = fmt.Sprintf("%s-%s", agentKey.Name, time.Now().Format("20060102-150405"))
	}
	if len(tokenName) > 50 {
		tokenName = tokenName[:50]
	}

	// Validate models against pricing
	if len(models) > 0 {
		pricing := model.GetPricing()
		availableModels := make(map[string]bool, len(pricing))
		for _, p := range pricing {
			availableModels[p.ModelName] = true
		}
		validModels := make([]string, 0, len(models))
		for _, m := range models {
			if availableModels[m] {
				validModels = append(validModels, m)
			}
		}
		models = validModels
	}

	// --- Step 3: Check user balance ---
	user, err := model.GetUserCache(agentKey.UserId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to load account information.",
		})
		return
	}

	if user.Quota < quota {
		c.JSON(http.StatusPaymentRequired, gin.H{
			"success": false,
			"message": fmt.Sprintf("Insufficient account balance. Available: %s, Requested: %s",
				logger.FormatQuota(user.Quota), logger.FormatQuota(quota)),
		})
		return
	}

	// --- Step 4: Budget check ---
	monthlySpending, err := model.GetAgentKeyMonthlySpending(agentKey.Id)
	if err != nil {
		common.SysLog("agent: failed to get monthly spending: " + err.Error())
		monthlySpending = 0
	}

	withinBudget := agentKey.MonthlyBudget == 0 || (int(monthlySpending)+quota <= agentKey.MonthlyBudget)

	if !withinBudget {
		// Over budget → create pending purchase, send confirmation email
		confirmToken := common.GenerateVerificationCode(0)
		purchase := &model.MarketplacePurchase{
			UserId:         agentKey.UserId,
			TokenName:      tokenName,
			Models:         strings.Join(models, ","),
			Quota:          quota,
			ExpireDuration: expire,
			Status:         model.MarketplacePurchaseStatusPending,
			ConfirmToken:   confirmToken,
			AgentKeyId:     agentKey.Id,
			SourceIP:       clientIP,
		}
		if err := purchase.Insert(); err != nil {
			common.SysLog("agent: failed to create pending purchase: " + err.Error())
			common.ApiErrorI18n(c, i18n.MsgMarketplacePurchaseFailed)
			return
		}

		// Send confirmation email
		purchaseReq := &MarketplacePurchaseRequest{
			Name:   tokenName,
			Models: models,
			Quota:  quota,
			Expire: expire,
		}
		go sendPurchaseConfirmationEmail(user, purchaseReq, purchase, clientIP)

		// Update last used
		go agentKey.TouchLastUsed()

		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": AgentProvisionResponse{
				Status: "pending_confirmation",
				Quota:  quota,
				Models: models,
				Message: fmt.Sprintf(
					"Purchase of %s exceeds monthly budget (%s spent of %s). Confirmation email sent to account owner. Retry in a few minutes.",
					logger.FormatQuota(quota),
					logger.FormatQuota(int(monthlySpending)),
					logger.FormatQuota(agentKey.MonthlyBudget),
				),
			},
		})
		return
	}

	// --- Step 5: Within budget → Execute immediately ---
	serverAddress := getServerAddress(c)
	resp, err := provisionAgentToken(agentKey, user, tokenName, models, quota, expire, clientIP, serverAddress)
	if err != nil {
		common.SysLog("agent: provision failed: " + err.Error())
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"message": "Failed to provision token. Please retry.",
		})
		return
	}

	// Notify owner (async)
	purchaseReq := &MarketplacePurchaseRequest{
		Name:   tokenName,
		Models: models,
		Quota:  quota,
		Expire: expire,
	}
	go sendPurchaseNotificationEmail(user, purchaseReq, resp.Token, clientIP, false)
	go agentKey.TouchLastUsed()

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    resp,
	})
}

// provisionAgentToken handles the actual token creation for the agent gateway.
func provisionAgentToken(agentKey *model.AgentKey, user *model.UserBase, tokenName string, models []string, quota int, expire string, clientIP string, serverAddress string) (*AgentProvisionResponse, error) {
	userId := agentKey.UserId

	// Deduct quota
	if err := model.DecreaseUserQuota(userId, quota); err != nil {
		return nil, fmt.Errorf("failed to deduct quota: %v", err)
	}

	// Parse expiry
	expiredTime := int64(-1)
	if expire != "" {
		duration, _ := parseExpireDuration(expire)
		expiredTime = common.GetTimestamp() + duration
	}

	// Generate key
	key, err := common.GenerateKey()
	if err != nil {
		_ = model.IncreaseUserQuota(userId, quota, true)
		return nil, fmt.Errorf("failed to generate key: %v", err)
	}

	// User group
	group := ""
	autoGroups := service.GetUserAutoGroup(user.Group)
	if len(autoGroups) > 0 {
		group = autoGroups[0]
	}

	// Build token
	modelLimitsEnabled := len(models) > 0
	modelLimits := ""
	if modelLimitsEnabled {
		modelLimits = strings.Join(models, ",")
	}

	token := model.Token{
		UserId:             userId,
		Name:               tokenName,
		Key:                key,
		CreatedTime:        common.GetTimestamp(),
		AccessedTime:       common.GetTimestamp(),
		ExpiredTime:        expiredTime,
		RemainQuota:        quota,
		UnlimitedQuota:     false,
		ModelLimitsEnabled: modelLimitsEnabled,
		ModelLimits:        modelLimits,
		Group:              group,
	}

	if err := token.Insert(); err != nil {
		_ = model.IncreaseUserQuota(userId, quota, true)
		return nil, fmt.Errorf("failed to create token: %v", err)
	}

	// Record purchase
	confirmToken := common.GenerateVerificationCode(0)
	now := time.Now()
	purchase := &model.MarketplacePurchase{
		UserId:         userId,
		TokenName:      tokenName,
		Models:         modelLimits,
		Quota:          quota,
		ExpireDuration: expire,
		Status:         model.MarketplacePurchaseStatusCompleted,
		ConfirmToken:   confirmToken,
		TokenId:        token.Id,
		TokenKey:       key,
		AgentKeyId:     agentKey.Id,
		SourceIP:       clientIP,
		CompletedAt:    &now,
	}
	if err := purchase.Insert(); err != nil {
		common.SysLog("agent: failed to record purchase: " + err.Error())
	}

	// Log
	model.RecordLog(userId, model.LogTypeTopup,
		fmt.Sprintf("Agent [%s] 自动购买: %s, 额度 %s", agentKey.Name, tokenName, logger.LogQuota(quota)))

	expiresAt := int64(0)
	if expiredTime != -1 {
		expiresAt = expiredTime
	}

	return &AgentProvisionResponse{
		Token:     "sk-" + key,
		Endpoint:  serverAddress + "/v1/chat/completions",
		Models:    models,
		ExpiresAt: expiresAt,
		Quota:     quota,
		Status:    "ready",
		Message:   "Token provisioned successfully. Ready to use.",
	}, nil
}
