# GoIP - IP 地理位置查詢服務設計文檔

## 專案概述

基於 MaxMind GeoLite2 資料庫的 IP 地理位置查詢 RESTful API 服務，使用 Golang 開發。

## 專案架構

### 目錄結構

```
goip/
├── cmd/
│   └── server/
│       └── main.go                 # 應用程式入口點
├── internal/
│   ├── handler/
│   │   └── ip_handler.go          # HTTP 請求處理器
│   ├── service/
│   │   └── ip_service.go          # 業務邏輯層
│   ├── repository/
│   │   ├── maxmind_repository.go  # MaxMind DB 存取層
│   │   └── cache_repository.go    # Redis Cache 存取層
│   ├── model/
│   │   └── ip_info.go             # 資料模型定義
│   └── middleware/
│       ├── logger.go              # 日誌中間件
│       ├── rate_limiter.go        # 限流中間件
│       └── recovery.go            # 錯誤恢復中間件
├── pkg/
│   └── validator/
│       └── ip_validator.go        # IP 驗證工具
├── config/
│   └── config.go                  # 配置管理
├── data/
│   └── GeoLite2-City.mmdb         # MaxMind 資料庫檔案
├── build/                         # 建置產物目錄
│   └── README.md
├── deployments/                   # 部署配置
│   ├── Dockerfile                 # Docker 映像建置檔案
│   ├── docker-compose.yml         # Docker Compose 配置（含 Redis）
│   └── README.md
├── scripts/                       # 輔助腳本
│   ├── build.sh                   # 建置腳本（跨平台）
│   └── README.md
├── go.mod
├── go.sum
├── Makefile
├── README.md
└── .env.example
```

## 核心功能模組

### API 端點設計

| 方法 | 路徑 | 描述 |
|------|------|------|
| GET | `/api/v1/ip/{ip}` | 查詢單一 IP 的國家資訊 |
| POST | `/api/v1/ip/batch` | 批次查詢多個 IP |
| GET | `/api/v1/health` | 健康檢查 |
| GET | `/api/v1/metrics` | 服務指標（可選）|

### API 回應格式

#### 單一 IP 查詢

**請求範例：**
```http
GET /api/v1/ip/8.8.8.8
```

**成功回應：**
```json
{
  "ip": "8.8.8.8",
  "country": {
    "iso_code": "US",
    "name": "United States",
    "name_zh": "美國"
  },
  "continent": {
    "code": "NA",
    "name": "North America"
  },
  "query_time_ms": 2
}
```

#### 批次查詢

**請求範例：**
```http
POST /api/v1/ip/batch
Content-Type: application/json

{
  "ips": ["8.8.8.8", "1.1.1.1", "140.112.0.1"]
}
```

**成功回應：**
```json
{
  "results": [
    {
      "ip": "8.8.8.8",
      "country": {
        "iso_code": "US",
        "name": "United States",
        "name_zh": "美國"
      },
      "continent": {
        "code": "NA",
        "name": "North America"
      }
    },
    {
      "ip": "1.1.1.1",
      "country": {
        "iso_code": "AU",
        "name": "Australia",
        "name_zh": "澳洲"
      },
      "continent": {
        "code": "OC",
        "name": "Oceania"
      }
    }
  ],
  "total": 3,
  "success": 2,
  "failed": 1
}
```

#### 錯誤回應

```json
{
  "error": "invalid IP address format",
  "code": "INVALID_IP",
  "timestamp": "2026-02-20T10:30:00Z"
}
```

### 錯誤碼定義

| 錯誤碼 | HTTP 狀態碼 | 描述 |
|--------|-------------|------|
| `INVALID_IP` | 400 | IP 格式無效 |
| `IP_NOT_FOUND` | 404 | IP 不在資料庫中 |
| `RATE_LIMIT_EXCEEDED` | 429 | 超過請求限制 |
| `INTERNAL_ERROR` | 500 | 內部伺服器錯誤 |
| `DB_ERROR` | 503 | 資料庫讀取錯誤 |

## 技術元件設計

### 1. Repository 層（資料存取層）

#### MaxMind Repository

**檔案：** `internal/repository/maxmind_repository.go`

**職責：**
- 初始化並管理 MaxMind DB 連接
- 提供 IP 查詢介面
- 處理資料庫讀取錯誤
- 支援資料庫熱更新（可選）

