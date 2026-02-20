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

// LookupCountry 查詢 IP 的國家資訊
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

	// 查詢國家資訊
	record, err := r.reader.Country(ip)
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
