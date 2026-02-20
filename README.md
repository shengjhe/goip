# GoIP - IP 地理位置查詢服務

基於 MaxMind GeoLite2 資料庫的高效能 IP 地理位置查詢 RESTful API 服務。

## 特色

- 🚀 **高效能**: Redis 分散式快取 + 本地快取雙層架構
- 🌍 **準確資料**: 基於 MaxMind GeoLite2 資料庫
- 🔒 **限流保護**: Redis 實現的分散式限流
- 📊 **批次查詢**: 支援批次 IP 查詢，使用 Pipeline 優化
- 🐳 **容器化**: Docker Compose 一鍵部署
- 📈 **可監控**: 支援健康檢查和統計 API

## 技術棧

- **語言**: Go 1.26+
- **Web 框架**: Gin
- **快取**: Redis 7+
- **資料庫**: MaxMind GeoLite2
- **日誌**: Zerolog
- **配置**: Viper

## 快速開始

### 前置需求

- Docker & Docker Compose（推薦）
- 或 Go 1.26+ & Redis 7.0+
- MaxMind GeoLite2 資料庫

### 安裝

1. Clone 專案
```bash
git clone https://github.com/axiom/goip.git
cd goip
```

2. 下載 MaxMind 資料庫
- 註冊 [MaxMind](https://www.maxmind.com/en/geolite2/signup) 帳號
- 下載 GeoLite2-City.mmdb
- 放置到 `data/` 目錄
- **資料庫更新**: MaxMind 每週二更新 GeoLite2 資料庫，建議定期更新以確保資料準確性

3. 複製環境變數範例（可選）
```bash
cp .env.example .env
# 根據需要編輯 .env
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

### 查詢單一 IP

```bash
GET /api/v1/ip/{ip}
```

**範例請求:**
```bash
curl http://localhost:8080/api/v1/ip/8.8.8.8
```

**範例回應:**
```json
{
  "ip": "8.8.8.8",
  "country": {
    "iso_code": "US",
    "name": "United States",
    "name_zh": "美国"
  },
  "continent": {
    "code": "NA",
    "name": "North America"
  },
  "query_time_ms": 1
}
```

### 批次查詢

```bash
POST /api/v1/ip/batch
Content-Type: application/json
```

**範例請求:**
```bash
curl -X POST http://localhost:8080/api/v1/ip/batch \
  -H "Content-Type: application/json" \
  -d '{"ips": ["8.8.8.8", "1.1.1.1", "140.112.1.1"]}'
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
      "continent": {
        "code": "NA",
        "name": "North America"
      },
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

主要環境變數（詳見 `.env.example`）：

| 變數名稱 | 預設值 | 說明 |
|---------|--------|------|
| SERVER_PORT | 8080 | HTTP 服務端口 |
| REDIS_HOST | redis | Redis 主機位址 |
| REDIS_PORT | 6379 | Redis 端口 |
| MAXMIND_DB_PATH | ./data/GeoLite2-City.mmdb | MaxMind 資料庫路徑 |
| CACHE_TTL | 24h | 快取過期時間 |
| RATE_LIMIT_RPM | 100 | 每分鐘請求限制 |
| LOG_LEVEL | info | 日誌級別 |

完整配置說明請參考 [DESIGN.md](DESIGN.md)。

### MaxMind 資料庫維護

MaxMind GeoLite2 資料庫需要定期更新以確保資料準確性：

- **更新頻率**: MaxMind 每週二發布新版本
- **檔案位置**: `data/GeoLite2-City.mmdb` (約 54MB)
- **建議**: 每月至少更新一次資料庫
- **下載方式**: 從 [MaxMind 官網](https://dev.maxmind.com/geoip/geolite2-free-geolocation-data)下載最新版本
- **熱更新**: 更新資料庫檔案後需重啟服務以載入新資料

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
├── data/               # MaxMind 資料庫檔案
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

## 授權

MIT License

## 貢獻

歡迎提交 Issue 和 Pull Request！

## 參考資源

- [設計文檔](DESIGN.md) - 完整架構設計
- [建置說明](build/README.md) - 建置和 Docker 映像
- [部署說明](deployments/README.md) - Docker Compose 部署
- [MaxMind GeoLite2](https://dev.maxmind.com/geoip/geolite2-free-geolocation-data)
- [Gin Framework](https://gin-gonic.com/)
