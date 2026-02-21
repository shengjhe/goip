package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/shengjhe/goip/internal/model"
	"github.com/redis/go-redis/v9"
)

const (
	keyPrefix = "goip:country:"
)

// CacheRepository Redis 快取存取介面
type CacheRepository interface {
	Get(ctx context.Context, ip string) (*model.IPInfo, error)
	Set(ctx context.Context, ip string, info *model.IPInfo, ttl time.Duration) error
	MGet(ctx context.Context, ips []string) (map[string]*model.IPInfo, error)
	MSet(ctx context.Context, items map[string]*model.IPInfo, ttl time.Duration) error
	Delete(ctx context.Context, ips ...string) error
	Exists(ctx context.Context, ip string) (bool, error)
	FlushAll(ctx context.Context) error
	GetStats(ctx context.Context) (*model.CacheStats, error)
	Close() error
	HealthCheck(ctx context.Context) error
}

type cacheRepository struct {
	client *redis.Client
}

// NewCacheRepository 建立新的 Cache repository
func NewCacheRepository(client *redis.Client) CacheRepository {
	return &cacheRepository{
		client: client,
	}
}

// Get 獲取單一快取
func (r *cacheRepository) Get(ctx context.Context, ip string) (*model.IPInfo, error) {
	key := keyPrefix + ip
	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var info model.IPInfo
	if err := json.Unmarshal([]byte(val), &info); err != nil {
		return nil, err
	}

	return &info, nil
}

// Set 設定快取
func (r *cacheRepository) Set(ctx context.Context, ip string, info *model.IPInfo, ttl time.Duration) error {
	key := keyPrefix + ip
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}

	return r.client.Set(ctx, key, data, ttl).Err()
}

// MGet 批次獲取多個快取
func (r *cacheRepository) MGet(ctx context.Context, ips []string) (map[string]*model.IPInfo, error) {
	if len(ips) == 0 {
		return make(map[string]*model.IPInfo), nil
	}

	pipe := r.client.Pipeline()

	// 批次查詢
	cmds := make([]*redis.StringCmd, len(ips))
	for i, ip := range ips {
		key := keyPrefix + ip
		cmds[i] = pipe.Get(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	// 忽略 redis.Nil 錯誤（部分鍵不存在是正常的）
	if err != nil && err != redis.Nil {
		return nil, err
	}

	// 解析結果
	results := make(map[string]*model.IPInfo)
	for i, cmd := range cmds {
		val, err := cmd.Result()
		if err == nil {
			var info model.IPInfo
			if json.Unmarshal([]byte(val), &info) == nil {
				results[ips[i]] = &info
			}
		}
	}

	return results, nil
}

// MSet 批次設定多個快取
func (r *cacheRepository) MSet(ctx context.Context, items map[string]*model.IPInfo, ttl time.Duration) error {
	if len(items) == 0 {
		return nil
	}

	pipe := r.client.Pipeline()

	for ip, info := range items {
		key := keyPrefix + ip
		data, err := json.Marshal(info)
		if err != nil {
			continue
		}
		pipe.Set(ctx, key, data, ttl)
	}

	_, err := pipe.Exec(ctx)
	return err
}

// Delete 刪除快取
func (r *cacheRepository) Delete(ctx context.Context, ips ...string) error {
	if len(ips) == 0 {
		return nil
	}

	keys := make([]string, len(ips))
	for i, ip := range ips {
		keys[i] = keyPrefix + ip
	}

	return r.client.Del(ctx, keys...).Err()
}

// Exists 檢查快取是否存在
func (r *cacheRepository) Exists(ctx context.Context, ip string) (bool, error) {
	key := keyPrefix + ip
	count, err := r.client.Exists(ctx, key).Result()
	return count > 0, err
}

// FlushAll 清空所有快取（謹慎使用）
func (r *cacheRepository) FlushAll(ctx context.Context) error {
	// 只刪除符合前綴的鍵
	iter := r.client.Scan(ctx, 0, keyPrefix+"*", 0).Iterator()
	var keys []string

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())
		// 批次刪除，避免一次刪除太多
		if len(keys) >= 1000 {
			if err := r.client.Del(ctx, keys...).Err(); err != nil {
				return err
			}
			keys = keys[:0]
		}
	}

	if err := iter.Err(); err != nil {
		return err
	}

	// 刪除剩餘的鍵
	if len(keys) > 0 {
		return r.client.Del(ctx, keys...).Err()
	}

	return nil
}

// GetStats 獲取快取統計
func (r *cacheRepository) GetStats(ctx context.Context) (*model.CacheStats, error) {
	// 獲取連線池統計
	poolStats := r.client.PoolStats()

	// 獲取 Redis INFO 統計
	info, err := r.client.Info(ctx, "stats", "memory").Result()
	if err != nil {
		return nil, err
	}

	stats := &model.CacheStats{
		PoolHits:     uint64(poolStats.Hits),
		PoolMisses:   uint64(poolStats.Misses),
		PoolTimeouts: uint64(poolStats.Timeouts),
	}

	// 解析 INFO 輸出（簡化版，實際應更詳細解析）
	// 這裡僅為示範，實際可使用更完整的解析邏輯
	fmt.Sscanf(info, "used_memory:%d", &stats.UsedMemory)

	// 獲取鍵數量（符合前綴的）
	iter := r.client.Scan(ctx, 0, keyPrefix+"*", 0).Iterator()
	count := uint64(0)
	for iter.Next(ctx) {
		count++
	}
	stats.KeyCount = count

	return stats, nil
}

// Close 關閉連接
func (r *cacheRepository) Close() error {
	return r.client.Close()
}

// HealthCheck 健康檢查
func (r *cacheRepository) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	return r.client.Ping(ctx).Err()
}
