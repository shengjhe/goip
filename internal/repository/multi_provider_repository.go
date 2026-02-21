package repository

import (
	"errors"
	"sort"
	"sync"

	"github.com/axiom/goip/internal/model"
)

var (
	ErrNoProviders  = errors.New("no providers configured")
	ErrAllFailed    = errors.New("all providers failed to lookup IP")
	ErrUnknownRegion = errors.New("unknown region type")
)

// ProviderInfo 提供者資訊
type ProviderInfo struct {
	Provider GeoIPRepository
	Priority int
	Region   string // cn, global, all
}

// MultiProviderRepository 多提供者 Repository，支持智能路由
type MultiProviderRepository struct {
	providers []ProviderInfo
	mu        sync.RWMutex
}

// NewMultiProviderRepository 建立新的多提供者 repository
func NewMultiProviderRepository(providers []ProviderInfo) (*MultiProviderRepository, error) {
	if len(providers) == 0 {
		return nil, ErrNoProviders
	}

	// 按優先級排序
	sort.Slice(providers, func(i, j int) bool {
		return providers[i].Priority < providers[j].Priority
	})

	return &MultiProviderRepository{
		providers: providers,
	}, nil
}

// LookupCountry 智能查詢 IP 的國家和城市資訊
// 策略：
// 1. 先用 MaxMind 判斷國家
// 2. 根據國家選擇最佳資料庫
// 3. 如果 city 為空，自動嘗試其他 provider
func (r *MultiProviderRepository) LookupCountry(ipStr string) (*model.IPInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 先用 MaxMind 快速判斷國家（MaxMind 速度快且準確）
	var countryCode string
	var maxmindInfo *model.IPInfo
	maxmindProvider := r.getProviderByType("maxmind")
	if maxmindProvider != nil {
		info, err := maxmindProvider.LookupCountry(ipStr)
		if err == nil && info != nil {
			maxmindInfo = info
			if info.Country.ISOCode != "" {
				countryCode = info.Country.ISOCode
			}
		}
	}

	// 根據國家代碼選擇最佳資料庫
	var primaryInfo *model.IPInfo
	var primaryProvider string

	if countryCode == "CN" {
		// 中國大陸：優先使用 IPIP
		primaryProvider = "ipip"
		primaryInfo = r.tryProvider("ipip", ipStr)
	} else {
		// 其他國家：優先使用 MaxMind
		primaryProvider = "maxmind"
		primaryInfo = maxmindInfo
		if primaryInfo != nil {
			primaryInfo.Provider = "maxmind"
		}
	}

	// 檢查 primary 結果是否有城市資訊
	if primaryInfo != nil && r.hasCityInfo(primaryInfo) {
		return primaryInfo, nil
	}

	// City 為空，嘗試其他所有可用的 providers
	for _, p := range r.providers {
		providerType := p.Provider.GetProviderType()

		// 跳過已經嘗試過的 primary provider
		if providerType == primaryProvider {
			continue
		}

		info := r.tryProvider(providerType, ipStr)
		if info != nil && r.hasCityInfo(info) {
			return info, nil
		}
	}

	// 如果所有 provider 都沒有 city，返回 primary 結果（至少有國家資訊）
	if primaryInfo != nil {
		return primaryInfo, nil
	}

	// 所有提供者都失敗
	return nil, ErrAllFailed
}

// tryProvider 嘗試使用指定的 provider 查詢
func (r *MultiProviderRepository) tryProvider(providerType, ipStr string) *model.IPInfo {
	provider := r.getProviderByType(providerType)
	if provider == nil {
		return nil
	}

	info, err := provider.LookupCountry(ipStr)
	if err == nil && info != nil {
		info.Provider = providerType
		return info
	}
	return nil
}

// hasCityInfo 檢查是否有城市資訊
func (r *MultiProviderRepository) hasCityInfo(info *model.IPInfo) bool {
	return info.City.Name != "" || info.City.NameZh != ""
}

// getProviderByType 根據類型取得提供者
func (r *MultiProviderRepository) getProviderByType(providerType string) GeoIPRepository {
	for _, p := range r.providers {
		if p.Provider.GetProviderType() == providerType {
			return p.Provider
		}
	}
	return nil
}

// Close 關閉所有提供者的資料庫連接
func (r *MultiProviderRepository) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	var errs []error
	for _, p := range r.providers {
		if err := p.Provider.Close(); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// Reload 重新載入所有提供者的資料庫（暫不支持）
func (r *MultiProviderRepository) Reload(dbPath string) error {
	return errors.New("multi-provider reload not supported, reload individual providers instead")
}

// GetProviderType 取得提供者類型
func (r *MultiProviderRepository) GetProviderType() string {
	return "multi-provider"
}

// LookupByProvider 使用指定的提供者查詢 IP
func (r *MultiProviderRepository) LookupByProvider(ipStr, providerType string) (*model.IPInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// 尋找指定的提供者
	for _, p := range r.providers {
		if p.Provider.GetProviderType() == providerType {
			info, err := p.Provider.LookupCountry(ipStr)
			if err == nil && info != nil {
				info.Provider = providerType
			}
			return info, err
		}
	}

	return nil, errors.New("provider not found: " + providerType)
}

// GetProviders 取得所有可用的提供者類型
func (r *MultiProviderRepository) GetProviders() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var providers []string
	seen := make(map[string]bool)

	for _, p := range r.providers {
		providerType := p.Provider.GetProviderType()
		if !seen[providerType] {
			providers = append(providers, providerType)
			seen[providerType] = true
		}
	}

	return providers
}
