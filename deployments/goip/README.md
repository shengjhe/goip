# GoIP 服務部署

GoIP 主服務的部署配置檔案。

## 檔案說明

- **Dockerfile**: GoIP 服務的 Docker 映像建置檔案
- **docker-compose.yml**: GoIP 服務的編排配置
- **.env.example**: 環境變數範例（可選）

## 前置需求

1. Docker 與 Docker Compose 已安裝
2. 依賴服務（Redis）已啟動
3. MaxMind 資料庫檔案已放置於 `data/` 目錄

## 部署步驟

### 1. 啟動依賴服務

先啟動 Redis 等依賴服務：

```bash
cd ../dependencies
docker-compose up -d
```

### 2. 建置 Docker 映像

使用提供的建置腳本：

```bash
# 從專案根目錄執行
./scripts/docker-build.sh -v 1.0.0

# 或使用 Makefile
make docker-build
```

### 3. 啟動 GoIP 服務

```bash
# 在此目錄下
docker-compose up -d

# 或從專案根目錄
docker-compose -f deployments/goip/docker-compose.yml up -d
```

### 4. 驗證服務

```bash
# 健康檢查
curl http://localhost:8080/api/v1/health

# 測試查詢
curl http://localhost:8080/api/v1/ip/8.8.8.8
```

## 環境變數

GoIP 服務支援以下環境變數（可在 docker-compose.yml 中配置）：

| 變數名稱 | 預設值 | 說明 |
|---------|--------|------|
| SERVER_PORT | 8080 | HTTP 服務埠 |
| REDIS_HOST | redis | Redis 主機位址 |
| REDIS_PORT | 6379 | Redis 埠 |
| REDIS_PASSWORD | - | Redis 密碼 |
| MAXMIND_DB_PATH | /app/data/GeoLite2-City.mmdb | MaxMind 資料庫路徑 |
| CACHE_ENABLED | true | 是否啟用快取 |
| CACHE_TTL | 24h | 快取過期時間 |
| RATE_LIMIT_ENABLED | true | 是否啟用限流 |
| RATE_LIMIT_RPM | 100 | 每分鐘請求限制 |
| RATE_LIMIT_RPH | 5000 | 每小時請求限制 |
| LOG_LEVEL | info | 日誌級別 |
| LOG_FORMAT | json | 日誌格式 |

## 常用命令

```bash
# 啟動服務
docker-compose up -d

# 停止服務
docker-compose down

# 查看日誌
docker-compose logs -f goip

# 重啟服務
docker-compose restart goip

# 查看狀態
docker-compose ps
```

## 網路配置

GoIP 服務使用外部網路 `goip-network` 與依賴服務通訊。確保該網路已由依賴服務建立。

## 故障排除

### 服務無法啟動

1. 檢查依賴服務是否正常運行：
   ```bash
   docker network inspect goip-network
   ```

2. 檢查日誌：
   ```bash
   docker-compose logs goip
   ```

### 無法連接 Redis

1. 確認 Redis 服務在 `goip-network` 網路中
2. 檢查環境變數 `REDIS_HOST` 和 `REDIS_PORT`
3. 測試 Redis 連接：
   ```bash
   docker exec -it goip-service ping redis
   ```

### MaxMind 資料庫找不到

確認 `data/` 目錄包含 MaxMind 資料庫檔案，並且 volume 映射正確：
```bash
docker exec -it goip-service ls -la /app/data
```
