package middleware

import "github.com/gin-gonic/gin"

// RBAC enforces a minimum org role for the route.
//
// Example: RBAC("member") on /v1/*, RBAC("admin") on /api/v2/org/:id settings,
// RBAC("owner") on org-wide dangerous operations.
//
// TODO: read org_member_role from context (set by TenantResolver) and reject
// with 403 if below the required level.
func RBAC(minRole string) gin.HandlerFunc {
	return func(c *gin.Context) { _ = minRole; c.Next() }
}