**介面定義：**
```go
type MaxMindRepository interface {
    LookupCountry(ip string) (*CountryInfo, error)
    Close() error
    Reload() error
}
```

**關鍵方法：**
- `NewMaxMindRepository(dbPath string)` - 初始化資料庫連接
- `LookupCountry(ip string)` - 查詢 IP 對應的國家資訊
- `Close()` - 關閉資料庫連接
- `Reload()` - 重新載入資料庫（用於熱更新）

#### Cache Repository (Redis)

**檔案：** `internal/repository/cache_repository.go`

**職責：**
- 管理 Redis 連接池
- 提供快取的 CRUD 操作
- 處理快取過期與淘汰
- 支援批次操作

**介面定義：**
```go
type CacheRepository interface {
    Get(ctx context.Context, key string) (*IPInfo, error)
    Set(ctx context.Context, key string, value *IPInfo, ttl time.Duration) error
    MGet(ctx context.Context, keys []string) (map[string]*IPInfo, error)
    MSet(ctx context.Context, items map[string]*IPInfo, ttl time.Duration) error
    Delete(ctx context.Context, keys ...string) error
    Exists(ctx context.Context, key string) (bool, error)
    FlushAll(ctx context.Context) error
    GetStats(ctx context.Context) (*CacheStats, error)
    Close() error
}
```

**關鍵方法：**
- `Get(key)` - 獲取單一快取
- `Set(key, value, ttl)` - 設定快取與過期時間
- `MGet(keys)` - 批次獲取多個快取
- `MSet(items, ttl)` - 批次設定多個快取
- `Delete(keys)` - 刪除快取
- `GetStats()` - 獲取快取統計（命中率、鍵數量等）

### 2. Service 層（業務邏輯層）

**檔案：** `internal/service/ip_service.go`

**職責：**
- 協調業務流程
- IP 格式驗證
- 快取管理（優先查詢 Redis）
- 查詢統計與監控

**介面定義：**
```go
type IPService interface {
    LookupIP(ip string) (*IPInfo, error)
    BatchLookup(ips []string) (*BatchResult, error)
    GetStats() *ServiceStats
    InvalidateCache(ips ...string) error
}
```

**查詢流程（Cache-Aside Pattern）：**
```
1. 接收 IP 查詢請求
2. 檢查 Redis 快取
   ├─ 命中 → 直接返回
   └─ 未命中 → 查詢 MaxMind DB
       ├─ 找到 → 寫入 Redis → 返回結果
       └─ 未找到 → 返回錯誤
```

**批次查詢優化：**
```
1. 使用 Redis MGET 批次查詢快取
2. 收集未命中的 IP
3. 批次查詢 MaxMind DB
4. 使用 Redis MSET 批次寫入快取
5. 合併結果返回
```

**關鍵功能：**
- `LookupIP(ip string)` - 單一 IP 查詢（Redis → MaxMind）
- `BatchLookup(ips []string)` - 批次查詢（優化快取存取）
- `GetStats()` - 獲取服務統計（包含快取命中率）
- `InvalidateCache(ips)` - 手動清除快取

### 3. Handler 層（HTTP 處理層）

**檔案：** `internal/handler/ip_handler.go`

**職責：**
- HTTP 請求/回應處理
- 參數解析與驗證
- 錯誤處理與格式化
- 回應序列化

**關鍵方法：**
- `HandleIPLookup()` - 處理單一 IP 查詢
- `HandleBatchLookup()` - 處理批次查詢
- `HandleHealth()` - 健康檢查
- `HandleMetrics()` - 服務指標

### 4. Middleware 層（中間件）

#### Logger Middleware
- 記錄每個請求的詳細資訊
- 包含請求方法、路徑、狀態碼、耗時
- 結構化日誌輸出

#### Rate Limiter Middleware
- **儲存後端：** Redis（分散式限流）
- **限流策略：** 滑動窗口計數器（Redis Sorted Set）
- **限流維度：**
  - 基於 IP 地址
  - 基於 API Key（可選）
- **限流配置：** 可配置每分鐘/每小時請求次數
- **超限處理：** 返回 429 狀態碼 + Retry-After header
- **多實例支援：** 使用 Redis 實現跨實例限流

