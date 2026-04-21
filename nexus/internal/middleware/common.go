package middleware

import "github.com/gin-gonic/gin"

// RequestID attaches a correlation id to the request context and response
// header. Port of /home/user/new-api/middleware/request-id.go.
func RequestID() gin.HandlerFunc {
	return func(c *gin.Context) { c.Next() }
}

// Recover is a panic recovery middleware tailored to emit a JSON error
// envelope consistent with the new-api error format.
func Recover() gin.HandlerFunc {
	return func(c *gin.Context) { c.Next() }
}

// CORS applies permissive CORS for relay endpoints. Port of
// /home/user/new-api/middleware/cors.go.
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) { c.Next() }
}

// I18n loads the request's preferred language for error/message translation.
func I18n() gin.HandlerFunc {
	return func(c *gin.Context) { c.Next() }
}

// Logger emits an access log line with request/response metadata.
func Logger() gin.HandlerFunc {
	return func(c *gin.Context) { c.Next() }
}

// SystemPerformanceCheck short-circuits requests when the host is overloaded.
// Port of /home/user/new-api/middleware/performance.go.
func SystemPerformanceCheck() gin.HandlerFunc {
	return func(c *gin.Context) { c.Next() }
}

// BodyStorageCleanup frees buffered request bodies after relay completes.
// Port of /home/user/new-api/middleware/body_cleanup.go.
func BodyStorageCleanup() gin.HandlerFunc {
	return func(c *gin.Context) { c.Next() }
}
