package config

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
)

// Config 應用程式配置
type Config struct {
	Server    ServerConfig    `mapstructure:"server"`
	MaxMind   MaxMindConfig   `mapstructure:"maxmind"`
	Redis     RedisConfig     `mapstructure:"redis"`
	Cache     CacheConfig     `mapstructure:"cache"`
	RateLimit RateLimitConfig `mapstructure:"rate_limit"`
	Batch     BatchConfig     `mapstructure:"batch"`
	Log       LogConfig       `mapstructure:"log"`
}

// ServerConfig 伺服器配置
type ServerConfig struct {
	Port            int           `mapstructure:"port"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

// MaxMindConfig MaxMind 配置
type MaxMindConfig struct {
	DBPath         string        `mapstructure:"db_path"`
	AutoUpdate     bool          `mapstructure:"auto_update"`
	UpdateInterval time.Duration `mapstructure:"update_interval"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Host         string        `mapstructure:"host"`
	Port         int           `mapstructure:"port"`
	Password     string        `mapstructure:"password"`
	DB           int           `mapstructure:"db"`
	PoolSize     int           `mapstructure:"pool_size"`
	MinIdleConns int           `mapstructure:"min_idle_conns"`
	MaxRetries   int           `mapstructure:"max_retries"`
	DialTimeout  time.Duration `mapstructure:"dial_timeout"`
	ReadTimeout  time.Duration `mapstructure:"read_timeout"`
	WriteTimeout time.Duration `mapstructure:"write_timeout"`
}

// CacheConfig 快取配置
type CacheConfig struct {
	Enabled           bool          `mapstructure:"enabled"`
	TTL               time.Duration `mapstructure:"ttl"`
	LocalCacheEnabled bool          `mapstructure:"local_cache_enabled"`
	LocalCacheSize    int           `mapstructure:"local_cache_size"`
	LocalCacheTTL     time.Duration `mapstructure:"local_cache_ttl"`
}

// RateLimitConfig 限流配置
type RateLimitConfig struct {
	Enabled           bool   `mapstructure:"enabled"`
	RequestsPerMinute int    `mapstructure:"requests_per_minute"`
	RequestsPerHour   int    `mapstructure:"requests_per_hour"`
	Burst             int    `mapstructure:"burst"`
	Storage           string `mapstructure:"storage"` // redis 或 memory
}

// BatchConfig 批次查詢配置
type BatchConfig struct {
	MaxSize int `mapstructure:"max_size"`
}

// LogConfig 日誌配置
type LogConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"` // json 或 console
	Output string `mapstructure:"output"` // stdout 或檔案路徑
}

// Load 載入配置
func Load() (*Config, error) {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("./config")
	viper.AddConfigPath("/etc/goip")

	// 設定預設值
	setDefaults()

	// 環境變數優先
	viper.AutomaticEnv()

	// 讀取配置檔案（可選）
	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("failed to read config file: %w", err)
		}
		// 配置檔案不存在時使用環境變數和預設值
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	return &config, nil
}

// setDefaults 設定預設值
func setDefaults() {
	// Server
	viper.SetDefault("server.port", 8080)
	viper.SetDefault("server.read_timeout", "10s")
	viper.SetDefault("server.write_timeout", "10s")
	viper.SetDefault("server.shutdown_timeout", "30s")

	// MaxMind
	viper.SetDefault("maxmind.db_path", "./data/GeoLite2-Country.mmdb")
	viper.SetDefault("maxmind.auto_update", false)
	viper.SetDefault("maxmind.update_interval", "24h")

	// Redis
	viper.SetDefault("redis.host", "localhost")
	viper.SetDefault("redis.port", 6379)
	viper.SetDefault("redis.password", "")
	viper.SetDefault("redis.db", 0)
	viper.SetDefault("redis.pool_size", 10)
	viper.SetDefault("redis.min_idle_conns", 5)
	viper.SetDefault("redis.max_retries", 3)
	viper.SetDefault("redis.dial_timeout", "5s")
	viper.SetDefault("redis.read_timeout", "3s")
	viper.SetDefault("redis.write_timeout", "3s")

	// Cache
	viper.SetDefault("cache.enabled", true)
	viper.SetDefault("cache.ttl", "24h")
	viper.SetDefault("cache.local_cache_enabled", false)
	viper.SetDefault("cache.local_cache_size", 1000)
	viper.SetDefault("cache.local_cache_ttl", "5m")

	// Rate Limit
	viper.SetDefault("rate_limit.enabled", true)
	viper.SetDefault("rate_limit.requests_per_minute", 100)
	viper.SetDefault("rate_limit.requests_per_hour", 5000)
	viper.SetDefault("rate_limit.burst", 10)
	viper.SetDefault("rate_limit.storage", "redis")

	// Batch
	viper.SetDefault("batch.max_size", 100)

	// Log
	viper.SetDefault("log.level", "info")
	viper.SetDefault("log.format", "json")
	viper.SetDefault("log.output", "stdout")

	// 環境變數綁定
	bindEnvVars()
}

