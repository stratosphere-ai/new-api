package router

import "github.com/gin-gonic/gin"

// SetX402Router mounts /pay/x402/* endpoints for the Coinbase x402 protocol.
//
// The x402 protocol reuses HTTP 402 Payment Required to enable per-request
// micro-settlement without pre-loading quota. Endpoints:
//
//   - POST /pay/x402/quote               return a PaymentRequirements header
//                                        for a given (model, estUSD)
//   - POST /pay/x402/settle              client-driven settle (rare; usually
//                                        settlement is embedded in relay via
//                                        middleware.X402Gate)
//   - POST /pay/x402/facilitator/callback Facilitator webhook (on-chain or
//                                         signed-intent settlement notify)
//
// TODO: wire controller.PaymentX402, delegating to payment/x402.Verifier and
// payment/x402.Settler.
func SetX402Router(engine *gin.Engine) {
	g := engine.Group("/pay/x402")
	_ = g
	// TODO: register handlers.
}
