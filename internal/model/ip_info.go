package model

import "time"

// IPInfo 表示 IP 查詢結果
type IPInfo struct {
	IP          string         `json:"ip"`                    // 必填：IP 地址
	Country     CountryInfo    `json:"country"`               // 必填：國家資訊
	City        CityInfo       `json:"city"`                  // 必填：城市資訊
	Provider    string         `json:"provider"`              // 必填：資料來源（maxmind, ipip, ip-api, etc.）
	Source      string         `json:"source,omitempty"`      // 資料來源：cache / db / api
	Continent   *ContinentInfo `json:"continent,omitempty"`   // 選填：大洲資訊（只在有資料時顯示）
	Location    *LocationInfo  `json:"location,omitempty"`    // 選填：經緯度資訊（只在有資料時顯示）
	QueryTimeMs int64          `json:"query_time_ms"`         // 查詢耗時
	CachedAt    *time.Time     `json:"cached_at,omitempty"`   // 快取時間
}

// CountryInfo 國家資訊
type CountryInfo struct {
	ISOCode string `json:"iso_code"`
	Name    string `json:"name"`
	NameZh  string `json:"name_zh"`
}

// ContinentInfo 大洲資訊
type ContinentInfo struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

// CityInfo 城市資訊
type CityInfo struct {
	Name       string `json:"name"`        // 城市名稱（英文），可能為空
	NameZh     string `json:"name_zh"`     // 城市名稱（中文），可能為空
	PostalCode string `json:"postal_code"` // 郵遞區號，可能為空
}

// LocationInfo 地理位置資訊
type LocationInfo struct {
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
	TimeZone  string  `json:"time_zone,omitempty"`
}

// BatchResult 批次查詢結果
type BatchResult struct {
	Results []IPInfo `json:"results"`
	Total   int      `json:"total"`
	Success int      `json:"success"`
	Failed  int      `json:"failed"`
}

// BatchRequest 批次查詢請求
type BatchRequest struct {
	IPs []string `json:"ips" binding:"required,min=1,max=100"`
}

// ErrorResponse 錯誤回應
type ErrorResponse struct {
	Error     string    `json:"error"`
	Code      string    `json:"code"`
	Timestamp time.Time `json:"timestamp"`
}

// HealthResponse 健康檢查回應
type HealthResponse struct {
	Status   string            `json:"status"`
	Services map[string]string `json:"services"`
}

// ServiceStats 服務統計資訊
type ServiceStats struct {
	TotalQueries uint64  `json:"total_queries"`
	CacheHits    uint64  `json:"cache_hits"`
	CacheMisses  uint64  `json:"cache_misses"`
	CacheHitRate float64 `json:"cache_hit_rate"`
	AvgQueryTime float64 `json:"avg_query_time_ms"`
	TotalErrors  uint64  `json:"total_errors"`
}

// CacheStats Redis 快取統計
type CacheStats struct {
	PoolHits     uint64  `json:"pool_hits"`
	PoolMisses   uint64  `json:"pool_misses"`
	PoolTimeouts uint64  `json:"pool_timeouts"`
	CacheHits    uint64  `json:"cache_hits"`
	CacheMisses  uint64  `json:"cache_misses"`
	AvgLatency   float64 `json:"avg_latency_ms"`
	UsedMemory   uint64  `json:"used_memory"`
	KeyCount     uint64  `json:"key_count"`
	EvictedKeys  uint64  `json:"evicted_keys"`
}
