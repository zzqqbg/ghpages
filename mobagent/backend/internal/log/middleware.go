package log

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func Recover() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if rec := recover(); rec != nil {
				Logger().Error("panic recovered",
					zap.Any("panic", rec),
					zap.String("method", c.Request.Method),
					zap.String("path", c.Request.URL.Path),
					zap.String("client_ip", c.ClientIP()),
					zap.ByteString("stack", debug.Stack()),
				)
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "internal server error",
				})
			}
		}()
		c.Next()
	}
}

func RequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		method := c.Request.Method
		clientIP := c.ClientIP()
		ua := c.Request.UserAgent()

		c.Next()

		latency := time.Since(start)
		status := c.Writer.Status()
		size := c.Writer.Size()
		errs := c.Errors.ByType(gin.ErrorTypePrivate).String()

		fields := []zap.Field{
			zap.String("method", method),
			zap.String("path", path),
			zap.Int("status", status),
			zap.Duration("latency", latency),
			zap.Int("bytes", size),
			zap.String("client_ip", clientIP),
		}
		if query != "" {
			fields = append(fields, zap.String("query", query))
		}
		if ua != "" {
			fields = append(fields, zap.String("user_agent", truncate(ua, 120)))
		}
		if errs != "" {
			fields = append(fields, zap.String("errors", errs))
		}

		msg := fmt.Sprintf("%s %s", method, path)
		switch {
		case status >= 500:
			Logger().Error(msg, fields...)
		case status >= 400:
			Logger().Warn(msg, fields...)
		case path == "/health" || path == "/ws":
			Logger().Debug(msg, fields...)
		default:
			Logger().Info(msg, fields...)
		}
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
