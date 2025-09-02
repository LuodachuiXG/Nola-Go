package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ZapLogger Zap 请求日志日志中间件
func ZapLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 继续请求
		c.Next()

		// 结束时间
		cost := time.Since(start)

		// 输出日志
		logger.Info("HTTP Request",
			zap.Int("status", c.Writer.Status()),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", query),
			zap.String("ip", c.ClientIP()),
			zap.String("user-agent", c.Request.UserAgent()),
			zap.Duration("latency", cost),
			zap.String("error", c.Errors.ByType(gin.ErrorTypePrivate).String()),
		)
	}
}

// ZapRecovery Zap 错误日志中间件
func ZapRecovery(logger *zap.Logger, stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				if stack {
					logger.Error("[Recover from panic]",
						zap.Any("error", err),
						zap.String("path", c.Request.URL.Path),
						zap.Stack("stacktrace"),
					)
				} else {
					logger.Error(
						"[Recover from panic]",
						zap.Any("error", err),
						zap.String("path", c.Request.URL.Path),
					)
				}

				// 返回 500
				c.AbortWithStatus(500)
			}
		}()
		c.Next()
	}
}
