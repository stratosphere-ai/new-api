package middleware

import "github.com/gin-gonic/gin"

// ModelRequestRateLimit enforces per-(org, model, token) RPM/TPM ceilings.
// Port of /home/user/new-api/middleware/model-rate-limit.go with OrgID
// injected into the Redis key.
//
// TODO: port implementation.
func ModelRequestRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) { c.Next() }
}
