package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

// SlowRequests 记录耗时超过 threshold 的请求（地址、状态码、耗时），用来排查网站偶尔变慢。
// 只记地址路径，不记参数和请求内容。
func SlowRequests(threshold time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		if elapsed := time.Since(start); elapsed >= threshold {
			log.Printf("慢请求 %s %s 状态 %d 耗时 %s", c.Request.Method, c.Request.URL.Path, c.Writer.Status(), elapsed.Round(time.Millisecond))
		}
	}
}
