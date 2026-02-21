package repository

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/shengjhe/goip/internal/model"
)

// ExternalAPIType 外部 API 類型
type ExternalAPIType string

const (
	ExternalAPIIPAPI   ExternalAPIType = "ip-api"
	ExternalAPIIPInfo  ExternalAPIType = "ipinfo"
	ExternalAPIIPAPIco ExternalAPIType = "ipapi.co"
)

// ExternalAPIRepository 外部 IP API 查詢 repository
type ExternalAPIRepository struct {
	apiType    ExternalAPIType
	httpClient *http.Client
	mu         sync.RWMutex
}

// NewExternalAPIRepository 建立新的外部 API repository
func NewExternalAPIRepository(apiType ExternalAPIType) (*ExternalAPIRepository, error) {
	return &ExternalAPIRepository{
		apiType: apiType,
		httpClient: &http.Client{
			Timeout: 5 * time.Second,
		},
	}, nil
}

// LookupCountry 查詢 IP 的國家和城市資訊
func (r *ExternalAPIRepository) LookupCountry(ipStr string) (*model.IPInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	switch r.apiType {
	case ExternalAPIIPAPI:
		return r.queryIPAPI(ipStr)
	case ExternalAPIIPInfo:
		return r.queryIPInfo(ipStr)
	case ExternalAPIIPAPIco:
		return r.queryIPAPIco(ipStr)
	default:
		return nil, fmt.Errorf("unknown external API type: %s", r.apiType)
	}
}

// queryIPAPI 查詢 ip-api.com
func (r *ExternalAPIRepository) queryIPAPI(ipStr string) (*model.IPInfo, error) {
	url := fmt.Sprintf("http://ip-api.com/json/%s", ipStr)
	resp, err := r.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("ip-api request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp struct {
		Status      string  `json:"status"`
		Country     string  `json:"country"`
		CountryCode string  `json:"countryCode"`
		Region      string  `json:"region"`
		RegionName  string  `json:"regionName"`
		City        string  `json:"city"`
		Zip         string  `json:"zip"`
		Lat         float64 `json:"lat"`
		Lon         float64 `json:"lon"`
		Timezone    string  `json:"timezone"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	if apiResp.Status != "success" {
		return nil, fmt.Errorf("ip-api query failed")
	}

	ipInfo := &model.IPInfo{
		IP: ipStr,
		Country: model.CountryInfo{
			ISOCode: apiResp.CountryCode,
			Name:    apiResp.Country,
		},
		City: model.CityInfo{
			Name:       apiResp.City,
			PostalCode: apiResp.Zip,
		},
		Provider: string(r.apiType),
	}

	// 位置資訊
	if apiResp.Lat != 0 || apiResp.Lon != 0 {
		ipInfo.Location = &model.LocationInfo{
			Latitude:  apiResp.Lat,
			Longitude: apiResp.Lon,
			TimeZone:  apiResp.Timezone,
		}
	}

	return ipInfo, nil
}

// queryIPInfo 查詢 ipinfo.io
func (r *ExternalAPIRepository) queryIPInfo(ipStr string) (*model.IPInfo, error) {
	url := fmt.Sprintf("https://ipinfo.io/%s/json", ipStr)
	resp, err := r.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("ipinfo.io request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp struct {
		IP       string `json:"ip"`
		City     string `json:"city"`
		Region   string `json:"region"`
		Country  string `json:"country"`
		Loc      string `json:"loc"`
		Postal   string `json:"postal"`
		Timezone string `json:"timezone"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	ipInfo := &model.IPInfo{
		IP: ipStr,
		Country: model.CountryInfo{
			ISOCode: apiResp.Country,
		},
		City: model.CityInfo{
			Name:       apiResp.City,
			PostalCode: apiResp.Postal,
		},
		Provider: string(r.apiType),
	}

	// 解析經緯度
	if apiResp.Loc != "" {
		parts := strings.Split(apiResp.Loc, ",")
		if len(parts) == 2 {
			lat, _ := strconv.ParseFloat(parts[0], 64)
			lon, _ := strconv.ParseFloat(parts[1], 64)
			if lat != 0 || lon != 0 {
				ipInfo.Location = &model.LocationInfo{
					Latitude:  lat,
					Longitude: lon,
					TimeZone:  apiResp.Timezone,
				}
			}
		}
	}

	return ipInfo, nil
}

// queryIPAPIco 查詢 ipapi.co
func (r *ExternalAPIRepository) queryIPAPIco(ipStr string) (*model.IPInfo, error) {
	url := fmt.Sprintf("https://ipapi.co/%s/json/", ipStr)
	resp, err := r.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("ipapi.co request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp struct {
		IP            string  `json:"ip"`
		City          string  `json:"city"`
		Region        string  `json:"region"`
		RegionCode    string  `json:"region_code"`
		Country       string  `json:"country"`
		CountryName   string  `json:"country_name"`
		CountryCode   string  `json:"country_code"`
		ContinentCode string  `json:"continent_code"`
		Postal        string  `json:"postal"`
		Latitude      float64 `json:"latitude"`
		Longitude     float64 `json:"longitude"`
		Timezone      string  `json:"timezone"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	ipInfo := &model.IPInfo{
		IP: ipStr,
		Country: model.CountryInfo{
			ISOCode: apiResp.CountryCode,
			Name:    apiResp.CountryName,
		},
		City: model.CityInfo{
			Name:       apiResp.City,
			PostalCode: apiResp.Postal,
		},
		Provider: string(r.apiType),
	}

	// 大洲資訊
	if apiResp.ContinentCode != "" {
		ipInfo.Continent = &model.ContinentInfo{
			Code: apiResp.ContinentCode,
		}
	}

	// 位置資訊
	if apiResp.Latitude != 0 || apiResp.Longitude != 0 {
		ipInfo.Location = &model.LocationInfo{
			Latitude:  apiResp.Latitude,
			Longitude: apiResp.Longitude,
			TimeZone:  apiResp.Timezone,
		}
	}

	return ipInfo, nil
}

// Close 關閉連接
func (r *ExternalAPIRepository) Close() error {
	// HTTP client 不需要明確關閉
	return nil
}

// Reload 重新載入（外部 API 不需要）
func (r *ExternalAPIRepository) Reload(dbPath string) error {
	return nil
}

// GetProviderType 取得提供者類型
func (r *ExternalAPIRepository) GetProviderType() string {
	return string(r.apiType)
}
