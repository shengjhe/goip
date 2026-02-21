package repository

import (
	"errors"
	"net"
	"sync"

	"github.com/shengjhe/goip/internal/model"
	"github.com/ipipdotnet/ipdb-go"
)

var (
	ErrIPIPNotFound = errors.New("IP not found in IPIP database")
)

// IPIPRepository IPIP.NET DB 存取介面
type IPIPRepository interface {
	GeoIPRepository
}

type ipipRepository struct {
	reader       *ipdb.City
	dbPath       string
	providerType string
	mu           sync.RWMutex
}

// NewIPIPRepository 建立新的 IPIP repository
func NewIPIPRepository(dbPath string) (IPIPRepository, error) {
	reader, err := ipdb.NewCity(dbPath)
	if err != nil {
		return nil, err
	}

	return &ipipRepository{
		reader:       reader,
		dbPath:       dbPath,
		providerType: "ipip",
	}, nil
}

// LookupCountry 查詢 IP 的國家和城市資訊
func (r *ipipRepository) LookupCountry(ipStr string) (*model.IPInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.reader == nil {
		return nil, ErrDatabaseClosed
	}

	// 解析 IP 地址
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, ErrInvalidIP
	}

	// 查詢資訊
	info, err := r.reader.FindMap(ipStr, "CN")
	if err != nil {
		return nil, ErrIPIPNotFound
	}

	// 組裝回應 - 確保所有必填欄位都有值
	ipInfo := &model.IPInfo{
		IP:       ipStr,
		Country:  model.CountryInfo{},
		City:     model.CityInfo{},
		Provider: "", // 會在 MultiProvider 中設定
	}

	// IPIP.NET 免費版的資料結構：
	// 國家、省份、城市、組織/運營商
	if countryName, ok := info["country_name"]; ok && countryName != "" {
		ipInfo.Country.Name = countryName
	}
	if countryCode, ok := info["country_code"]; ok && countryCode != "" {
		ipInfo.Country.ISOCode = countryCode
	}

	// 省份資訊（如果有）
	regionName, hasRegion := info["region_name"]

	// 城市資訊
	if cityName, ok := info["city_name"]; ok && cityName != "" {
		ipInfo.City.Name = cityName
		// 將省份資訊加入中文名稱（IPIP 沒有獨立的省份欄位）
		if hasRegion && regionName != "" {
			ipInfo.City.NameZh = regionName + cityName
		}
	}

	// 大洲資訊（IPIP 免費版通常沒有）
	if continentCode, ok := info["continent_code"]; ok && continentCode != "" {
		ipInfo.Continent = &model.ContinentInfo{
			Code: continentCode,
		}
	}

	// 位置資訊（如果有經緯度）
	// IPIP 免費版通常不包含經緯度
	if lat, ok := info["latitude"]; ok && lat != "" {
		if ipInfo.Location == nil {
			ipInfo.Location = &model.LocationInfo{}
		}
		// IPIP 返回的是字串，需要轉換為 float64
		// 免費版可能不包含經緯度，這裡預留處理
	}
	if lng, ok := info["longitude"]; ok && lng != "" {
		if ipInfo.Location == nil {
			ipInfo.Location = &model.LocationInfo{}
		}
	}

	return ipInfo, nil
}

// Close 關閉資料庫連接
func (r *ipipRepository) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// IPIP SDK 不需要明確關閉
	r.reader = nil
	return nil
}

// Reload 重新載入資料庫（用於熱更新）
func (r *ipipRepository) Reload(dbPath string) error {
	// 開啟新的資料庫
	newReader, err := ipdb.NewCity(dbPath)
	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// 替換為新的資料庫
	r.reader = newReader
	r.dbPath = dbPath

	return nil
}

// GetProviderType 取得提供者類型
func (r *ipipRepository) GetProviderType() string {
	return r.providerType
}
