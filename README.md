# GoIP - IP 地理位置查詢服務

支援多資料庫的高效能 IP 地理位置查詢 RESTful API 服務。

## 特色

- 🚀 **高效能**: Redis 分散式快取 + 本地快取雙層架構
- 🌐 **多資料庫支援**: 整合 MaxMind GeoLite2、IPIP.NET 及外部 API
- 🌍 **外部 API 整合**: 支援 ip-api.com、ipinfo.io、ipapi.co
- 🎯 **智能路由**: 中國大陸 IP 使用 IPIP，其他地區使用 MaxMind
- 🔄 **智能 Fallback**: 本地資料庫無城市資訊時自動切換至其他 provider
- 🏙️ **詳細資訊**: 支援國家、城市、郵遞區號、經緯度、時區等完整地理資訊
- 🔒 **限流保護**: Redis 實現的分散式限流
- 📊 **批次查詢**: 支援批次 IP 查詢，使用 Pipeline 優化
- 🗑️ **緩存管理**: 支援啟動時自動清空 DNS 緩存
- 🐳 **容器化**: Docker Compose 一鍵部署
- 📈 **可監控**: 支援健康檢查和統計 API

## 技術棧

- **語言**: Go 1.26+
- **Web 框架**: Gin
- **快取**: Redis 7+
- **IP 資料庫**:
  - MaxMind GeoLite2 City (全球覆蓋，含經緯度)
  - IPIP.NET 免費版 (中國地區詳細城市資訊)
  - 外部 API（選用）: ip-api.com, ipinfo.io, ipapi.co
- **日誌**: Zerolog
- **配置**: Viper

## 快速開始

### 前置需求

- Docker & Docker Compose（推薦）
- 或 Go 1.26+ & Redis 7.0+
- IP 資料庫檔案（至少一個）：
  - MaxMind GeoLite2 City
  - IPIP.NET 免費版

### 安裝

1. Clone 專案
```bash
git clone https://github.com/axiom/goip.git
cd goip
```

2. 下載 IP 資料庫

