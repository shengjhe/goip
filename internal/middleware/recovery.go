package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/shengjhe/goip/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// Recovery 錯誤恢復中間件
func Recovery(logger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 記錄堆疊追蹤
				stack := debug.Stack()

				logger.Error().
					Interface("error", err).
					Str("stack", string(stack)).
					Str("path", c.Request.URL.Path).
					Str("method", c.Request.Method).
					Str("ip", c.ClientIP()).
					Msg("Panic recovered")

				// 回應錯誤
				c.JSON(http.StatusInternalServerError, model.ErrorResponse{
					Error:     "Internal server error",
					Code:      "PANIC_RECOVERED",
					Timestamp: time.Now(),
				})

				// 終止請求
				c.Abort()
			}
		}()

		c.Next()
	}
}

// Custom recovery with custom error message
func RecoveryWithWriter(logger zerolog.Logger, showDetails bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// 記錄堆疊追蹤
				stack := debug.Stack()

				logger.Error().
					Interface("error", err).
					Str("stack", string(stack)).
					Str("path", c.Request.URL.Path).
					Str("method", c.Request.Method).
					Str("ip", c.ClientIP()).
					Msg("Panic recovered")

				// 準備錯誤訊息
				errMsg := "Internal server error"
				if showDetails {
					errMsg = fmt.Sprintf("%v", err)
				}

				// 回應錯誤
				c.JSON(http.StatusInternalServerError, model.ErrorResponse{
					Error:     errMsg,
					Code:      "PANIC_RECOVERED",
					Timestamp: time.Now(),
				})

				// 終止請求
				c.Abort()
			}
		}()

		c.Next()
	}
}
