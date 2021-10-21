package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/panjichang1990/tianzong/tzlog"
	"time"
)

func Logger() gin.HandlerFunc {

	return func(c *gin.Context) {
		// 开始时间
		start := time.Now()
		// 处理请求
		c.Next()
		// 结束时间
		end := time.Now()
		//执行时间
		latency := end.Sub(start)

		path := c.Request.URL.Path

		clientIP := c.ClientIP()
		method := c.Request.Method
		statusCode := c.Writer.Status()
		tzlog.I(" %3d | %3v | %15s | %s  %s ",
			statusCode,
			latency,
			clientIP,
			method,
			path,
		)

	}
}
