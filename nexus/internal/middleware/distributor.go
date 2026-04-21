package middleware

import "github.com/gin-gonic/gin"

// Distribute picks a channel for the incoming relay request by delegating to
// routing/policy.RoutingPolicy.Pick. Default policy is "weighted" (ported
// from /home/user/new-api/middleware/distributor.go), but per-org policies
// can override (cost/latency/quality/ab/canary/shadow).
//
// Sets context keys for downstream relay:
//   - channel_id, channel_type, model, upstream_base_url
//
// TODO: implement. Read candidates from Channel cache filtered by
// (org_id, model, group); call registry.Get(policyName).Pick(ctx).
func Distribute() gin.HandlerFunc {
	return func(c *gin.Context) { c.Next() }
}