**MaxMind GeoLite2 City**（必需）
- 註冊 [MaxMind](https://www.maxmind.com/en/geolite2/signup) 帳號
- 下載 GeoLite2-City.mmdb
- 放置到 `data/GeoLite2-City.mmdb`
- **更新頻率**: 每週二更新

**IPIP.NET 免費版**（選用，提供中國地區詳細城市資訊）
- 下載 [ipipfree.ipdb](https://www.ipip.net/product/client.html)
- 放置到 `data/ipipfree.ipdb`

3. 配置多資料庫（可選）
```bash
cp config.yaml.example config.yaml
# 編輯 config.yaml 設定資料庫路徑
```

### 使用 Docker Compose 部署（推薦）

**方式一：使用 Makefile**
```bash
# 一鍵建置並部署
make full-deploy

# 或分步執行
make docker-build      # 建置 GoIP Docker 映像
make docker-up         # 啟動所有服務（Redis + GoIP）
```

**方式二：使用啟動腳本**
```bash
# 1. 建置 Docker 映像
./build/docker-build.sh

# 2. 啟動 Redis
cd deployments/redis
./start.sh

# 3. 啟動 GoIP
cd ../goip
./start.sh
```

**方式三：直接使用 docker-compose**
```bash
# 啟動 Redis
docker-compose -f deployments/redis/docker-compose.yml up -d

# 啟動 GoIP
docker-compose -f deployments/goip/docker-compose.yml up -d
```

服務將在 `http://localhost:8080` 啟動。

### 本地開發運行

```bash
# 安裝依賴
go mod download

# 啟動 Redis
make docker-redis-up
# 或
cd deployments/redis && ./start.sh

# 運行服務
make run
# 或
go run cmd/server/main.go
```

## API 文檔

### 智能路由查詢

```bash
GET /api/v1/ip/{ip}
```

系統會自動選擇最佳資料庫：
- **中國大陸 IP** → 使用 IPIP（中文城市資訊詳細）
- **其他國家** → 使用 MaxMind（全球覆蓋，含經緯度）

**範例 1：海外 IP (使用 MaxMind)**
```bash
curl http://localhost:8080/api/v1/ip/8.8.8.8
```

```json
{
  "ip": "8.8.8.8",
  "country": {
    "iso_code": "US",
    "name": "United States",
    "name_zh": "美国"
  },
  "city": {
    "name": "",
    "name_zh": "",
    "postal_code": ""
  },
  "provider": "maxmind",
  "continent": {
    "code": "NA",
    "name": "North America"
  },
  "location": {
    "latitude": 37.751,
    "longitude": -97.822,
    "time_zone": "America/Chicago"
  },
  "query_time_ms": 1
}
```

**範例 2：中國 IP (使用 IPIP)**
```bash
curl http://localhost:8080/api/v1/ip/42.120.160.1
```

```json
{
  "ip": "42.120.160.1",
  "country": {
    "iso_code": "",
    "name": "中国",
    "name_zh": ""
  },
  "city": {
    "name": "杭州",
    "name_zh": "浙江杭州",
    "postal_code": ""
  },
  "provider": "ipip",
  "query_time_ms": 1
}
```

### 指定資料庫查詢

```bash
GET /api/v1/ip/{ip}/provider?provider={maxmind|ipip}
```

強制使用特定資料庫進行查詢。

**範例：使用 MaxMind 查詢中國 IP（獲取經緯度）**
```bash
curl "http://localhost:8080/api/v1/ip/42.120.160.1/provider?provider=maxmind"
```

```json
{
  "ip": "42.120.160.1",
  "country": {
    "iso_code": "CN",
    "name": "China",
    "name_zh": "中国"
  },
  "city": {
    "name": "Hangzhou",
    "name_zh": "杭州",
    "postal_code": ""
  },
  "provider": "maxmind",
  "continent": {
    "code": "AS",
    "name": "Asia"
  },
  "location": {
    "latitude": 30.2943,
    "longitude": 120.1663,
    "time_zone": "Asia/Shanghai"
  },
  "query_time_ms": 2
}
```

### 列出可用資料庫

```bash
GET /api/v1/providers
```

**回應:**
```json
{
  "count": 2,
  "providers": ["ipip", "maxmind"]
}
```

### 回應格式說明

**必填欄位**（總是存在）：
- `ip` - IP 地址
- `country` - 國家資訊
- `city` - 城市資訊
- `provider` - 資料來源（`maxmind` 或 `ipip`）
- `query_time_ms` - 查詢耗時

**選填欄位**（只在有資料時出現）：
- `continent` - 大洲資訊
- `location` - 經緯度和時區

### 批次查詢

```bash
POST /api/v1/ip/batch
Content-Type: application/json
```

**範例請求:**
```bash
curl -X POST http://localhost:8080/api/v1/ip/batch \
  -H "Content-Type: application/json" \
  -d '{"ips": ["8.8.8.8", "114.114.114.114", "42.120.160.1"]}'
```

**範例回應:**
```json
{
  "results": [
    {
      "ip": "8.8.8.8",
      "country": {
        "iso_code": "US",
        "name": "United States",
        "name_zh": "美国"
      },
      "city": {
        "name": "",
        "name_zh": "",
        "postal_code": ""
      },
      "provider": "maxmind",
      "continent": {
        "code": "NA",
        "name": "North America"
      },
      "location": {
        "latitude": 37.751,
        "longitude": -97.822,
        "time_zone": "America/Chicago"
      },
      "query_time_ms": 1
    },
    {
      "ip": "114.114.114.114",
      "country": {
        "iso_code": "",
        "name": "114DNS.COM",
        "name_zh": ""
      },
      "city": {
        "name": "",
        "name_zh": "",
        "postal_code": ""
      },
      "provider": "ipip",
      "query_time_ms": 1
    },
    {
      "ip": "42.120.160.1",
      "country": {
        "iso_code": "",
        "name": "中国",
        "name_zh": ""
      },
      "city": {
        "name": "杭州",
        "name_zh": "浙江杭州",
        "postal_code": ""
      },
      "provider": "ipip",
      "query_time_ms": 1
    }
  ],
  "total": 3,
  "success": 3,
  "failed": 0
}
```

### 健康檢查

```bash
GET /api/v1/health
```

**範例回應:**
```json
{
  "status": "healthy",
  "services": {
    "maxmind": "healthy",
    "redis": "healthy"
  }
}
```

### 統計資訊

```bash
GET /api/v1/stats
```

## 配置說明

服務支援使用 YAML 配置檔或環境變數進行配置。

### 配置檔案 (config.yaml)

在專案根目錄建立 `config.yaml` 檔案：

```yaml
# 服務設定
server:
  port: 8080                  # HTTP 服務端口
  read_timeout: 10s           # 讀取超時
  write_timeout: 10s          # 寫入超時
  shutdown_timeout: 30s       # 優雅關閉超時

# 多提供者 GeoIP 配置（推薦）
geoip:
  providers:
    # IPIP.NET - 中國地區優先
    - type: ipip
      db_path: ./data/ipipfree.ipdb
      priority: 1
      region: cn              # 適用於中國地區

    # MaxMind - 海外地區優先
    - type: maxmind
      db_path: ./data/GeoLite2-City.mmdb
      priority: 1
      region: global          # 適用於海外地區

# 向後相容：單一 MaxMind 資料庫配置
# 如果 geoip.providers 未設定，則使用此配置
# maxmind:
#   db_path: ./data/GeoLite2-City.mmdb
#   auto_update: false
#   update_interval: 24h

# Redis 配置
redis:
  host: localhost
  port: 6379
  password: ""                # Redis 密碼（選用）
  db: 0                       # 資料庫編號
  pool_size: 10               # 連接池大小
  min_idle_conns: 5           # 最小閒置連接數
  max_retries: 3              # 最大重試次數
  dial_timeout: 5s            # 連線超時
  read_timeout: 3s            # 讀取超時
  write_timeout: 3s           # 寫入超時

# 快取配置
cache:
  enabled: true               # 啟用快取
  ttl: 24h                    # 快取過期時間
  local_cache_enabled: false  # 啟用本地快取
  local_cache_size: 1000      # 本地快取大小
  local_cache_ttl: 5m         # 本地快取過期時間

# 限流配置
rate_limit:
  enabled: true               # 啟用限流
  requests_per_minute: 100    # 每分鐘請求限制
  requests_per_hour: 5000     # 每小時請求限制
  burst: 10                   # 突發流量上限
  storage: redis              # 儲存方式 (redis 或 memory)

# 批次查詢配置
batch:
  max_size: 100               # 批次查詢最大數量

# 日誌配置
log:
  level: info                 # 日誌級別 (debug/info/warn/error)
  format: json                # 日誌格式 (json 或 console)
  output: stdout              # 輸出位置 (stdout 或檔案路徑)
```

### 環境變數

也可以使用環境變數覆蓋配置（詳見 `.env.example`）：

| 變數名稱 | 預設值 | 說明 |
|---------|--------|------|
| SERVER_PORT | 8080 | HTTP 服務端口 |
| REDIS_HOST | redis | Redis 主機位址 |
| REDIS_PORT | 6379 | Redis 端口 |
| MAXMIND_DB_PATH | ./data/GeoLite2-City.mmdb | MaxMind 資料庫路徑（向後相容） |
| CACHE_TTL | 24h | 快取過期時間 |
| RATE_LIMIT_RPM | 100 | 每分鐘請求限制 |
| LOG_LEVEL | info | 日誌級別 |
| FLUSH_DNS | false | 啟動時清空 DNS 緩存（true/false） |

完整架構設計請參考 [docs/DESIGN.md](docs/DESIGN.md)。

## 資料庫特性對比

| 特性 | MaxMind GeoLite2 | IPIP.NET 免費版 | 外部 API (ip-api) |
|------|-----------------|----------------|------------------|
| **類型** | 本地資料庫 | 本地資料庫 | HTTP API |
| **覆蓋範圍** | 全球 | 全球 | 全球 |
| **中國地區準確性** | 中等 | 高 | 中等 |
| **城市資訊** | 英文 + 部分中文 | 中文（含省份） | 英文 |
| **經緯度** | ✅ 有 | ❌ 無（付費版有） | ✅ 有 |
| **時區** | ✅ 有 | ❌ 無 | ✅ 有 |
| **ISO 國碼** | ✅ 有 | ❌ 無 | ✅ 有 |
| **更新頻率** | 每週二 | 不定期 | 即時 |
| **查詢速度** | ~1ms | ~1ms | ~300-500ms |
| **使用限制** | 無 | 無 | 45 req/min（免費） |
| **資料庫大小** | ~70MB | ~3.5MB | N/A |

### 使用建議

- **需要全球 IP 查詢 + 經緯度** → 使用智能路由（預設）
- **只需中國地區查詢** → 配置僅使用 IPIP
- **需要精確經緯度** → 使用 `/provider?provider=maxmind` 指定 MaxMind
- **需要中文省份+城市** → 中國 IP 會自動使用 IPIP
- **本地資料庫無城市資訊時** → 啟用外部 API 作為 fallback（會消耗 API 配額）
- **指定查詢外部 API** → 使用 `/provider?provider=ip-api`

### 資料庫維護

**MaxMind GeoLite2**
- **更新頻率**: 每週二
- **下載**: [MaxMind 官網](https://dev.maxmind.com/geoip/geolite2-free-geolocation-data)
- **建議**: 每月更新一次

**IPIP.NET**
- **更新頻率**: 不定期
- **下載**: [IPIP 官網](https://www.ipip.net/product/client.html)

```bash
# 更新資料庫後重啟服務
make docker-goip-restart
```

## 效能指標

- 單一查詢回應時間: < 10ms (P95)
- 批次查詢回應時間: < 50ms (P95, 100 IPs)
- 併發處理能力: > 1000 req/s
- 快取命中率: > 80%

## 開發

### 專案結構

```
goip/
├── cmd/server/          # 應用程式入口
├── internal/            # 私有應用程式碼
│   ├── handler/        # HTTP 處理器
│   ├── service/        # 業務邏輯
│   ├── repository/     # 資料存取層
│   ├── model/          # 資料模型
│   └── middleware/     # 中間件
├── pkg/                # 可共享的函式庫
├── config/             # 配置管理
├── data/               # IP 資料庫檔案 (MaxMind, IPIP)
├── build/              # 建置腳本、Dockerfile 和編譯產物
└── deployments/        # 部署配置（docker-compose）
    ├── goip/          # GoIP 服務部署
    └── redis/         # Redis 服務部署
```

### 運行測試

```bash
# 運行所有測試
make test

# 運行測試並生成覆蓋率報告
make test-coverage
```

### 建置

**編譯二進制檔案：**
```bash
# 編譯當前平台
make build
# 或
./build/build.sh

# 跨平台編譯
make build-all
# 或
./build/build.sh all
```

建置產物會放在 `build/` 目錄。

**建置 Docker 映像：**
```bash
# 使用 Makefile
make docker-build

# 指定版本
make docker-build-version VERSION=1.0.0

# 使用建置腳本
./build/docker-build.sh -v 1.0.0
```

## 部署

### Docker

```bash
# 建置映像
make docker-build

# 運行容器
docker run -d -p 8080:8080 \
  -v $(pwd)/data:/app/data:ro \
  --name goip \
  goip:latest
```

### Docker Compose

```bash
# 啟動所有服務
make docker-up

# 停止所有服務
make docker-down

# 重啟服務
make docker-restart

# 查看日誌
make docker-logs

# 查看服務狀態
make docker-ps
```

詳細部署說明請參考：
- [build/README.md](build/README.md) - 建置說明
- [deployments/README.md](deployments/README.md) - 部署說明

## 常用命令

```bash
# 開發
make run                # 本地運行服務
make test               # 運行測試
make build              # 編譯二進制

# Docker 建置
make docker-build       # 建置 Docker 映像

# Docker Compose 部署
make docker-up          # 啟動所有服務
make docker-down        # 停止所有服務
make docker-logs        # 查看日誌
make docker-ps          # 查看狀態

# 單獨管理服務
make docker-redis-up    # 啟動 Redis
make docker-redis-down  # 停止 Redis
make docker-goip-up     # 啟動 GoIP
make docker-goip-down   # 停止 GoIP

# 完整部署
make full-deploy        # 建置 + 啟動所有服務
```

## CI/CD

專案使用 GitHub Actions 進行自動化建置和測試：

- **觸發條件**: Push 到 `main` 或 `develop` 分支，或建立 Pull Request
- **執行項目**:
  - 單元測試（包含 race detector）
  - 程式碼品質檢查（golangci-lint）
  - Docker 映像建置驗證
  - 測試覆蓋率報告

詳見 [.github/workflows/build.yml](.github/workflows/build.yml)

## 授權

MIT License

## 貢獻

歡迎提交 Issue 和 Pull Request！

## 參考資源

### 文件
- [CLAUDE.md](CLAUDE.md) - AI 開發指南
- [設計文檔](docs/DESIGN.md) - 完整架構設計
- [多資料庫指南](docs/MULTI_DB_GUIDE.md) - 多資料庫使用說明
- [API 回應格式](docs/API_RESPONSE_FORMAT.md) - 詳細格式說明
- [建置說明](build/README.md) - 建置和 Docker 映像
- [部署說明](deployments/README.md) - Docker Compose 部署

### 外部資源
- [MaxMind GeoLite2](https://dev.maxmind.com/geoip/geolite2-free-geolocation-data)
- [IPIP.NET](https://www.ipip.net/product/client.html)
- [Gin Framework](https://gin-gonic.com/)
