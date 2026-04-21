package router

import "github.com/gin-gonic/gin"

// SetOrgRouter mounts /api/v2/org/* for multi-tenant organization management.
//
// Endpoints:
//   - POST   /api/v2/org                 create org (owner = caller)
//   - GET    /api/v2/org                 list orgs current user belongs to
//   - GET    /api/v2/org/:id             get org detail
//   - PATCH  /api/v2/org/:id             update settings (owner/admin)
//   - DELETE /api/v2/org/:id             delete org (owner)
//   - POST   /api/v2/org/:id/members     invite member
//   - DELETE /api/v2/org/:id/members/:uid remove member
//   - PATCH  /api/v2/org/:id/members/:uid/role change role
//
// TODO: wire controller.Organization, controller.OrgMember, controller.Role.
func SetOrgRouter(engine *gin.Engine) {
	g := engine.Group("/api/v2/org")
	_ = g
	// TODO: g.Use(middleware.UserAuth())
}