#### Recovery Middleware
- 捕獲 panic 錯誤
- 記錄錯誤堆疊
- 返回友善的錯誤訊息

### 5. Model 層（資料模型）

**檔案：** `internal/model/ip_info.go`

**核心結構：**
```go
type IPInfo struct {
    IP        string       `json:"ip"`
    Country   CountryInfo  `json:"country"`
    Continent ContinentInfo `json:"continent"`
    QueryTimeMs int64      `json:"query_time_ms"`
}

type CountryInfo struct {
    ISOCode string `json:"iso_code"`
    Name    string `json:"name"`
    NameZh  string `json:"name_zh"`
}

type ContinentInfo struct {
    Code string `json:"code"`
    Name string `json:"name"`
}

type BatchResult struct {
    Results []IPInfo `json:"results"`
    Total   int      `json:"total"`
    Success int      `json:"success"`
    Failed  int      `json:"failed"`
}
```

## 非功能性需求

### 效能優化

#### 快取策略（Redis）

**技術選擇：** Redis 分散式快取

**快取層級：**
- **L1 - 本地快取（可選）：** 使用 sync.Map 或 go-cache 快取熱門 IP（1000 筆，TTL 5 分鐘）
- **L2 - Redis 快取：** 主要快取層，所有查詢結果都會寫入

**快取設計：**
- **Key 格式：** `goip:country:{ip}` （例如：`goip:country:8.8.8.8`）
- **Value 格式：** JSON 序列化的 IPInfo 結構
- **TTL 設定：** 預設 24 小時（可配置）
- **淘汰策略：** Redis LRU (allkeys-lru)
- **最大記憶體：** 建議 512MB - 1GB（視流量調整）

**快取模式：**
- **Cache-Aside：** 應用層控制快取讀寫
- **寫入策略：** 查詢成功後立即寫入快取
- **失效策略：**
  - TTL 自動過期
  - MaxMind DB 更新時可選擇性清空快取
  - 提供手動清除 API

**批次操作優化：**
- 使用 Redis Pipeline 減少網路往返
- MGET/MSET 批次讀寫
- 單次批次上限 100 個 key

**快取預熱（可選）：**
- 啟動時預載常用 IP（如 Google DNS、Cloudflare DNS）
- 支援從檔案載入熱門 IP 列表

**容錯機制：**
- Redis 不可用時自動降級到直接查詢 MaxMind DB
- 快取操作錯誤不影響主要查詢流程
- 記錄快取錯誤日誌供監控

#### 連線池管理
- MaxMind DB reader 使用單例模式
- 共享 DB 連接，避免重複開啟
- 支援併發讀取

#### 批次查詢限制
- 單次批次查詢上限：100 個 IP
- 超過限制返回 400 錯誤
- 可透過配置調整限制

### 可靠性

#### 限流保護
- **儲存：** Redis 實現分散式限流
- **策略：** 基於 IP 地址的滑動窗口限流
- **預設限制：** 100 requests/minute per IP
- **可配置：** 支援動態調整限流規則（熱更新）
- **回應：** 429 Too Many Requests + Retry-After header
- **容錯：** Redis 故障時降級到本地限流（sync.Map）

#### 超時控制
- **讀取超時：** 10 秒
- **寫入超時：** 10 秒
- **資料庫查詢超時：** 5 秒

#### 優雅關閉
- 監聽 SIGTERM/SIGINT 信號
- 等待現有請求處理完成
- 關閉資料庫連接
- 最長等待時間：30 秒

#### 錯誤處理
- 統一的錯誤碼系統
- 結構化錯誤訊息
- 錯誤堆疊記錄（非生產環境）
- 友善的用戶錯誤提示

### 可維護性

#### 結構化日誌
- **技術選擇：** zerolog 或 zap
- **日誌級別：** DEBUG, INFO, WARN, ERROR
- **輸出格式：** JSON（生產）/ 彩色文字（開發）
- **欄位包含：** timestamp, level, method, path, status, duration, error

#### 配置管理
- 支援環境變數（優先）
- 支援 YAML 配置檔案
- 使用 viper 統一管理
- 敏感資訊從環境變數讀取

#### 健康檢查
- 檢查 HTTP 伺服器狀態
- 檢查 MaxMind DB 是否可讀
- 檢查 Redis 連線狀態
- 返回詳細的健康狀態與依賴服務狀態

