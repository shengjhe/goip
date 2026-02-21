# Claude Code 專案開發指南

本文件提供給 Claude Code 的專案背景、架構說明和開發規範。

## 專案概述

**GoIP** 是一個高效能的 IP 地理位置查詢服務，支援多個 IP 資料庫提供者：
- **MaxMind GeoLite2**: 提供全球 IP 地理位置資料，特別是經緯度、時區等詳細資訊
- **IPIP.NET**: 提供準確的中國地區 IP 資料，包含省份和城市資訊

### 核心特色
- 🔄 **智能路由**: 中國 IP 優先使用 IPIP，海外 IP 優先使用 MaxMind
- 🎯 **指定查詢**: 支援強制指定使用特定資料庫
- 📋 **統一格式**: 所有查詢回應使用統一的 JSON 結構
- ⚡ **高效能**: Redis 快取 + 雙層架構
- 🐳 **容器化**: Docker Compose 部署

## 專案結構

```
goip/
├── cmd/server/              # 應用程式入口
├── internal/                # 私有程式碼
│   ├── handler/            # HTTP 處理器
│   ├── service/            # 業務邏輯
│   ├── repository/         # 資料存取層
│   │   ├── geoip_repository.go        # 統一介面
│   │   ├── maxmind_repository.go      # MaxMind 實作
│   │   ├── ipip_repository.go         # IPIP 實作
│   │   └── multi_provider_repository.go # 多資料庫智能路由
│   ├── model/              # 資料模型
│   └── middleware/         # 中間件
├── config/                 # 配置管理
├── data/                   # 資料庫檔案目錄
│   ├── GeoLite2-City.mmdb # MaxMind 資料庫
│   └── ipipfree.ipdb      # IPIP 資料庫
├── build/                  # 建置相關
│   ├── Dockerfile
│   ├── build.sh
│   └── docker-build.sh
├── deployments/            # 部署配置
│   ├── goip/              # GoIP 服務
│   └── redis/             # Redis 服務
├── docs/                   # 額外文件
└── config.yaml            # 服務配置檔
```

## 核心架構

### 1. 多資料庫架構

```go
// 統一介面
type GeoIPRepository interface {
    LookupCountry(ip string) (*model.IPInfo, error)
    Close() error
    Reload(dbPath string) error
    GetProviderType() string
}

// 實作類別
- MaxMindRepository  // 使用 oschwald/geoip2-golang
- IPIPRepository     // 使用 ipipdotnet/ipdb-go
- MultiProviderRepository // 智能路由管理器
```

### 2. 智能路由邏輯

**MultiProviderRepository** ([internal/repository/multi_provider_repository.go](internal/repository/multi_provider_repository.go)) 根據 IP 地址自動選擇最適合的資料庫：

```go
// 中國 IP 範圍優先使用 IPIP
// 其他地區優先使用 MaxMind
func (r *MultiProviderRepository) LookupCountry(ipStr string) (*model.IPInfo, error)

// 強制使用指定提供者
func (r *MultiProviderRepository) LookupByProvider(ipStr, providerType string) (*model.IPInfo, error)
```

### 3. 統一回應格式

所有 API 回應都使用統一的 `IPInfo` 結構 ([internal/model/ip_info.go](internal/model/ip_info.go)):

```json
{
  "ip": "必填 - IP 地址",
  "country": {必填 - 國家資訊},
  "city": {必填 - 城市資訊},
  "provider": "必填 - 資料來源 (maxmind/ipip)",
  "continent": {選填 - 大洲資訊},
  "location": {選填 - 經緯度、時區},
  "query_time_ms": 必填 - 查詢耗時
}
```

**重要**:
- 必填欄位總是存在，即使為空字串
- 選填欄位使用指標類型 (`*ContinentInfo`, `*LocationInfo`)，只在有資料時才出現在 JSON
- 使用 `omitempty` tag 確保 nil 指標不會序列化為空物件

## API 端點

| 端點 | 方法 | 說明 |
|-----|------|------|
| `/api/v1/ip/:ip` | GET | 智能路由查詢 IP |
| `/api/v1/ip/:ip/provider?provider=xxx` | GET | 指定提供者查詢 |
| `/api/v1/providers` | GET | 列出可用的資料庫提供者 |
| `/api/v1/ip/batch` | POST | 批次查詢 |
| `/api/v1/health` | GET | 健康檢查 |
| `/api/v1/stats` | GET | 統計資訊 |
| `/api/v1/cache/invalidate` | POST | 清除快取 |

## 配置說明

### 多資料庫配置 (config.yaml)

```yaml
geoip:
  providers:
    # IPIP.NET - 中國地區優先
    - type: ipip
      db_path: ./data/ipipfree.ipdb
      priority: 1
      region: cn

    # MaxMind - 海外地區優先
    - type: maxmind
      db_path: ./data/GeoLite2-City.mmdb
      priority: 1
      region: global
```

### 環境變數

Docker 部署時可通過環境變數覆蓋配置：
- `SERVER_PORT`: HTTP 端口
- `REDIS_HOST`: Redis 主機
- `CACHE_ENABLED`: 啟用快取
- `LOG_LEVEL`: 日誌級別

