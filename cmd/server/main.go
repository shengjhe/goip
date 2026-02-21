package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/shengjhe/goip/config"
	"github.com/shengjhe/goip/internal/handler"
	"github.com/shengjhe/goip/internal/middleware"
	"github.com/shengjhe/goip/internal/repository"
	"github.com/shengjhe/goip/internal/service"
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

	// 初始化 GeoIP Repository（支持多提供者）
	geoipRepo, err := initGeoIPRepository(cfg, logger)
	if err != nil {
		logger.Fatal().Err(err).Msg("Failed to initialize GeoIP repository")
	}
	defer geoipRepo.Close()

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

		// 檢查是否需要清空緩存
		if os.Getenv("FLUSH_DNS") == "true" {
			if err := flushDNSCache(ctx, redisClient, logger); err != nil {
				logger.Warn().Err(err).Msg("Failed to flush DNS cache")
			} else {
				logger.Info().Msg("DNS cache flushed successfully")
			}
		}
	}

	// 初始化 Cache Repository
	cacheRepo := repository.NewCacheRepository(redisClient)

	// 初始化 Service
	ipService := service.NewIPService(
		geoipRepo,
		cacheRepo,
		logger,
		cfg.Cache.TTL,
	)

	// 初始化 Handler
	ipHandler := handler.NewIPHandler(
		ipService,
		cacheRepo,
		geoipRepo,
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

// flushDNSCache 清空 DNS 緩存
func flushDNSCache(ctx context.Context, redisClient *redis.Client, logger zerolog.Logger) error {
	// 掃描所有 goip: 開頭的 key
	var cursor uint64
	var deletedCount int64

	for {
		keys, newCursor, err := redisClient.Scan(ctx, cursor, "goip:*", 100).Result()
		if err != nil {
			return fmt.Errorf("failed to scan keys: %w", err)
		}

		if len(keys) > 0 {
			deleted, err := redisClient.Del(ctx, keys...).Result()
			if err != nil {
				return fmt.Errorf("failed to delete keys: %w", err)
			}
			deletedCount += deleted
		}

		cursor = newCursor
		if cursor == 0 {
			break
		}
	}

	logger.Info().Int64("deleted_keys", deletedCount).Msg("Flushed DNS cache")
	return nil
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
	} else {
		// JSON 格式（預設）
		zerolog.TimeFieldFormat = time.RFC3339
		log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
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

	// 簡單的健康檢查端點（不記錄日誌）
	router.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})
	router.GET("/health", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// API 路由群組
	v1 := router.Group("/api/v1")
	{
		// IP 查詢
		v1.GET("/ip/:ip", ipHandler.HandleIPLookup)
		v1.GET("/ip/:ip/provider", ipHandler.HandleIPLookupByProvider)
		v1.POST("/ip/batch", ipHandler.HandleBatchLookup)

		// 系統
		v1.GET("/health", ipHandler.HandleHealth)
		v1.GET("/stats", ipHandler.HandleStats)
		v1.GET("/providers", ipHandler.HandleGetProviders)

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

// initGeoIPRepository 初始化 GeoIP Repository
func initGeoIPRepository(cfg *config.Config, logger zerolog.Logger) (repository.GeoIPRepository, error) {
	// 優先使用新的多提供者配置
	if len(cfg.GeoIP.Providers) > 0 {
		return initMultiProviderRepository(cfg.GeoIP.Providers, logger)
	}

	// 向後相容：使用舊的 MaxMind 配置
	if cfg.MaxMind.DBPath != "" {
		logger.Warn().Msg("Using legacy maxmind.db_path configuration. Consider migrating to geoip.providers")
		maxmindRepo, err := repository.NewMaxMindRepository(cfg.MaxMind.DBPath)
		if err != nil {
			return nil, err
		}
		logger.Info().Str("db_path", cfg.MaxMind.DBPath).Msg("MaxMind DB loaded (legacy mode)")
		return maxmindRepo, nil
	}

	return nil, fmt.Errorf("no GeoIP database configured")
}

// initMultiProviderRepository 初始化多提供者 Repository
func initMultiProviderRepository(providers []config.ProviderConfig, logger zerolog.Logger) (repository.GeoIPRepository, error) {
	var providerInfos []repository.ProviderInfo

	for i, providerCfg := range providers {
		var geoipRepo repository.GeoIPRepository
		var err error

		switch providerCfg.Type {
		case "maxmind":
			geoipRepo, err = repository.NewMaxMindRepository(providerCfg.DBPath)
			if err != nil {
				return nil, fmt.Errorf("failed to initialize MaxMind provider %d: %w", i, err)
			}
			logger.Info().
				Str("type", "maxmind").
				Str("db_path", providerCfg.DBPath).
				Int("priority", providerCfg.Priority).
				Str("region", providerCfg.Region).
				Msg("MaxMind DB loaded")

		case "ipip":
			geoipRepo, err = repository.NewIPIPRepository(providerCfg.DBPath)
			if err != nil {
				return nil, fmt.Errorf("failed to initialize IPIP provider %d: %w", i, err)
			}
			logger.Info().
				Str("type", "ipip").
				Str("db_path", providerCfg.DBPath).
				Int("priority", providerCfg.Priority).
				Str("region", providerCfg.Region).
				Msg("IPIP DB loaded")

		case "ip-api", "ipinfo", "ipapi.co":
			geoipRepo, err = repository.NewExternalAPIRepository(repository.ExternalAPIType(providerCfg.Type))
			if err != nil {
				return nil, fmt.Errorf("failed to initialize external API provider %d: %w", i, err)
			}
			logger.Info().
				Str("type", providerCfg.Type).
				Int("priority", providerCfg.Priority).
				Str("region", providerCfg.Region).
				Msg("External API provider loaded")

		default:
			return nil, fmt.Errorf("unknown provider type: %s", providerCfg.Type)
		}

		providerInfos = append(providerInfos, repository.ProviderInfo{
			Provider: geoipRepo,
			Priority: providerCfg.Priority,
			Region:   providerCfg.Region,
		})
	}

	multiRepo, err := repository.NewMultiProviderRepository(providerInfos)
	if err != nil {
		return nil, fmt.Errorf("failed to create multi-provider repository: %w", err)
	}

	logger.Info().Int("provider_count", len(providerInfos)).Msg("Multi-provider GeoIP repository initialized")
	return multiRepo, nil
}
