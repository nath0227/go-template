package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/your-org/service-name/pkg/logger"
)

const headerTID = "X-Trace-ID"

// TID injects a trace ID into the request context and response header.
// It reads the X-Trace-ID header if present; otherwise generates a new ID
// from the active OTel span or a fresh UUID.
func TID() gin.HandlerFunc {
	return func(c *gin.Context) {
		tid := c.GetHeader(headerTID)
		if tid == "" {
			tid = logger.NewTID(c.Request.Context())
		}
		ctx := logger.WithTID(c.Request.Context(), tid)
		c.Request = c.Request.WithContext(ctx)
		c.Header(headerTID, tid)
		c.Next()
	}
}
