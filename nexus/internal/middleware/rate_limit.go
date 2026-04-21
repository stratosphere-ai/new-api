package middleware

import "github.com/gin-gonic/gin"

// RateLimit applies a global Redis-backed rate limit for /api/v2/* routes.
// Port of /home/user/new-api/middleware/rate-limit.go.
//
// TODO: port implementation.
func RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) { c.Next() }
}