#### 版本管理
- API 路徑包含版本號（/api/v1/）
- 支援多版本並存
- 向後兼容保證

## 配置設計

### 配置檔案範例

**檔案：** `config/config.yaml`

```yaml
server:
  port: 8080
  read_timeout: 10s
  write_timeout: 10s
  shutdown_timeout: 30s

maxmind:
  db_path: "./data/GeoLite2-Country.mmdb"
  auto_update: false
  update_interval: 24h

redis:
  host: "localhost"
  port: 6379
  password: ""
  db: 0
  pool_size: 10
  min_idle_conns: 5
  max_retries: 3
  dial_timeout: 5s
  read_timeout: 3s
  write_timeout: 3s

cache:
  enabled: true
  ttl: 24h
  local_cache_enabled: false  # L1 本地快取（可選）
  local_cache_size: 1000
  local_cache_ttl: 5m

rate_limit:
  enabled: true
  requests_per_minute: 100
  requests_per_hour: 5000
  burst: 10
  storage: "redis"  # redis 或 memory

batch:
  max_size: 100

log:
  level: "info"
  format: "json"
  output: "stdout"
```

### 環境變數

```bash
# 伺服器配置
SERVER_PORT=8080
SERVER_READ_TIMEOUT=10s
SERVER_WRITE_TIMEOUT=10s

# MaxMind 配置
MAXMIND_DB_PATH=./data/GeoLite2-Country.mmdb

# Redis 配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
REDIS_POOL_SIZE=10

# 快取配置
CACHE_ENABLED=true
CACHE_TTL=24h
LOCAL_CACHE_ENABLED=false

# 限流配置
RATE_LIMIT_ENABLED=true
RATE_LIMIT_RPM=100
RATE_LIMIT_RPH=5000
RATE_LIMIT_STORAGE=redis

# 日誌配置
LOG_LEVEL=info
LOG_FORMAT=json
```

## 相依套件

### 核心套件

| 套件 | 用途 | 版本建議 |
|------|------|----------|
| `github.com/gin-gonic/gin` | Web 框架 | v1.9+ |
| `github.com/oschwald/geoip2-golang` | MaxMind DB 讀取 | v1.9+ |
| `github.com/redis/go-redis/v9` | Redis 客戶端 | v9.5+ |
| `github.com/spf13/viper` | 配置管理 | v1.18+ |
| `github.com/rs/zerolog` | 結構化日誌 | v1.32+ |
| `github.com/ulule/limiter/v3` | 限流中間件（支援 Redis）| v3.11+ |
| `github.com/go-playground/validator/v10` | 資料驗證 | v10.19+ |
| `github.com/patrickmn/go-cache` | 本地快取（L1，可選）| v2.1+ |

### 工具套件

| 套件 | 用途 |
|------|------|
| `github.com/stretchr/testify` | 單元測試 |
| `github.com/golang/mock` | Mock 測試 |
| `github.com/swaggo/gin-swagger` | API 文件（可選）|

## 開發階段規劃

### Phase 1: 基礎功能（Week 1）

1. **專案初始化**
   - 建立專案結構
   - 初始化 Go modules
   - 設定基本配置

2. **MaxMind DB 整合**
   - Repository 層實作
   - DB 讀取與查詢功能
   - 錯誤處理

3. **單一 IP 查詢 API**
   - Handler 層實作
   - Service 層實作
   - 基本路由設定

4. **基本錯誤處理**
   - 統一錯誤回應格式
   - IP 格式驗證
   - 錯誤碼定義

**預期成果：** 可運行的基本 API，支援單一 IP 查詢

### Phase 2: 增強功能（Week 2）

5. **批次查詢 API**
   - 批次查詢端點
   - 請求大小限制
   - 批次結果處理

6. **快取機制**
   - Cache 層整合
   - TTL 管理
   - 快取統計

7. **請求驗證與限流**
   - Rate limiter middleware
   - 輸入驗證強化
   - API Key 認證（可選）

8. **結構化日誌**
   - Logger middleware
   - 請求追蹤
   - 效能監控

**預期成果：** 功能完整的 API，具備快取和限流

### Phase 3: 生產就緒（Week 3）

9. **健康檢查與監控**
   - Health check 端點
   - Metrics 端點
   - Prometheus 整合（可選）

10. **測試覆蓋**
    - 單元測試
    - 整合測試
    - 效能測試

