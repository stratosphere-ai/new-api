package middleware

import "github.com/gin-gonic/gin"

// OTel starts an OpenTelemetry root span for the request and injects the
// span context into the gin.Context for downstream components.
//
// TODO: initialize via observability/otel package; tag span with
// org_id, token_id, model, channel_id.
func OTel() gin.HandlerFunc {
	return func(c *gin.Context) { c.Next() }
}
