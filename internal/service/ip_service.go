package service

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/axiom/goip/internal/model"
	"github.com/axiom/goip/internal/repository"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

// IPService IP 查詢服務介面
type IPService interface {
	LookupIP(ctx context.Context, ip string) (*model.IPInfo, error)
	LookupIPByProvider(ctx context.Context, ip string, provider string) (*model.IPInfo, error)
	BatchLookup(ctx context.Context, ips []string) (*model.BatchResult, error)
	GetStats() *model.ServiceStats
	InvalidateCache(ctx context.Context, ips ...string) error
	GetAvailableProviders() []string
}

type ipService struct {
	geoip    repository.GeoIPRepository
	cache    repository.CacheRepository
	logger   zerolog.Logger
	cacheTTL time.Duration

	// 統計資料
	stats struct {
		totalQueries uint64
		cacheHits    uint64
		cacheMisses  uint64
		totalErrors  uint64
		totalTime    uint64 // 累計查詢時間（微秒）
		queryCount   uint64 // 用於計算平均時間
	}
}

// NewIPService 建立新的 IP Service
func NewIPService(
	geoip repository.GeoIPRepository,
	cache repository.CacheRepository,
	logger zerolog.Logger,
	cacheTTL time.Duration,
) IPService {
	return &ipService{
		geoip:    geoip,
		cache:    cache,
		logger:   logger,
		cacheTTL: cacheTTL,
	}
}

// LookupIP 查詢單一 IP（Cache-Aside Pattern）
func (s *ipService) LookupIP(ctx context.Context, ip string) (*model.IPInfo, error) {
	startTime := time.Now()
	atomic.AddUint64(&s.stats.totalQueries, 1)

	// 1. 嘗試從 Redis 快取讀取
	result, err := s.cache.Get(ctx, ip)
	if err == nil {
		atomic.AddUint64(&s.stats.cacheHits, 1)
		s.recordQueryTime(startTime)
		// 標記資料來源為 cache
		result.Source = "cache"
		return result, nil
	}

	// 2. Redis 錯誤時記錄但不中斷服務
	if err != redis.Nil {
		s.logger.Warn().Err(err).Str("ip", ip).Msg("Redis cache error, fallback to DB")
	}
	atomic.AddUint64(&s.stats.cacheMisses, 1)

	// 3. 查詢 GeoIP (DB or API)
	result, err = s.geoip.LookupCountry(ip)
	if err != nil {
		atomic.AddUint64(&s.stats.totalErrors, 1)
		return nil, err
	}

	// 標記資料來源：根據 provider 判斷是 db 還是 api
	switch result.Provider {
	case "ip-api", "ipinfo", "ipapi.co":
		result.Source = "api"
	default:
		result.Source = "db"
	}

	// 記錄查詢時間
	queryTime := time.Since(startTime)
	result.QueryTimeMs = queryTime.Milliseconds()

	// 4. 嘗試寫入快取（失敗不影響回應）
	if cacheErr := s.cache.Set(ctx, ip, result, s.cacheTTL); cacheErr != nil {
		s.logger.Warn().Err(cacheErr).Str("ip", ip).Msg("Failed to cache result")
	}

	s.recordQueryTime(startTime)
	return result, nil
}

// BatchLookup 批次查詢（優化快取存取）
func (s *ipService) BatchLookup(ctx context.Context, ips []string) (*model.BatchResult, error) {
	if len(ips) == 0 {
		return &model.BatchResult{
			Results: []model.IPInfo{},
			Total:   0,
			Success: 0,
			Failed:  0,
		}, nil
	}

	atomic.AddUint64(&s.stats.totalQueries, uint64(len(ips)))

	// 1. 使用 Redis MGET 批次查詢快取
	cachedResults, err := s.cache.MGet(ctx, ips)
	if err != nil {
		s.logger.Warn().Err(err).Msg("Batch cache lookup failed")
		cachedResults = make(map[string]*model.IPInfo)
	}

	// 標記快取結果來源
	for _, info := range cachedResults {
		info.Source = "cache"
	}

	// 2. 收集未命中的 IP
	var missedIPs []string
	for _, ip := range ips {
		if _, found := cachedResults[ip]; !found {
			missedIPs = append(missedIPs, ip)
		}
	}

	atomic.AddUint64(&s.stats.cacheHits, uint64(len(cachedResults)))
	atomic.AddUint64(&s.stats.cacheMisses, uint64(len(missedIPs)))

	// 3. 並行查詢 MaxMind DB（未命中的 IP）
	dbResults := s.parallelLookup(ctx, missedIPs)

	// 標記 DB/API 結果來源
	for _, info := range dbResults {
		switch info.Provider {
		case "ip-api", "ipinfo", "ipapi.co":
			info.Source = "api"
		default:
			info.Source = "db"
		}
	}

	// 4. 批次寫入快取
	if len(dbResults) > 0 {
		if err := s.cache.MSet(ctx, dbResults, s.cacheTTL); err != nil {
			s.logger.Warn().Err(err).Msg("Batch cache write failed")
		}
	}

	// 5. 合併結果
	allResults := make([]model.IPInfo, 0, len(ips))
	successCount := 0
	failedCount := 0

	for _, ip := range ips {
		// 先從快取找
		if info, found := cachedResults[ip]; found {
			allResults = append(allResults, *info)
			successCount++
			continue
		}

		// 再從 DB 結果找
		if info, found := dbResults[ip]; found {
			allResults = append(allResults, *info)
			successCount++
		} else {
			failedCount++
		}
	}

	return &model.BatchResult{
		Results: allResults,
		Total:   len(ips),
		Success: successCount,
		Failed:  failedCount,
	}, nil
}

