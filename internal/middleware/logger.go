package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// Logger 日誌中間件
func Logger(logger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 開始時間
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery

		// 處理請求
		c.Next()

		// 計算耗時
		latency := time.Since(start)
		statusCode := c.Writer.Status()
		clientIP := c.ClientIP()
		method := c.Request.Method
		userAgent := c.Request.UserAgent()

		// 組裝日誌
		logEvent := logger.Info()

		// 如果有錯誤，改用 Error 級別
		if len(c.Errors) > 0 {
			logEvent = logger.Error().Strs("errors", c.Errors.Errors())
		} else if statusCode >= 500 {
			logEvent = logger.Error()
		} else if statusCode >= 400 {
			logEvent = logger.Warn()
		}

		logEvent.
			Str("method", method).
			Str("path", path).
			Str("query", query).
			Int("status", statusCode).
			Dur("latency", latency).
			Str("ip", clientIP).
			Str("user_agent", userAgent).
			Msg("HTTP Request")
	}
}