11. **Docker 化**
    - Dockerfile
    - docker-compose.yml
    - 多階段建置

12. **文件與部署**
    - API 文件
    - 部署指南
    - 使用說明

**預期成果：** 可部署到生產環境的完整服務

## 安全考量

### 輸入驗證
- **IP 格式驗證：** 使用標準函式庫驗證 IPv4/IPv6
- **參數清理：** 過濾特殊字元
- **長度限制：** 限制批次查詢數量
- **類型檢查：** 強制類型驗證

### API 保護
- **限流：** 防止 API 濫用
- **認證：** API Key 機制（可選）
- **CORS：** 配置適當的跨域政策
- **請求大小限制：** 防止大型請求攻擊

### 傳輸安全
- **HTTPS：** 生產環境強制使用 TLS
- **HSTS：** 強制 HTTPS 重導向
- **安全 Headers：** X-Frame-Options, X-Content-Type-Options 等

### 資料安全
- **敏感資訊：** 不記錄完整 IP（遵守隱私政策）
- **日誌脫敏：** 移除敏感資料
- **資料庫保護：** 只讀權限

### 錯誤處理
- **資訊洩漏：** 不暴露內部錯誤細節
- **堆疊追蹤：** 僅在開發環境顯示
- **友善訊息：** 給用戶清楚的錯誤說明

## 效能指標

### 預期效能

| 指標 | 目標值 |
|------|--------|
| 單一查詢回應時間 | < 10ms (P95) |
| 批次查詢回應時間 | < 50ms (P95, 100 IPs) |
| 併發處理能力 | > 1000 req/s |
| 快取命中率 | > 80% |
| 記憶體使用 | < 200MB |

### 監控指標

- **請求指標：** QPS, 回應時間, 錯誤率
- **快取指標：** Redis 命中率, 本地快取命中率, 淘汰次數, 記憶體使用
- **Redis 指標：** 連線數, 指令延遲, 網路流量, 鍵數量
- **資源指標：** CPU, 記憶體, 磁碟 I/O
- **業務指標：** 熱門 IP, 國家分布

## 部署架構

### 單機部署
```
[Client] -> [Nginx/Traefik] -> [GoIP Service]
                                     ├─> [Redis]
                                     └─> [MaxMind DB]
```

### 高可用部署
```
[Client] -> [Load Balancer]
              ├─> [GoIP Instance 1] ─┐
              ├─> [GoIP Instance 2] ─┼─> [Redis Cluster/Sentinel]
              └─> [GoIP Instance 3] ─┘        ├─> Primary
                      │                        └─> Replica(s)
                      └──────────────> [MaxMind DB] (共享或各自載入)
```

### Docker Compose 部署
```yaml
version: '3.8'
services:
  goip:
    build: .
    ports:
      - "8080:8080"
    depends_on:
      - redis
    environment:
      - REDIS_HOST=redis
      - REDIS_PORT=6379
    volumes:
      - ./data:/app/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    command: redis-server --maxmemory 512mb --maxmemory-policy allkeys-lru

volumes:
  redis-data:
```

**Docker 部署要點：**
- 使用多階段建置減少映像大小
- 包含 MaxMind DB 於映像中（或掛載）
- 健康檢查配置（檢查 HTTP + Redis）
- 資源限制設定
- Redis 持久化配置（RDB + AOF）

## 未來擴展

### 短期擴展
- [ ] 支援 City 級別查詢（GeoLite2-City）
- [ ] 增加 ASN 資訊查詢
- [ ] WebSocket 即時查詢
- [ ] Prometheus metrics 導出

### 長期擴展
- [ ] 自動更新 MaxMind DB
- [ ] 多資料源整合（IP2Location, IPIP.net）
- [ ] Redis Cluster 支援
- [ ] Redis Sentinel 高可用
- [ ] GraphQL API 支援
- [ ] gRPC 介面

## Redis 整合詳細設計

### Redis 使用場景

本專案使用 Redis 作為：
1. **查詢結果快取** - 減少 MaxMind DB 讀取
2. **分散式限流** - 跨實例的請求限流
3. **會話管理** - API Key 驗證（可選）
4. **統計資料** - 即時統計與監控

### Redis 資料結構設計

#### 1. IP 查詢快取

**資料類型：** String