// parallelLookup 並行查詢 MaxMind DB
func (s *ipService) parallelLookup(ctx context.Context, ips []string) map[string]*model.IPInfo {
	results := make(map[string]*model.IPInfo)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// 限制並行數量
	semaphore := make(chan struct{}, 10)

	for _, ip := range ips {
		wg.Add(1)
		go func(ipAddr string) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			info, err := s.geoip.LookupCountry(ipAddr)
			if err != nil {
				atomic.AddUint64(&s.stats.totalErrors, 1)
				s.logger.Debug().Err(err).Str("ip", ipAddr).Msg("Failed to lookup IP")
				return
			}

			mu.Lock()
			results[ipAddr] = info
			mu.Unlock()
		}(ip)
	}

	wg.Wait()
	return results
}

// GetStats 獲取服務統計
func (s *ipService) GetStats() *model.ServiceStats {
	totalQueries := atomic.LoadUint64(&s.stats.totalQueries)
	cacheHits := atomic.LoadUint64(&s.stats.cacheHits)
	cacheMisses := atomic.LoadUint64(&s.stats.cacheMisses)
	totalErrors := atomic.LoadUint64(&s.stats.totalErrors)
	totalTime := atomic.LoadUint64(&s.stats.totalTime)
	queryCount := atomic.LoadUint64(&s.stats.queryCount)

	hitRate := 0.0
	if totalQueries > 0 {
		hitRate = float64(cacheHits) / float64(totalQueries) * 100
	}

	avgQueryTime := 0.0
	if queryCount > 0 {
		avgQueryTime = float64(totalTime) / float64(queryCount) / 1000.0 // 轉換為毫秒
	}

	return &model.ServiceStats{
		TotalQueries: totalQueries,
		CacheHits:    cacheHits,
		CacheMisses:  cacheMisses,
		CacheHitRate: hitRate,
		AvgQueryTime: avgQueryTime,
		TotalErrors:  totalErrors,
	}
}

// InvalidateCache 清除指定 IP 的快取
func (s *ipService) InvalidateCache(ctx context.Context, ips ...string) error {
	return s.cache.Delete(ctx, ips...)
}

// LookupIPByProvider 使用指定的提供者查詢 IP
func (s *ipService) LookupIPByProvider(ctx context.Context, ip string, provider string) (*model.IPInfo, error) {
	startTime := time.Now()
	atomic.AddUint64(&s.stats.totalQueries, 1)

	// 檢查是否為 MultiProvider
	if multiRepo, ok := s.geoip.(*repository.MultiProviderRepository); ok {
		result, err := multiRepo.LookupByProvider(ip, provider)
		if err != nil {
			atomic.AddUint64(&s.stats.totalErrors, 1)
			s.logger.Error().Err(err).Str("ip", ip).Str("provider", provider).Msg("Provider lookup failed")
			return nil, err
		}

		// 標記資料來源：指定 provider 時直接查詢，不使用快取
		switch provider {
		case "ip-api", "ipinfo", "ipapi.co":
			result.Source = "api"
		default:
			result.Source = "db"
		}

		// 記錄查詢時間
		result.QueryTimeMs = time.Since(startTime).Milliseconds()
		s.recordQueryTime(startTime)

		return result, nil
	}

	// 如果不是 MultiProvider，使用一般查詢
	return s.LookupIP(ctx, ip)
}

// GetAvailableProviders 取得所有可用的提供者
func (s *ipService) GetAvailableProviders() []string {
	if multiRepo, ok := s.geoip.(*repository.MultiProviderRepository); ok {
		return multiRepo.GetProviders()
	}

	// 如果不是 MultiProvider，返回單一提供者
	return []string{s.geoip.GetProviderType()}
}

// recordQueryTime 記錄查詢時間
func (s *ipService) recordQueryTime(startTime time.Time) {
	duration := time.Since(startTime).Microseconds()
	atomic.AddUint64(&s.stats.totalTime, uint64(duration))
	atomic.AddUint64(&s.stats.queryCount, 1)
}
