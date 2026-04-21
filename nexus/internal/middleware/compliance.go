package middleware

import "github.com/gin-gonic/gin"

// CompliancePolicy enforces org-level compliance rules such as allowed
// regions, banned models, and PII handling options.
//
// TODO: load CompliancePolicy by org_id and reject requests that violate.
func CompliancePolicy() gin.HandlerFunc {
	return func(c *gin.Context) { c.Next() }
}
