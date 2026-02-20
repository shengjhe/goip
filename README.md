# GoIP - IP 地理位置查詢服務

基於 MaxMind GeoLite2 資料庫的高效能 IP 地理位置查詢 RESTful API 服務。

## 特色

- 🚀 **高效能**: Redis 分散式快取 + 本地快取雙層架構
- 🌍 **準確資料**: 基於 MaxMind GeoLite2 資料庫
- 🔒 **限流保護**: Redis 實現的分散式限流
- 📊 **批次查詢**: 支援批次 IP 查詢，使用 Pipeline 優化
- 🐳 **容器化**: Docker Compose 一鍵部署
- 📈 **可監控**: 支援 Prometheus metrics 導出

## 技術棧

- **語言**: Go 1.21+
- **Web 框架**: Gin
- **快取**: Redis 7+
- **資料庫**: MaxMind GeoLite2
- **日誌**: Zerolog
- **配置**: Viper

## 快速開始

### 前置需求

- Go 1.21 或更高版本
- Redis 7.0 或更高版本
- MaxMind GeoLite2 資料庫

### 安裝

1. Clone 專案
```bash
git clone https://github.com/axiom/goip.git
cd goip
```

2. 複製並配置環境變數
```bash
cp .env.example .env
# 編輯 .env 填入你的配置
```

3. 下載 MaxMind 資料庫
- 註冊 [MaxMind](https://www.maxmind.com/en/geolite2/signup) 帳號
- 下載 GeoLite2-Country.mmdb 或 GeoLite2-City.mmdb
- 放置到 `data/` 目錄

### 使用 Docker Compose 運行

```bash
docker-compose up -d
```

服務將在 `http://localhost:8080` 啟動。

### 本地開發運行

```bash
# 安裝依賴
go mod download

# 啟動 Redis（如果沒有運行）
docker run -d -p 6379:6379 redis:7-alpine

# 運行服務
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
    "name_zh": "美國"
  },
  "continent": {
    "code": "NA",
    "name": "North America"
  },
  "query_time_ms": 2
}
```

### 批次查詢

```bash
POST /api/v1/ip/batch
Content-Type: application/json

{
  "ips": ["8.8.8.8", "1.1.1.1"]
}
```

### 健康檢查

```bash
GET /api/v1/health
```

## 配置說明

主要配置項目請參考 `.env.example`：

- **SERVER_PORT**: 服務監聽端口（預設 8080）
- **REDIS_HOST**: Redis 伺服器地址
- **CACHE_TTL**: 快取過期時間（預設 24h）
- **RATE_LIMIT_RPM**: 每分鐘請求限制（預設 100）

完整配置說明請參考 [DESIGN.md](DESIGN.md)。

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
└── data/               # MaxMind 資料庫檔案
```

### 運行測試

```bash
go test ./...
```

### 建置

```bash
go build -o bin/goip cmd/server/main.go
```

## 部署

### Docker

```bash
docker build -t goip:latest .
docker run -d -p 8080:8080 --env-file .env goip:latest
```

### Docker Compose

參考 `docker-compose.yml` 進行部署。

## 授權

MIT License

## 貢獻

歡迎提交 Issue 和 Pull Request！

## 參考資源

- [設計文檔](DESIGN.md)
- [MaxMind GeoLite2](https://dev.maxmind.com/geoip/geolite2-free-geolocation-data)
- [Gin Framework](https://gin-gonic.com/)
