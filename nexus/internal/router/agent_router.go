package router

import "github.com/gin-gonic/gin"

// SetAgentRouter mounts /agent/* endpoints for Agent orchestration.
//
//   - POST /agent/:agent_id/run          start an agent run (streams SSE)
//   - GET  /agent/:agent_id/runs/:run_id fetch run status / history
//   - POST /agent/:agent_id/tool/:name   manual tool invocation (ToolProxy)
//
// TODO: wire controller.Agent, agent.Orchestrator, agent/mcp bridge.
func SetAgentRouter(engine *gin.Engine) {
	g := engine.Group("/agent")
	_ = g
	// TODO: g.Use(middleware.TokenAuth(), middleware.TenantResolver(), middleware.RBAC())
}
