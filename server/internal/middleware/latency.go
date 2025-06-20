package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

func XResponseTime() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		latency := time.Since(start)
		c.Writer.Header().Set("X-Response-Time", latency.String())
	}
}
