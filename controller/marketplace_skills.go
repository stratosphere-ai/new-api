package controller

import (
	"net/http"

	"github.com/QuantumNous/new-api/common"
	"github.com/QuantumNous/new-api/setting/system_setting"

	"github.com/gin-gonic/gin"
)

// GetMarketplaceAgentSkills returns a structured instruction set for AI agents.
// This is the "operation manual" that tells an agent how to autonomously
// browse models, purchase quota, and create tokens.
//
// GET /.well-known/ai-marketplace-skills.json
func GetMarketplaceAgentSkills(c *gin.Context) {
	serverAddress := system_setting.ServerAddress
	if serverAddress == "" {
		proto := c.Request.Header.Get("X-Forwarded-Proto")
		if proto == "" {
			proto = "http"
		}
		serverAddress = proto + "://" + c.Request.Host
	}

	c.JSON(http.StatusOK, gin.H{
		"schema_version": "1.0",
		"name":           common.SystemName + " Marketplace Agent Skills",
		"description":    "Instruction set for AI agents to autonomously purchase API quota and manage tokens.",

		"prerequisites": gin.H{
			"description": "Before using these skills, the agent needs valid user credentials (username + password) provided by the account owner.",
			"required":    []string{"username", "password"},
			"optional":    []string{"preferred_models", "monthly_budget_hint"},
		},

		"skills": []gin.H{
			// ── Skill 1: Authenticate ──
			{
				"id":          "authenticate",
				"name":        "Login to New API",
				"description": "Authenticate with user credentials to obtain a session. Must be called before any protected endpoint.",
				"when_to_use": "At the start of any marketplace operation, or when a 401 response is received.",
				"request": gin.H{
					"method": "POST",
					"url":    serverAddress + "/api/user/login",
					"headers": gin.H{
						"Content-Type": "application/json",
					},
					"body": gin.H{
						"username": "{{username}}",
						"password": "{{password}}",
					},
				},
				"response_handling": gin.H{
					"success": gin.H{
						"condition": "response.success == true && !response.data.require_2fa",
						"action":    "Store the session cookie from Set-Cookie header. Proceed to next skill.",
					},
					"require_2fa": gin.H{
						"condition": "response.data.require_2fa == true",
						"action":    "STOP. Cannot proceed automatically. Notify the user that 2FA verification is required.",
					},
					"failure": gin.H{
						"condition": "response.success == false",
						"action":    "STOP. Credentials are invalid. Notify the user.",
					},
				},
				"important_notes": []string{
					"The session is cookie-based. Preserve cookies across all subsequent requests.",
					"Do NOT retry with the same credentials if login fails — notify the user instead.",
				},
			},

			// ── Skill 2: Check Account Status ──
			{
				"id":          "check_account",
				"name":        "Check Account Balance and Settings",
				"description": "Retrieve current user info including quota balance and marketplace budget settings.",
				"when_to_use": "After login, before making a purchase decision. Helps determine if a purchase is needed and affordable.",
				"request": gin.H{
					"method": "GET",
					"url":    serverAddress + "/api/user/self",
					"headers": gin.H{
						"Cookie": "{{session_cookie}}",
					},
				},
				"response_handling": gin.H{
					"key_fields": gin.H{
						"quota":      "Available balance (in quota units). This is what can be spent.",
						"used_quota": "Total quota consumed historically.",
						"setting":    "JSON string containing user settings. Parse it to find marketplace_monthly_budget.",
					},
					"decision_logic": "If quota is low and the agent needs API access, proceed to browse_models and then purchase.",
				},
			},

			// ── Skill 3: Browse Models ──
			{
				"id":          "browse_models",
				"name":        "Browse Available Models",
				"description": "Get the catalog of available AI models with pricing info. No authentication required.",
				"when_to_use": "When the agent needs to choose which model to purchase access to.",
				"request": gin.H{
					"method": "GET",
					"url":    serverAddress + "/api/marketplace/models",
					"query_params": gin.H{
						"capability": "(optional) Filter: chat_completions, embeddings, images, audio, etc.",
						"vendor":     "(optional) Filter by vendor name, e.g. 'openai', 'anthropic'",
						"q":          "(optional) Search model name, e.g. 'gpt-4'",
					},
				},
				"response_handling": gin.H{
					"success": "response.data is an array of models with: model_name, input_price, output_price, quota_type, endpoints, tags, vendor",
					"selection_advice": []string{
						"Choose models based on the task requirements (e.g., gpt-4o for complex reasoning, gpt-4o-mini for simple tasks).",
						"Consider price: lower input_price/output_price = cheaper. quota_type 0 = per-token billing, 1 = fixed price.",
						"Check endpoints to ensure the model supports the API path you need (e.g., /v1/chat/completions).",
					},
				},
			},

			// ── Skill 4: Purchase Quota ──
			{
				"id":          "purchase_quota",
				"name":        "Purchase Quota and Create Token",
				"description": "Purchase API quota from the user's balance and create a ready-to-use API token. This is the core marketplace action.",
				"when_to_use": "When the agent's current token is exhausted or about to expire, and it needs a new one with fresh quota.",
				"request": gin.H{
					"method": "POST",
					"url":    serverAddress + "/api/marketplace/purchase",
					"headers": gin.H{
						"Content-Type": "application/json",
						"Cookie":       "{{session_cookie}}",
					},
					"body": gin.H{
						"name":   "(required) Token name, e.g. 'agent-worker-20260325'. Max 50 chars, alphanumeric + dash/underscore.",
						"models": "(optional) Array of model names to restrict, e.g. [\"gpt-4o\"]. Empty = all models.",
						"quota":  "(required) Integer. Amount of quota to allocate. Reference: 500000 quota ≈ $1 USD.",
						"expire": "(optional) Token expiry: '1h', '1d', '7d', '30d', '90d', '365d'. Empty = never expires.",
					},
				},
				"response_handling": gin.H{
					"auto_approved": gin.H{
						"condition": "response.data.status == 'completed'",
						"action":    "Token is ready. Use response.data.token as the Bearer token for API calls. The endpoint is in response.data.endpoint.",
						"fields":    "token (sk-xxx), name, models, endpoint, expires_at, purchase_id",
					},
					"pending_confirmation": gin.H{
						"condition": "response.data.status == 'pending'",
						"action":    "Purchase exceeds monthly budget. The account owner has been sent a confirmation email. WAIT and retry check_account periodically, or notify the user that manual confirmation is needed.",
						"fields":    "name, models, status, purchase_id, message",
					},
					"insufficient_balance": gin.H{
						"condition": "response.success == false && message contains 'insufficient' or 'balance'",
						"action":    "STOP. Not enough quota. Notify the user to top up their account.",
					},
				},
				"quota_sizing_guide": gin.H{
					"description": "How much quota to request depends on expected usage.",
					"reference": gin.H{
						"500000_quota":   "≈ $1 USD. Good for ~500K input tokens on cheap models, or ~50K on expensive ones.",
						"2500000_quota":  "≈ $5 USD. Good for moderate daily usage.",
						"25000000_quota": "≈ $50 USD. Good for heavy multi-day workloads.",
					},
					"strategy": "Start small (e.g., 500000). If the agent runs out quickly, increase the amount on the next purchase. Avoid over-purchasing — unused quota in expired tokens is wasted.",
				},
				"important_notes": []string{
					"Rate limit: 10 purchases per hour per user.",
					"The user's monthly budget (marketplace_monthly_budget in settings) controls auto-approval. Within budget = instant. Over budget = email confirmation required.",
					"If budget is 0, ALL purchases are auto-approved (no budget limit).",
					"Quota is deducted from the user's account balance immediately on auto-approval.",
				},
			},

			// ── Skill 5: Use the Token ──
			{
				"id":          "use_token",
				"name":        "Make API Calls with Purchased Token",
				"description": "Use the token from the purchase response to call AI APIs.",
				"when_to_use": "After a successful purchase, to actually use the AI models.",
				"request": gin.H{
					"method": "POST",
					"url":    serverAddress + "/v1/chat/completions",
					"headers": gin.H{
						"Content-Type":  "application/json",
						"Authorization": "Bearer {{token}}",
					},
					"body": gin.H{
						"model":    "gpt-4o",
						"messages": []gin.H{{"role": "user", "content": "Hello"}},
					},
				},
				"error_handling": gin.H{
					"token_exhausted": gin.H{
						"condition": "HTTP 401 or response contains 'quota exhausted'",
						"action":    "Token quota is used up. Go back to purchase_quota to get a new token.",
					},
					"token_expired": gin.H{
						"condition": "response contains 'token expired'",
						"action":    "Token has expired. Go back to purchase_quota to get a new token.",
					},
				},
			},
		},

		// ── Decision Flowchart ──
		"autonomous_workflow": gin.H{
			"description": "Complete decision flow for an agent that needs AI API access.",
			"steps": []gin.H{
				{"step": 1, "action": "authenticate", "on_success": "goto step 2", "on_failure": "STOP: notify user about invalid credentials"},
				{"step": 2, "action": "check_account", "on_success": "goto step 3"},
				{"step": 3, "action": "evaluate", "logic": "If existing token works → use it. If token exhausted/expired/missing → goto step 4."},
				{"step": 4, "action": "browse_models", "logic": "Select model based on task needs and price. Then goto step 5."},
				{"step": 5, "action": "purchase_quota", "on_completed": "goto step 6", "on_pending": "Wait for owner confirmation. Retry check every 5 minutes up to 24 hours.", "on_insufficient": "STOP: notify user to top up."},
				{"step": 6, "action": "use_token", "on_exhausted": "goto step 2 (re-evaluate and purchase again)"},
			},
		},

		// ── Safety Guidelines ──
		"safety": gin.H{
			"description": "Rules the agent MUST follow to protect the user's account.",
			"rules": []string{
				"NEVER purchase more quota than needed for the immediate task.",
				"ALWAYS prefer smaller, more frequent purchases over one large purchase.",
				"If a purchase is pending confirmation, do NOT retry with a larger amount to bypass the budget.",
				"NEVER store or log the full token key (sk-xxx) in plain text in user-visible outputs.",
				"If login fails, do NOT retry more than once — the credentials may be wrong.",
				"Respect rate limits. If rate-limited (HTTP 429), wait before retrying.",
				"The agent should inform the user about every purchase it makes, including amount and reason.",
			},
		},
	})
}
