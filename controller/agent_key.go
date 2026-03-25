package controller

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/i18n"
	"github.com/QuantumNous/new-api/model"
	"github.com/QuantumNous/new-api/setting/system_setting"

	"github.com/gin-gonic/gin"
)

// CreateAgentKeyRequest is the request body for creating an agent key.
type CreateAgentKeyRequest struct {
	Name               string   `json:"name" binding:"required"` // e.g. "Lobster Bot"
	Models             []string `json:"models"`                  // Allowed models (empty = all)
	MonthlyBudget      int      `json:"monthly_budget"`          // Monthly auto-approve budget (quota), 0 = unlimited
	MaxQuotaPerRequest int      `json:"max_quota_per_request"`   // Max per provision call, 0 = no limit
	DefaultExpire      string   `json:"default_expire"`          // Default token expiry, e.g. "30d"
	AllowIps           string   `json:"allow_ips"`               // IP whitelist (newline separated)
}

// CreateAgentKey creates a new agent key for the current user.
//
// POST /api/agent/keys
func CreateAgentKey(c *gin.Context) {
	var req CreateAgentKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if len(req.Name) == 0 || len(req.Name) > 50 {
		common.ApiErrorI18n(c, i18n.MsgTokenNameTooLong)
		return
	}

	// Validate default expire
	if req.DefaultExpire != "" {
		if _, ok := parseExpireDuration(req.DefaultExpire); !ok {
			req.DefaultExpire = "30d"
		}
	} else {
		req.DefaultExpire = "30d"
	}

	userId := c.GetInt("id")

	// Generate key
	key, err := common.GenerateKey()
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgTokenGenerateFailed)
		return
	}

	modelsStr := ""
	if len(req.Models) > 0 {
		modelsStr = strings.Join(req.Models, ",")
	}

	var allowIps *string
	if req.AllowIps != "" {
		allowIps = &req.AllowIps
	}

	agentKey := &model.AgentKey{
		UserId:             userId,
		Name:               req.Name,
		Key:                key,
		Status:             1,
		Models:             modelsStr,
		MonthlyBudget:      req.MonthlyBudget,
		MaxQuotaPerRequest: req.MaxQuotaPerRequest,
		DefaultExpire:      req.DefaultExpire,
		AllowIps:           allowIps,
	}

	if err := agentKey.Insert(); err != nil {
		common.ApiError(c, err)
		return
	}

	serverAddress := system_setting.ServerAddress
	if serverAddress == "" {
		proto := c.Request.Header.Get("X-Forwarded-Proto")
		if proto == "" {
			proto = "http"
		}
		serverAddress = proto + "://" + c.Request.Host
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"id":   agentKey.Id,
			"key":  "ak-" + key,
			"name": agentKey.Name,
			"provision_endpoint": serverAddress + "/api/agent/provision",
			"usage_example": fmt.Sprintf(`curl -X POST %s/api/agent/provision -H "Authorization: Bearer ak-%s" -H "Content-Type: application/json" -d '{}'`,
				serverAddress, key[:8]+"***"),
		},
	})
}

// GetAgentKeys lists all agent keys for the current user.
//
// GET /api/agent/keys
func GetAgentKeys(c *gin.Context) {
	userId := c.GetInt("id")
	keys, err := model.GetAgentKeysByUserId(userId)
	if err != nil {
		common.ApiError(c, err)
		return
	}

	// Mask keys for display
	for _, k := range keys {
		if len(k.Key) > 8 {
			k.Key = k.Key[:4] + "****" + k.Key[len(k.Key)-4:]
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    keys,
	})
}

// UpdateAgentKey updates an existing agent key.
//
// PUT /api/agent/keys/:id
func UpdateAgentKey(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	userId := c.GetInt("id")
	agentKey, err := model.GetAgentKeyByIdAndUserId(id, userId)
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgAgentKeyNotFound)
		return
	}

	var req CreateAgentKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	if len(req.Name) > 0 {
		agentKey.Name = req.Name
	}
	if req.DefaultExpire != "" {
		if _, ok := parseExpireDuration(req.DefaultExpire); ok {
			agentKey.DefaultExpire = req.DefaultExpire
		}
	}

	if len(req.Models) > 0 {
		agentKey.Models = strings.Join(req.Models, ",")
	} else {
		agentKey.Models = ""
	}

	agentKey.MonthlyBudget = req.MonthlyBudget
	agentKey.MaxQuotaPerRequest = req.MaxQuotaPerRequest

	if req.AllowIps != "" {
		agentKey.AllowIps = &req.AllowIps
	}

	if err := agentKey.Update(); err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}

// DeleteAgentKey deletes an agent key.
//
// DELETE /api/agent/keys/:id
func DeleteAgentKey(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgInvalidParams)
		return
	}

	userId := c.GetInt("id")
	agentKey, err := model.GetAgentKeyByIdAndUserId(id, userId)
	if err != nil {
		common.ApiErrorI18n(c, i18n.MsgAgentKeyNotFound)
		return
	}

	if err := agentKey.Delete(); err != nil {
		common.ApiError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
	})
}
