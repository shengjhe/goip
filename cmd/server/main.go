package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/axiom/goip/config"
	"github.com/axiom/goip/internal/handler"
	"github.com/axiom/goip/internal/middleware"
	"github.com/axiom/goip/internal/repository"
	"github.com/axiom/goip/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	// 載入配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	if err := cfg.Validate(); err != nil {
		log.Fatal().Err(err).Msg("Invalid config")
	}

	// 初始化 Logger
	logger := initLogger(cfg.Log)

	logger.Info().Msg("Starting GoIP service...")

	// 初始化 MaxMind Repository
	maxmindRepo, err := repository.NewMaxMindRepository(cfg.MaxMind.DBPath)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize MaxMind repository")
	}
	defer maxmindRepo.Close()

	logger.Info().Str("db_path", cfg.MaxMind.DBPath).Msg("MaxMind DB loaded")

	// 初始化 Redis Client
	redisClient := initRedis(cfg.Redis, logger)
	defer redisClient.Close()

	// 測試 Redis 連接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Warn().Err(err).Msg("Redis connection failed, cache will be disabled")
	} else {
		logger.Info().Str("host", cfg.Redis.Host).Int("port", cfg.Redis.Port).Msg("Redis connected")
	}

	// 初始化 Cache Repository
	cacheRepo := repository.NewCacheRepository(redisClient)

	// 初始化 Service
	ipService := service.NewIPService(
		maxmindRepo,
		cacheRepo,
		logger,
		cfg.Cache.TTL,
	)

	// 初始化 Handler
	ipHandler := handler.NewIPHandler(
		ipService,
		cacheRepo,
		maxmindRepo,
		logger,
		cfg.Batch.MaxSize,
	)

	// 初始化 Gin
	router := setupRouter(cfg, ipHandler, redisClient, logger)

	// 啟動 HTTP Server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// 在 goroutine 中啟動伺服器
	go func() {
		logger.Info().Int("port", cfg.Server.Port).Msg("HTTP server started")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("Failed to start server")
		}
	}()

	// 優雅關閉
	gracefulShutdown(srv, cfg.Server.ShutdownTimeout, logger)
}

// initLogger 初始化 Logger
func initLogger(cfg config.LogConfig) zerolog.Logger {
	// 設定日誌級別
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// 設定輸出格式
	if cfg.Format == "console" {
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	}

	return log.Logger
}

// initRedis 初始化 Redis 客戶端
func initRedis(cfg config.RedisConfig, logger zerolog.Logger) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		MaxRetries:   cfg.MaxRetries,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	return client
}

// setupRouter 設定路由
func setupRouter(
	cfg *config.Config,
	ipHandler *handler.IPHandler,
	redisClient *redis.Client,
	logger zerolog.Logger,
) *gin.Engine {
	// 設定 Gin 模式
	if cfg.Log.Level == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// 全域中間件
	router.Use(middleware.Recovery(logger))
	router.Use(middleware.Logger(logger))

	// 限流中間件（如果啟用）
	if cfg.RateLimit.Enabled {
		rateLimiter := middleware.NewRateLimiter(
			redisClient,
			logger,
			cfg.RateLimit.RequestsPerMinute,
			cfg.RateLimit.RequestsPerHour,
		)
		router.Use(rateLimiter.Limit())
	}

	// API 路由群組
	v1 := router.Group("/api/v1")
	{
		// IP 查詢
		v1.GET("/ip/:ip", ipHandler.HandleIPLookup)
		v1.POST("/ip/batch", ipHandler.HandleBatchLookup)

		// 系統
		v1.GET("/health", ipHandler.HandleHealth)
		v1.GET("/stats", ipHandler.HandleStats)

		// 快取管理
		cache := v1.Group("/cache")
		{
			cache.GET("/stats", ipHandler.HandleCacheStats)
			cache.POST("/invalidate", ipHandler.HandleInvalidateCache)
		}
	}

	// 根路徑
	router.GET("/", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"service": "GoIP",
			"version": "1.0.0",
			"status":  "running",
		})
	})

	return router
}

// gracefulShutdown 優雅關閉
func gracefulShutdown(srv *http.Server, timeout time.Duration, logger zerolog.Logger) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	<-quit
	logger.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error().Err(err).Msg("Server forced to shutdown")
	}

	logger.Info().Msg("Server exited")
}
