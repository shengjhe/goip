package repository

import (
	"github.com/shengjhe/goip/internal/model"
)

// GeoIPRepository 統一的 GeoIP 查詢介面
type GeoIPRepository interface {
	// LookupCountry 查詢 IP 的國家和城市資訊
	LookupCountry(ip string) (*model.IPInfo, error)

	// Close 關閉資料庫連接
	Close() error

	// Reload 重新載入資料庫
	Reload(dbPath string) error

	// GetProviderType 取得提供者類型
	GetProviderType() string
}