**Key 命名：**
```
goip:country:{ip}           # 國家級別查詢
goip:city:{ip}              # 城市級別查詢（未來擴展）
```

**Value 結構：**
```json
{
  "ip": "8.8.8.8",
  "country": {
    "iso_code": "US",
    "name": "United States",
    "name_zh": "美國"
  },
  "continent": {
    "code": "NA",
    "name": "North America"
  },
  "cached_at": 1708412345
}
```

**TTL：** 24 小時（86400 秒）

**預估容量：**
- 每筆記錄約 200 bytes
- 100 萬筆記錄約 200 MB
- 建議 Redis 記憶體：512MB - 1GB

#### 2. 限流計數器

**資料類型：** Sorted Set（滑動窗口）

**Key 命名：**
```
goip:ratelimit:{ip}:minute  # 每分鐘限流
goip:ratelimit:{ip}:hour    # 每小時限流
```

**實作方式：**
```redis
ZADD goip:ratelimit:192.168.1.1:minute {timestamp} {request_id}
ZREMRANGEBYSCORE goip:ratelimit:192.168.1.1:minute 0 {60_seconds_ago}
ZCARD goip:ratelimit:192.168.1.1:minute
```

**TTL：** 自動過期（65 秒 for minute, 3700 秒 for hour）

#### 3. 統計資料

**資料類型：** Hash / Sorted Set

**Key 命名：**
```
goip:stats:daily:{date}     # 每日統計
goip:stats:hot_ips          # 熱門 IP 排行（Sorted Set）
goip:stats:countries        # 國家分布（Hash）
```

**範例：**
```redis
# 每日統計
HINCRBY goip:stats:daily:2026-02-20 total_queries 1
HINCRBY goip:stats:daily:2026-02-20 cache_hits 1

# 熱門 IP 排行
ZINCRBY goip:stats:hot_ips 1 "8.8.8.8"

# 國家分布
HINCRBY goip:stats:countries US 1
```

### Redis 連線管理

#### 連線池配置

```go
redis.NewClient(&redis.Options{
    Addr:         "localhost:6379",
    Password:     "",
    DB:           0,
    PoolSize:     10,           // 最大連線數
    MinIdleConns: 5,            // 最小空閒連線
    MaxRetries:   3,            // 最大重試次數
    DialTimeout:  5 * time.Second,
    ReadTimeout:  3 * time.Second,
    WriteTimeout: 3 * time.Second,
    PoolTimeout:  4 * time.Second,
})
```

#### 健康檢查

```go
func (r *CacheRepository) HealthCheck(ctx context.Context) error {
    ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()

    return r.client.Ping(ctx).Err()
}
```

### Redis 容錯設計

#### 降級策略

```go
func (s *IPService) LookupIP(ip string) (*IPInfo, error) {
    // 1. 嘗試從 Redis 快取讀取
    result, err := s.cache.Get(ctx, ip)
    if err == nil {
        return result, nil
    }

    // 2. Redis 錯誤時記錄但不中斷服務
    if err != redis.Nil {
        s.logger.Warn().Err(err).Msg("Redis cache error, fallback to DB")
    }

    // 3. 查詢 MaxMind DB
    result, err = s.maxmind.LookupCountry(ip)
    if err != nil {
        return nil, err
    }

    // 4. 嘗試寫入快取（失敗不影響回應）
    if cacheErr := s.cache.Set(ctx, ip, result, 24*time.Hour); cacheErr != nil {
        s.logger.Warn().Err(cacheErr).Msg("Failed to cache result")
    }

    return result, nil
}
```

#### 斷路器模式（可選）

使用 `github.com/sony/gobreaker` 在 Redis 頻繁失敗時自動降級：

```go
cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
    Name:        "redis-cache",
    MaxRequests: 3,
    Interval:    60 * time.Second,
    Timeout:     30 * time.Second,
    ReadyToTrip: func(counts gobreaker.Counts) bool {
        failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
        return counts.Requests >= 10 && failureRatio >= 0.6
    },
})
```

### Redis 效能優化

#### Pipeline 批次操作

