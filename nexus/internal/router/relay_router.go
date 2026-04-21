package router

import "github.com/gin-gonic/gin"

// SetRelayRouter mounts AI-proxy routes. Prefixes are kept identical to
// /home/user/new-api/router/relay-router.go for client compatibility.
//
// Middleware chain (outer -> inner, see plan §3.2):
//
//	RequestId -> Recover -> CORS -> I18n -> Logger/Stats -> OTel
//	-> Decompress -> SystemPerformanceCheck
//	-> TokenAuth -> TenantResolver -> RBAC
//	-> ContentSafetyPre -> ModelRequestRateLimit -> CompliancePolicy
//	-> X402Gate -> Distribute -> AuditPre -> controller.Relay
//	-> AuditPost/ContentSafetyPost(defer) -> BillingPost
//
// Route groups:
//   - /v1/*, /v1beta/*                 OpenAI/Claude/Gemini unified relay
//   - /mj/*, /suno/*                   async task relay (Midjourney, Suno)
//   - /pg/*                            dashboard playground (UserAuth)
//
// TODO: port handlers from /home/user/new-api/router/relay-router.go and
// replace service.Distribute with routing/policy.RoutingPolicy.Pick.
func SetRelayRouter(engine *gin.Engine) {
	v1 := engine.Group("/v1")
	_ = v1
	v1beta := engine.Group("/v1beta")
	_ = v1beta
	mj := engine.Group("/mj")
	_ = mj
	suno := engine.Group("/suno")
	_ = suno
	pg := engine.Group("/pg")
	_ = pg
	// TODO: attach relay middleware chain and register endpoint handlers.
}
