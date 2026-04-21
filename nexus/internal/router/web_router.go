package router

import "github.com/gin-gonic/gin"

// SetWebRouter serves the SPA frontend. The new-api project embeds build
// assets via go:embed in /home/user/new-api/router/web-router.go. For the
// skeleton we leave this as a placeholder.
//
// TODO: port embed.FS + index.html fallback for client-side routing.
func SetWebRouter(engine *gin.Engine) {
	_ = engine
}
