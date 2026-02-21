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
// 策略：先用 MaxMind 快速判斷國家，再根據國家選擇最佳資料庫
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
	if countryCode == "CN" {
		// 中國大陸：優先使用 IPIP（中文城市資訊更詳細）
		ipipProvider := r.getProviderByType("ipip")
		if ipipProvider != nil {
			info, err := ipipProvider.LookupCountry(ipStr)
			if err == nil && info != nil {
				info.Provider = "ipip"
				return info, nil
			}
		}
		// IPIP 失敗，回退到 MaxMind
		if maxmindInfo != nil {
			maxmindInfo.Provider = "maxmind"
			return maxmindInfo, nil
		}
	} else {
		// 其他國家（包含台港澳）：直接使用 MaxMind
		if maxmindInfo != nil {
			maxmindInfo.Provider = "maxmind"
			return maxmindInfo, nil
		}
		// MaxMind 沒有結果，嘗試 IPIP
		ipipProvider := r.getProviderByType("ipip")
		if ipipProvider != nil {
			info, err := ipipProvider.LookupCountry(ipStr)
			if err == nil && info != nil {
				info.Provider = "ipip"
				return info, nil
			}
		}
	}

	// 所有提供者都失敗
	return nil, ErrAllFailed
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

// getProvidersForRegion 取得指定區域的提供者
func (r *MultiProviderRepository) getProvidersForRegion(region string) []ProviderInfo {
	var result []ProviderInfo
	for _, p := range r.providers {
		if p.Region == region || p.Region == "all" || p.Region == "" {
			result = append(result, p)
		}
	}
	return result
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