## 開發規範

### 1. 檔案管理規則

❌ **禁止**:
- 不要隨便產生 `.sh` 測試腳本
- 不要在專案根目錄建立文件檔案
- 不要建立重複的 `docker-compose.yml`

✅ **正確做法**:
- 額外的 `.md` 文件放在 `docs/` 目錄
- 部署相關檔案放在 `deployments/` 目錄
- 建置相關檔案放在 `build/` 目錄

### 2. 回應格式規範

**必須遵守**:
1. 所有 API 必須返回統一的 `IPInfo` 結構
2. 必填欄位 (`ip`, `country`, `city`, `provider`) 總是存在
3. 選填欄位 (`continent`, `location`) 使用指標類型
4. 只在有資料時才設定選填欄位，避免空物件

**示例**:
```go
// ❌ 錯誤 - 會產生空物件
ipInfo.Continent = model.ContinentInfo{Code: "", Name: ""}

// ✅ 正確 - 只在有資料時設定
if continentCode != "" {
    ipInfo.Continent = &model.ContinentInfo{Code: continentCode}
}
```

### 3. 資料庫提供者實作

新增資料庫提供者時:
1. 實作 `GeoIPRepository` 介面
2. 確保設定所有必填欄位
3. 只在有資料時設定選填欄位
4. 在 `MultiProviderRepository` 註冊新提供者

### 4. 快取處理

- Redis 快取 key 格式: `geoip:ip:{ip_address}`
- 修改資料結構後記得清除快取: `POST /api/v1/cache/invalidate`
- 本地測試時可能需要重啟 Docker 容器以清除快取

## 測試指南

### 本地測試流程

```bash
# 1. 建置 Docker 映像
docker build -t goip:latest -f build/Dockerfile .

# 2. 啟動服務
cd deployments/redis && ./start.sh
cd ../goip && ./start.sh

# 3. 測試 API
curl http://localhost:8080/api/v1/health
curl http://localhost:8080/api/v1/ip/8.8.8.8 | jq .

# 4. 清除快取（如果修改了資料結構）
curl -X POST http://localhost:8080/api/v1/cache/invalidate \
  -H "Content-Type: application/json" \
  -d '{"ips": ["8.8.8.8"]}'
```

### 驗證回應格式

測試不同資料庫的回應格式一致性：

```bash
# IPIP 查詢（中國 IP）
curl http://localhost:8080/api/v1/ip/114.114.114.114 | jq .

# MaxMind 查詢（海外 IP）
curl http://localhost:8080/api/v1/ip/8.8.8.8 | jq .

# 指定提供者
curl "http://localhost:8080/api/v1/ip/8.8.8.8/provider?provider=maxmind" | jq .
```

## 資料庫維護

### MaxMind GeoLite2
- **檔案**: `data/GeoLite2-City.mmdb`
- **更新頻率**: 每週二
- **建議**: 每月更新一次
- **下載**: https://dev.maxmind.com/geoip/geolite2-free-geolocation-data

### IPIP.NET
- **檔案**: `data/ipipfree.ipdb`
- **免費版**: 提供基本的國家、省份、城市資訊
- **付費版**: 提供經緯度、ISP 等額外資訊

## 常見問題

### Q1: 修改了 model 結構但回應格式沒變？
**A**: 清除 Redis 快取或重啟 Docker 容器

### Q2: 為什麼 IPIP 沒有經緯度資訊？
**A**: IPIP 免費版不包含經緯度，需要使用付費版或改用 MaxMind

### Q3: 如何測試特定資料庫？
**A**: 使用 `/api/v1/ip/:ip/provider?provider=xxx` 端點

### Q4: Docker build 失敗？
**A**: 確認 `data/` 目錄存在且包含資料庫檔案

## 相關文件

- [README.md](README.md) - 使用說明
- [docs/DESIGN.md](docs/DESIGN.md) - 詳細架構設計
- [docs/MULTI_DB_GUIDE.md](docs/MULTI_DB_GUIDE.md) - 多資料庫使用指南
- [docs/API_RESPONSE_FORMAT.md](docs/API_RESPONSE_FORMAT.md) - API 回應格式詳細說明
- [docs/RESPONSE_FORMAT_SUMMARY.md](docs/RESPONSE_FORMAT_SUMMARY.md) - 統一格式實作總結

## 開發歷史重點

1. **初始版本**: 單一 MaxMind 資料庫支援
2. **多資料庫支援**: 新增 IPIP.NET 支援，實作智能路由
3. **統一回應格式**: 修正 IPIP 和 MaxMind 回應格式不一致問題
4. **快取優化**: 實作 Redis 快取和清除機制
5. **容器化部署**: Docker Compose 多服務部署

## 開發注意事項

- ⚠️ 不要過度建立文件，優先更新 README
- ⚠️ 測試前記得清除快取
- ⚠️ 修改 model 結構時注意必填/選填欄位的處理
- ⚠️ 新增功能前先檢查是否與現有架構一致
