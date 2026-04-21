package router

import "github.com/gin-gonic/gin"

// SetWeb3AuthRouter mounts /api/v2/auth/web3/* for wallet sign-in.
//
// SIWE-style (EIP-4361) for Ethereum and ed25519 for Solana:
//   - POST   /api/v2/auth/web3/nonce   { chain, address } -> { nonce, message }
//   - POST   /api/v2/auth/web3/verify  { chain, address, signature } -> { jwt, user }
//   - POST   /api/v2/auth/web3/link    link a wallet to an existing user (UserAuth)
//   - DELETE /api/v2/auth/web3/link/:credential_id unlink (UserAuth)
//
// TODO: wire controller.Web3Auth which delegates to auth/web3.Web3Authenticator.
func SetWeb3AuthRouter(engine *gin.Engine) {
	g := engine.Group("/api/v2/auth/web3")
	_ = g
	// TODO: g.POST("/nonce", controller.Web3Nonce)
	// TODO: g.POST("/verify", controller.Web3Verify)
	// TODO: g.POST("/link", middleware.UserAuth(), controller.Web3Link)
	// TODO: g.DELETE("/link/:id", middleware.UserAuth(), controller.Web3Unlink)
}
