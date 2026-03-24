package controller

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/service"
	"github.com/QuantumNous/new-api/setting/operation_setting"
	"github.com/QuantumNous/new-api/setting/system_setting"

	"github.com/gin-gonic/gin"
)

// MarketplaceCatalogItem is the public-facing model info for AI agents.
type MarketplaceCatalogItem struct {
	ModelName       string   `json:"model_name"`
	Description     string   `json:"description,omitempty"`
	QuotaType       int      `json:"quota_type"` // 0=token-based, 1=fixed-price
	InputPrice      float64  `json:"input_price"`
	OutputPrice     float64  `json:"output_price"`
	CompletionRatio float64  `json:"completion_ratio,omitempty"`
	Endpoints       []string `json:"endpoints"`
	Tags            string   `json:"tags,omitempty"`
	Vendor          string   `json:"vendor,omitempty"`
	VendorIcon      string   `json:"vendor_icon,omitempty"`
}

// GetMarketplaceCatalog returns the list of available models with pricing info.
// Public endpoint, no auth required. Suitable for AI agent discovery.
//
// GET /api/marketplace/models
// Query params:
//   - capability: filter by endpoint type (e.g. "chat_completions")
//   - vendor: filter by vendor name
//   - q: search model name
func GetMarketplaceCatalog(c *gin.Context) {
	pricing := model.GetPricing()
	vendors := model.GetVendors()
	endpointMap := model.GetSupportedEndpointMap()

	// Build vendor lookup
	vendorLookup := make(map[int]model.PricingVendor)
	for _, v := range vendors {
		vendorLookup[v.ID] = v
	}

	// Parse filters
	capabilityFilter := strings.TrimSpace(c.Query("capability"))
	vendorFilter := strings.ToLower(strings.TrimSpace(c.Query("vendor")))
	queryFilter := strings.ToLower(strings.TrimSpace(c.Query("q")))

	items := make([]MarketplaceCatalogItem, 0, len(pricing))
	for _, p := range pricing {
		// Build endpoint list
		endpoints := make([]string, 0, len(p.SupportedEndpointTypes))
		for _, et := range p.SupportedEndpointTypes {
			if info, ok := endpointMap[string(et)]; ok {
				endpoints = append(endpoints, info.Path)
			} else {
				endpoints = append(endpoints, string(et))
			}
		}

		// Apply capability filter
		if capabilityFilter != "" {
			matched := false
			for _, et := range p.SupportedEndpointTypes {
				if string(et) == capabilityFilter {
					matched = true
					break
				}
			}
			if !matched {
				continue
			}
		}

		// Apply vendor filter
		vendorName := ""
		vendorIcon := ""
		if v, ok := vendorLookup[p.VendorID]; ok {
			vendorName = v.Name
			vendorIcon = v.Icon
		}
		if vendorFilter != "" && !strings.Contains(strings.ToLower(vendorName), vendorFilter) {
			continue
		}

		// Apply query filter
		if queryFilter != "" && !strings.Contains(strings.ToLower(p.ModelName), queryFilter) {
			continue
		}

		item := MarketplaceCatalogItem{
			ModelName:   p.ModelName,
			Description: p.Description,
			QuotaType:   p.QuotaType,
			Endpoints:   endpoints,
			Tags:        p.Tags,
			Vendor:      vendorName,
			VendorIcon:  vendorIcon,
		}

		// Calculate prices (same logic as pricing page)
		if p.QuotaType == 1 {
			// Fixed price
			item.InputPrice = p.ModelPrice
			item.OutputPrice = p.ModelPrice
		} else {
			// Token-based: model_ratio is the input multiplier
			item.InputPrice = p.ModelRatio
			item.OutputPrice = p.ModelRatio * p.CompletionRatio
			item.CompletionRatio = p.CompletionRatio
		}

		items = append(items, item)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    items,
	})
}

// MarketplaceCreateTokenRequest is the request body for one-step token creation.
type MarketplaceCreateTokenRequest struct {
	Name   string   `json:"name" binding:"required"`
	Models []string `json:"models"` // optional: restrict to these models
	Expire string   `json:"expire"` // optional: "1h", "1d", "7d", "30d", "" = never
}

// MarketplaceCreateTokenResponse is the response for one-step token creation.
type MarketplaceCreateTokenResponse struct {
	Token     string   `json:"token"`
	Name      string   `json:"name"`
	Models    []string `json:"models,omitempty"`
	Endpoint  string   `json:"endpoint"`
	ExpiresAt int64    `json:"expires_at"` // 0 = never
}

