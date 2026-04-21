package middleware

import "github.com/gin-gonic/gin"

// X402Gate implements per-request micro-settlement via the x402 protocol.
//
// Flow:
//  1. Inspect token/org: if quota sufficient, pass through.
//  2. If token.X402Enabled && org.X402Enabled:
//     - If request has no `X-PAYMENT` header, respond 402 Payment Required
//       with `X-PAYMENT-REQUIRED` header (payment/x402.RequirementsBuilder).
//     - Else decode + verify `X-PAYMENT` via payment/x402.Verifier, attach
//       pending *X402Payment to context; BillingPost will finalize.
//  3. Else reject with 402 but no payment requirements (hint user to top up).
//
// TODO: implement gating logic.
func X402Gate() gin.HandlerFunc {
	return func(c *gin.Context) { c.Next() }
}