// bindEnvVars 綁定環境變數
func bindEnvVars() {
	// Server
	viper.BindEnv("server.port", "SERVER_PORT")
	viper.BindEnv("server.read_timeout", "SERVER_READ_TIMEOUT")
	viper.BindEnv("server.write_timeout", "SERVER_WRITE_TIMEOUT")
	viper.BindEnv("server.shutdown_timeout", "SERVER_SHUTDOWN_TIMEOUT")

	// MaxMind
	viper.BindEnv("maxmind.db_path", "MAXMIND_DB_PATH")
	viper.BindEnv("maxmind.auto_update", "MAXMIND_AUTO_UPDATE")
	viper.BindEnv("maxmind.update_interval", "MAXMIND_UPDATE_INTERVAL")

	// Redis
	viper.BindEnv("redis.host", "REDIS_HOST")
	viper.BindEnv("redis.port", "REDIS_PORT")
	viper.BindEnv("redis.password", "REDIS_PASSWORD")
	viper.BindEnv("redis.db", "REDIS_DB")
	viper.BindEnv("redis.pool_size", "REDIS_POOL_SIZE")
	viper.BindEnv("redis.min_idle_conns", "REDIS_MIN_IDLE_CONNS")
	viper.BindEnv("redis.max_retries", "REDIS_MAX_RETRIES")
	viper.BindEnv("redis.dial_timeout", "REDIS_DIAL_TIMEOUT")
	viper.BindEnv("redis.read_timeout", "REDIS_READ_TIMEOUT")
	viper.BindEnv("redis.write_timeout", "REDIS_WRITE_TIMEOUT")

	// Cache
	viper.BindEnv("cache.enabled", "CACHE_ENABLED")
	viper.BindEnv("cache.ttl", "CACHE_TTL")
	viper.BindEnv("cache.local_cache_enabled", "LOCAL_CACHE_ENABLED")
	viper.BindEnv("cache.local_cache_size", "LOCAL_CACHE_SIZE")
	viper.BindEnv("cache.local_cache_ttl", "LOCAL_CACHE_TTL")

	// Rate Limit
	viper.BindEnv("rate_limit.enabled", "RATE_LIMIT_ENABLED")
	viper.BindEnv("rate_limit.requests_per_minute", "RATE_LIMIT_RPM")
	viper.BindEnv("rate_limit.requests_per_hour", "RATE_LIMIT_RPH")
	viper.BindEnv("rate_limit.burst", "RATE_LIMIT_BURST")
	viper.BindEnv("rate_limit.storage", "RATE_LIMIT_STORAGE")

	// Batch
	viper.BindEnv("batch.max_size", "BATCH_MAX_SIZE")

	// Log
	viper.BindEnv("log.level", "LOG_LEVEL")
	viper.BindEnv("log.format", "LOG_FORMAT")
	viper.BindEnv("log.output", "LOG_OUTPUT")
}

// Validate 驗證配置
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.MaxMind.DBPath == "" {
		return fmt.Errorf("maxmind db_path is required")
	}

	if c.Batch.MaxSize <= 0 || c.Batch.MaxSize > 1000 {
		return fmt.Errorf("invalid batch max_size: %d (must be 1-1000)", c.Batch.MaxSize)
	}

	return nil
}