// CreateMarketplaceToken creates a token with minimal input.
// Requires user authentication + Turnstile (if enabled) + rate limiting.
//
// POST /api/marketplace/tokens
func CreateMarketplaceToken(c *gin.Context) {
	var req MarketplaceCreateTokenRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	// Validate name
	req.Name = strings.TrimSpace(req.Name)
	if len(req.Name) == 0 || len(req.Name) > 50 {
		common.ApiErrorI18n(c, i18n.MsgTokenNameTooLong)
		return
	}

	// Validate model names: only allow alphanumeric, dash, dot, slash, colon, underscore
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

	// Limit model list size
	if len(req.Models) > 50 {
		common.ApiErrorI18n(c, i18n.MsgMarketplaceModelListTooLong)
		return
	}

	// Check max tokens per user
	userId := c.GetInt("id")
	maxTokens := operation_setting.GetMaxUserTokens()
	count, err := model.CountUserTokens(userId)
	if err != nil {
		common.ApiError(c, err)
		return
	}
	if int(count) >= maxTokens {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": fmt.Sprintf("Maximum token limit reached (%d)", maxTokens),
		})
		return
	}

	// Parse expiry
	expiredTime := int64(-1) // never
	if req.Expire != "" {
		duration, ok := parseExpireDuration(req.Expire)
		if !ok {
			common.ApiErrorI18n(c, i18n.MsgMarketplaceInvalidExpire)
			return
		}
		expiredTime = common.GetTimestamp() + duration
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

	// Generate token key
	key, err := common.GenerateKey()
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgTokenGenerateFailed)
		common.SysLog("marketplace: failed to generate token key: " + err.Error())
		return
	}

	// Get user's group for the token
	group := ""
	user, err := model.GetUserCache(userId)
	if err == nil {
		autoGroups := service.GetUserAutoGroup(user.Group)
		if len(autoGroups) > 0 {
			group = autoGroups[0]
		}
	}

	// Build token
	modelLimitsEnabled := len(req.Models) > 0
	modelLimits := ""
	if modelLimitsEnabled {
		modelLimits = strings.Join(req.Models, ",")
	}

	cleanToken := model.Token{
		UserId:             userId,
		Name:               req.Name,
		Key:                key,
		CreatedTime:        common.GetTimestamp(),
		AccessedTime:       common.GetTimestamp(),
		ExpiredTime:        expiredTime,
		RemainQuota:        0,
		UnlimitedQuota:     true,
		ModelLimitsEnabled: modelLimitsEnabled,
		ModelLimits:        modelLimits,
		Group:              group,
	}

	if err := cleanToken.Insert(); err != nil {
		common.ApiError(c, err)
		return
	}

	// Build server address for endpoint
	serverAddress := system_setting.ServerAddress
	if serverAddress == "" {
		serverAddress = c.Request.Host
		if c.Request.TLS != nil {
			serverAddress = "https://" + serverAddress
		} else {
			serverAddress = c.Request.Header.Get("X-Forwarded-Proto")
			if serverAddress == "" {
				serverAddress = "http"
			}
			serverAddress = serverAddress + "://" + c.Request.Host
		}
	}

	expiresAt := int64(0)
	if expiredTime != -1 {
		expiresAt = expiredTime
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": MarketplaceCreateTokenResponse{
			Token:     "sk-" + key,
			Name:      req.Name,
			Models:    req.Models,
			Endpoint:  serverAddress + "/v1/chat/completions",
			ExpiresAt: expiresAt,
		},
	})
}

// GetMarketplaceServiceDesc returns a machine-readable service description.
// Designed for AI agents to discover available API capabilities.
//
// GET /.well-known/ai-marketplace.json
func GetMarketplaceServiceDesc(c *gin.Context) {
	serverAddress := system_setting.ServerAddress
	if serverAddress == "" {
		proto := c.Request.Header.Get("X-Forwarded-Proto")
		if proto == "" {
			proto = "http"
		}
		serverAddress = proto + "://" + c.Request.Host
	}

	c.JSON(http.StatusOK, gin.H{
		"name":    common.SystemName,
		"version": "1.0",
		"capabilities": []string{
			"chat_completions",
			"completions",
			"embeddings",
			"images",
			"audio",
			"moderation",
		},
		"endpoints": gin.H{
			"catalog":  serverAddress + "/api/marketplace/models",
			"purchase": serverAddress + "/api/marketplace/tokens",
		},
		"authentication":  "bearer",
		"pricing_unit":    "quota",
		"requires_signup": true,
		"rate_limits": gin.H{
			"catalog_per_minute":  60,
			"purchase_per_hour":   10,
			"requires_turnstile":  common.TurnstileCheckEnabled,
		},
	})
}

// isModelNameChar checks if a character is valid in a model name.
func isModelNameChar(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') ||
		ch == '-' || ch == '_' || ch == '.' || ch == '/' || ch == ':'
}

// parseExpireDuration parses a human-readable duration string to seconds.
// Supported formats: "1h", "1d", "7d", "30d", "90d", "365d"
func parseExpireDuration(s string) (int64, bool) {
	s = strings.TrimSpace(strings.ToLower(s))
	switch s {
	case "1h":
		return 3600, true
	case "1d":
		return 86400, true
	case "7d":
		return 7 * 86400, true
	case "30d":
		return 30 * 86400, true
	case "90d":
		return 90 * 86400, true
	case "365d":
		return 365 * 86400, true
	default:
		return 0, false
	}
}
