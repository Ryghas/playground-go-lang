package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

// RequestLogger mencatat setiap request beserta error yang ditempel handler via c.Error.
func RequestLogger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		attrs := []any{
			"method", c.Request.Method,
			"path", c.FullPath(),
			"status", c.Writer.Status(),
			"latency_ms", time.Since(start).Milliseconds(),
			"client_ip", c.ClientIP(),
		}

		switch {
		case c.Writer.Status() >= 500:
			log.Error("request", append(attrs, "errors", c.Errors.String())...)
		case len(c.Errors) > 0:
			log.Warn("request", append(attrs, "errors", c.Errors.String())...)
		default:
			log.Info("request", attrs...)
		}
	}
}
