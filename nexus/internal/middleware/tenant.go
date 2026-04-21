package middleware

import "github.com/gin-gonic/gin"

// TenantResolver resolves the acting organization for a request.
//
// Resolution order (first match wins):
//  1. Token's bound OrgID (from TokenAuth context).
//  2. X-Org-ID header (if user is a member of that org).
//  3. Subdomain, e.g. acme.example.com -> org slug "acme".
//  4. User.DefaultOrgID fallback.
//
// Populates context keys:
//   - org_id uint
//   - org_slug string
//   - org_member_role string
//
// TODO: implement with tenant.OrgResolver.
func TenantResolver() gin.HandlerFunc {
	return func(c *gin.Context) { c.Next() }
}
