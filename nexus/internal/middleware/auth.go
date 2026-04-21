package middleware

import "github.com/gin-gonic/gin"

// UserAuth requires a valid session/JWT-authenticated user.
// Ported responsibility of /home/user/new-api/middleware/auth.go UserAuth().
//
// TODO: implement session/JWT lookup and role check (RoleCommonUser+).
func UserAuth() gin.HandlerFunc {
	return func(c *gin.Context) { c.Next() }
}

// AdminAuth requires RoleAdminUser or above.
func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) { c.Next() }
}

// RootAuth requires RoleRootUser.
func RootAuth() gin.HandlerFunc {
	return func(c *gin.Context) { c.Next() }
}

// TokenAuth validates an API token (Bearer, x-api-key, or ?key=).
// Ported from /home/user/new-api/middleware/auth.go TokenAuth().
//
// TODO: look up token, check status/quota/IP allowlist/model whitelist,
// populate context keys (user_id, token_id, org_id).
func TokenAuth() gin.HandlerFunc {
	return func(c *gin.Context) { c.Next() }
}

// Web3JWT accepts JWTs issued by the Web3 sign-in flow.
func Web3JWT() gin.HandlerFunc {
	return func(c *gin.Context) { c.Next() }
}
