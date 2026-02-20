package repository

import (
	"errors"
	"net"
	"sync"

	"github.com/axiom/goip/internal/model"
	"github.com/oschwald/geoip2-golang"
)

var (
	ErrInvalidIP      = errors.New("invalid IP address")
	ErrIPNotFound     = errors.New("IP not found in database")
	ErrDatabaseClosed = errors.New("database is closed")
)

// MaxMindRepository MaxMind DB 存取介面
type MaxMindRepository interface {
	LookupCountry(ip string) (*model.IPInfo, error)
	Close() error
	Reload(dbPath string) error
}

type maxMindRepository struct {
	reader *geoip2.Reader
	dbPath string
	mu     sync.RWMutex
}

// NewMaxMindRepository 建立新的 MaxMind repository
func NewMaxMindRepository(dbPath string) (MaxMindRepository, error) {
	reader, err := geoip2.Open(dbPath)
	if err != nil {
		return nil, err
	}

	return &maxMindRepository{
		reader: reader,
		dbPath: dbPath,
	}, nil
}

// LookupCountry 查詢 IP 的國家和城市資訊
func (r *maxMindRepository) LookupCountry(ipStr string) (*model.IPInfo, error) {
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

	// 查詢城市資訊（包含國家、城市、位置等完整資訊）
	record, err := r.reader.City(ip)
	if err != nil {
		return nil, ErrIPNotFound
	}

	// 組裝回應
	info := &model.IPInfo{
		IP: ipStr,
		Country: model.CountryInfo{
			ISOCode: record.Country.IsoCode,
			Name:    record.Country.Names["en"],
			NameZh:  record.Country.Names["zh-CN"],
		},
		Continent: model.ContinentInfo{
			Code: record.Continent.Code,
			Name: record.Continent.Names["en"],
		},
	}

	// 添加城市資訊（如果有）
	if record.City.Names != nil && len(record.City.Names) > 0 {
		info.City = &model.CityInfo{
			Name:       record.City.Names["en"],
			NameZh:     record.City.Names["zh-CN"],
			PostalCode: record.Postal.Code,
		}
	}

	// 添加位置資訊
	if record.Location.Latitude != 0 || record.Location.Longitude != 0 {
		info.Location = &model.LocationInfo{
			Latitude:  record.Location.Latitude,
			Longitude: record.Location.Longitude,
			TimeZone:  record.Location.TimeZone,
		}
	}

	return info, nil
}

// Close 關閉資料庫連接
func (r *maxMindRepository) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.reader != nil {
		err := r.reader.Close()
		r.reader = nil
		return err
	}
	return nil
}

// Reload 重新載入資料庫（用於熱更新）
func (r *maxMindRepository) Reload(dbPath string) error {
	// 開啟新的資料庫
	newReader, err := geoip2.Open(dbPath)
	if err != nil {
		return err
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// 關閉舊的資料庫
	if r.reader != nil {
		r.reader.Close()
	}

	// 替換為新的資料庫
	r.reader = newReader
	r.dbPath = dbPath

	return nil
}
