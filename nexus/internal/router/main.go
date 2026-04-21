package router

import "github.com/gin-gonic/gin"

// SetRouter wires all top-level route groups onto the given engine.
//
// Mirrors /home/user/new-api/router/main.go SetRouter() but adds new prefixes
// for org, web3 auth, x402 payment, and agent endpoints.
func SetRouter(engine *gin.Engine) {
	SetApiV2Router(engine)
	SetOrgRouter(engine)
	SetWeb3AuthRouter(engine)
	SetX402Router(engine)
	SetAgentRouter(engine)
	SetRelayRouter(engine)
	SetWebRouter(engine)

	engine.GET("/internal/healthz", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})
}
