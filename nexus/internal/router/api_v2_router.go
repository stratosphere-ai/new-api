package router

import "github.com/gin-gonic/gin"

// SetApiV2Router mounts /api/v2/* endpoints for the dashboard.
//
// Ported groups (see /home/user/new-api/router/api-router.go):
//   - /api/v2/user/*          : UserAuth (or public for register/login)
//   - /api/v2/token/*         : UserAuth
//   - /api/v2/channel/*       : AdminAuth (sensitive: RootAuth + SecureVerification)
//   - /api/v2/log/*           : UserAuth
//   - /api/v2/option/*        : RootAuth
//   - /api/v2/subscription/*  : UserAuth / AdminAuth
//
// TODO: attach UserAuth/AdminAuth/RootAuth/TenantResolver middleware;
// register controller handlers.
func SetApiV2Router(engine *gin.Engine) {
	v2 := engine.Group("/api/v2")
	_ = v2
	// TODO: v2.Use(middleware.TenantResolver(), middleware.RBAC())
	// TODO: register sub-groups once controllers are ported.
}
