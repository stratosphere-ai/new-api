package middleware

import "github.com/gin-gonic/gin"

// ContentSafetyPre invokes the org-configured SafetyFilter.PreCheck.
// Blocks the request with 451/400 when verdict is Reject.
//
// TODO: plug observability/safety.SafetyFilter.
func ContentSafetyPre() gin.HandlerFunc {
	return func(c *gin.Context) { c.Next() }
}

// ContentSafetyPost filters the outgoing response (streaming-safe).
// For streams, wraps the writer and inspects chunks incrementally.
//
// TODO: implement streaming-aware post-filter.
func ContentSafetyPost() gin.HandlerFunc {
	return func(c *gin.Context) { c.Next() }
}
