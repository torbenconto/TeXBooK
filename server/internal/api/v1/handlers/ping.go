package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func Ping(c *gin.Context) {
	start := time.Now()

	latency := time.Since(start)

	c.JSON(http.StatusOK, gin.H{
		"message": "ok",
		"latency": latency.Milliseconds(),
	})
}