```go
func (r *CacheRepository) MGet(ctx context.Context, ips []string) (map[string]*IPInfo, error) {
    pipe := r.client.Pipeline()

    // 批次查詢
    cmds := make([]*redis.StringCmd, len(ips))
    for i, ip := range ips {
        key := fmt.Sprintf("goip:country:%s", ip)
        cmds[i] = pipe.Get(ctx, key)
    }

    _, err := pipe.Exec(ctx)
    if err != nil && err != redis.Nil {
        return nil, err
    }

    // 解析結果
    results := make(map[string]*IPInfo)
    for i, cmd := range cmds {
        val, err := cmd.Result()
        if err == nil {
            var info IPInfo
            if json.Unmarshal([]byte(val), &info) == nil {
                results[ips[i]] = &info
            }
        }
    }

    return results, nil
}
```

#### 本地快取（L1）

對於極高頻查詢（如 8.8.8.8），使用本地記憶體快取：

```go
type TwoLevelCache struct {
    local  *cache.Cache              // L1: 本地快取
    redis  *CacheRepository          // L2: Redis
}

func (c *TwoLevelCache) Get(ip string) (*IPInfo, error) {
    // L1: 檢查本地快取
    if val, found := c.local.Get(ip); found {
        return val.(*IPInfo), nil
    }

    // L2: 檢查 Redis
    result, err := c.redis.Get(context.Background(), ip)
    if err == nil {
        c.local.Set(ip, result, 5*time.Minute) // 寫入 L1
        return result, nil
    }

    return nil, err
}
```

### Redis 監控指標

#### 關鍵指標

```go
type RedisMetrics struct {
    // 連線指標
    PoolHits        uint64  // 連線池命中
    PoolMisses      uint64  // 連線池未命中
    PoolTimeouts    uint64  // 連線池超時

    // 效能指標
    CacheHits       uint64  // 快取命中
    CacheMisses     uint64  // 快取未命中
    AvgLatency      float64 // 平均延遲（ms）

    // 容量指標
    UsedMemory      uint64  // 使用記憶體（bytes）
    KeyCount        uint64  // 鍵總數
    EvictedKeys     uint64  // 淘汰鍵數
}
```

#### Prometheus 匯出

```go
var (
    redisCacheHits = promauto.NewCounter(prometheus.CounterOpts{
        Name: "goip_redis_cache_hits_total",
        Help: "Total number of Redis cache hits",
    })

    redisCacheMisses = promauto.NewCounter(prometheus.CounterOpts{
        Name: "goip_redis_cache_misses_total",
        Help: "Total number of Redis cache misses",
    })

    redisLatency = promauto.NewHistogram(prometheus.HistogramOpts{
        Name:    "goip_redis_latency_seconds",
        Help:    "Redis operation latency",
        Buckets: []float64{.001, .005, .01, .025, .05, .1},
    })
)
```

### Redis 維運建議

#### 記憶體配置

```redis
# redis.conf
maxmemory 512mb
maxmemory-policy allkeys-lru
maxmemory-samples 5
```

#### 持久化策略

**開發/測試環境：**
```redis
# 不需持久化（快取可重建）
save ""
appendonly no
```

**生產環境（可選）：**
```redis
# RDB 備份（每小時一次）
save 3600 1
# AOF 持久化（每秒 fsync）
appendonly yes
appendfsync everysec
```

#### 高可用方案

**Redis Sentinel（3 節點）：**
```yaml
sentinel monitor goip-redis redis-master 6379 2
sentinel down-after-milliseconds goip-redis 5000
sentinel failover-timeout goip-redis 10000
```

**Redis Cluster（6 節點）：**
- 3 個主節點 + 3 個從節點
- 自動分片與故障轉移
- 適合大流量場景

### Redis 安全建議

1. **認證：** 設定強密碼（requirepass）
2. **網路隔離：** 只允許應用伺服器存取
3. **指令限制：** 禁用危險指令（FLUSHALL, FLUSHDB, CONFIG）
4. **加密傳輸：** 使用 TLS（Redis 6.0+）
5. **資源限制：** 設定 maxclients 限制連線數

## 參考資源

- [MaxMind GeoIP2 Golang API](https://github.com/oschwald/geoip2-golang)
- [MaxMind GeoLite2 資料庫](https://dev.maxmind.com/geoip/geolite2-free-geolocation-data)
- [Gin Web Framework](https://gin-gonic.com/)
- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [Go Redis Client](https://github.com/redis/go-redis)
- [Redis Best Practices](https://redis.io/docs/manual/patterns/)
