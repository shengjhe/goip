package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// generateRequestID 生成請求 ID
func generateRequestID() string {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		// Fallback to timestamp-based ID
		return hex.EncodeToString([]byte(time.Now().String()))[:32]
	}
	return hex.EncodeToString(b)
}

// Logger 日誌中間件
func Logger(logger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 生成請求 ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		// 將 request_id 設置到 context 中，供後續使用
		c.Set("request_id", requestID)
		// 設置回應 header
		c.Writer.Header().Set("X-Request-ID", requestID)

		// 開始時間
		start := time.Now()
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		method := c.Request.Method
		clientIP := c.ClientIP()
		userAgent := c.Request.UserAgent()

		// 過濾 healthcheck 請求的日誌
		isHealthCheck := path == "/healthz" || path == "/health"

		// 記錄請求日誌（排除 healthcheck）
		if !isHealthCheck {
			logger.Info().
				Str("request_id", requestID).
				Str("type", "request").
				Str("method", method).
				Str("path", path).
				Str("query", query).
				Str("client_ip", clientIP).
				Str("user_agent", userAgent).
				Send()
		}

		// 處理請求
		c.Next()

		// 計算耗時
		latency := time.Since(start)
		statusCode := c.Writer.Status()

		// 記錄回應日誌（排除 healthcheck）
		if !isHealthCheck {
			// 組裝回應日誌
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
				Str("request_id", requestID).
				Str("type", "response").
				Str("method", method).
				Str("path", path).
				Int("status", statusCode).
				Int64("latency_ms", latency.Milliseconds()).
				Str("client_ip", clientIP)

			// 如果有 source 資訊，加入日誌
			if source, exists := c.Get("source"); exists {
				if sourceStr, ok := source.(string); ok {
					logEvent.Str("source", sourceStr)
				}
			}

			// 如果有 provider 資訊，加入日誌
			if provider, exists := c.Get("provider"); exists {
				if providerStr, ok := provider.(string); ok {
					logEvent.Str("provider", providerStr)
				}
			}

			logEvent.Send()
		}
	}
}
