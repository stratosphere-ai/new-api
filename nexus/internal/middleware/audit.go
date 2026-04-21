package middleware

import "github.com/gin-gonic/gin"

// AuditPre captures a redacted prompt snapshot before the relay controller
// runs, so that AuditPost can correlate prompt and response.
//
// TODO: wire observability/audit.AuditLogger.
func AuditPre() gin.HandlerFunc {
	return func(c *gin.Context) { c.Next() }
}

// AuditPost emits the final AuditLog row (deferred until response completes).
//
// TODO: implement defer + flush to observability/audit.AuditLogger.
func AuditPost() gin.HandlerFunc {
	return func(c *gin.Context) { c.Next() }
}
