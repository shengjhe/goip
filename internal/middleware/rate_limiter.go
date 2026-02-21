package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/shengjhe/goip/internal/model"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

const (
	rateLimitKeyPrefix = "goip:ratelimit:"
)

// RateLimiter Redis 限流中間件
type RateLimiter struct {
	client          *redis.Client
	logger          zerolog.Logger
	requestsPerMin  int
	requestsPerHour int
}

// NewRateLimiter 建立新的限流中間件
func NewRateLimiter(client *redis.Client, logger zerolog.Logger, rpm, rph int) *RateLimiter {
	return &RateLimiter{
		client:          client,
		logger:          logger,
		requestsPerMin:  rpm,
		requestsPerHour: rph,
	}
}

// Limit 限流中間件
func (rl *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		ctx := c.Request.Context()

		// 檢查分鐘級限流
		if rl.requestsPerMin > 0 {
			allowed, retryAfter, err := rl.checkLimit(ctx, ip, "minute", rl.requestsPerMin, time.Minute)
			if err != nil {
				rl.logger.Warn().Err(err).Str("ip", ip).Msg("Rate limit check failed, allowing request")
			} else if !allowed {
				rl.respondRateLimitExceeded(c, retryAfter)
				return
			}
		}

		// 檢查小時級限流
		if rl.requestsPerHour > 0 {
			allowed, retryAfter, err := rl.checkLimit(ctx, ip, "hour", rl.requestsPerHour, time.Hour)
			if err != nil {
				rl.logger.Warn().Err(err).Str("ip", ip).Msg("Rate limit check failed, allowing request")
			} else if !allowed {
				rl.respondRateLimitExceeded(c, retryAfter)
				return
			}
		}

		c.Next()
	}
}

// checkLimit 使用 Sorted Set 實現滑動窗口限流
func (rl *RateLimiter) checkLimit(ctx context.Context, ip, window string, limit int, duration time.Duration) (bool, int64, error) {
	key := rateLimitKeyPrefix + ip + ":" + window
	now := time.Now().UnixNano()
	windowStart := now - duration.Nanoseconds()

	pipe := rl.client.Pipeline()

	// 1. 移除過期的記錄
	pipe.ZRemRangeByScore(ctx, key, "0", fmt.Sprintf("%d", windowStart))

	// 2. 獲取當前窗口內的請求數
	zcard := pipe.ZCard(ctx, key)

	// 3. 添加當前請求
	pipe.ZAdd(ctx, key, redis.Z{
		Score:  float64(now),
		Member: now,
	})

	// 4. 設定過期時間
	pipe.Expire(ctx, key, duration+time.Minute)

	_, err := pipe.Exec(ctx)
	if err != nil {
		return false, 0, err
	}

	count := zcard.Val()

	// 如果超過限制
	if count >= int64(limit) {
		// 計算需要等待的時間
		oldestCmd := rl.client.ZRange(ctx, key, 0, 0)
		oldest, err := oldestCmd.Result()
		if err != nil || len(oldest) == 0 {
			return false, int64(duration.Seconds()), nil
		}

		var oldestTime int64
		fmt.Sscanf(oldest[0], "%d", &oldestTime)
		retryAfter := (oldestTime + duration.Nanoseconds() - now) / 1e9

		if retryAfter < 0 {
			retryAfter = 1
		}

		return false, retryAfter, nil
	}

	return true, 0, nil
}

// respondRateLimitExceeded 回應限流錯誤
func (rl *RateLimiter) respondRateLimitExceeded(c *gin.Context, retryAfter int64) {
	c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))
	c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", rl.requestsPerMin))
	c.Header("X-RateLimit-Remaining", "0")

	c.JSON(http.StatusTooManyRequests, model.ErrorResponse{
		Error:     fmt.Sprintf("Rate limit exceeded. Retry after %d seconds", retryAfter),
		Code:      "RATE_LIMIT_EXCEEDED",
		Timestamp: time.Now(),
	})

	c.Abort()
}
