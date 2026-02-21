# GoIP Docker 使用說明

## 🚀 快速開始

### 1. 建置 Docker 映像

```bash
docker build -t goip:latest -f build/Dockerfile .
```

### 2. 啟動服務

```bash
# 使用 docker-compose（推薦）
docker-compose -f docker-compose.test.yml up -d

# 查看日誌
docker-compose -f docker-compose.test.yml logs -f goip

# 檢查狀態
docker-compose -f docker-compose.test.yml ps
```

### 3. 快速測試

```bash
./docker-quick-test.sh
```

## 📋 服務資訊

### 端口映射

| 服務 | 容器端口 | 主機端口 |
|------|---------|---------|
| GoIP | 8080 | 8080 |
| Redis | 6379 | 6380 |

### 容器名稱

- GoIP: `goip-test`
- Redis: `goip-redis-test`

## 🔧 配置

### 資料庫檔案

確保以下檔案存在於 `./data/` 目錄：

```
./data/
├── GeoLite2-City.mmdb    # MaxMind 資料庫
└── ipipfree.ipdb         # IPIP.NET 資料庫
```

### 配置檔案

使用 `config.yaml` 配置多資料庫：

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

## 🧪 API 測試

### 基本查詢

```bash
# 健康檢查
curl http://localhost:8080/api/v1/health | jq .

# 查看可用提供者
curl http://localhost:8080/api/v1/providers | jq .

# 查詢 IP（智能路由）
curl http://localhost:8080/api/v1/ip/8.8.8.8 | jq .
```

### 智能路由測試

```bash
# 中國 IP - 自動使用 IPIP
curl http://localhost:8080/api/v1/ip/114.114.114.114 | jq .

# 海外 IP - 自動使用 MaxMind
curl http://localhost:8080/api/v1/ip/8.8.8.8 | jq .

# 中國城市 IP（杭州）
curl http://localhost:8080/api/v1/ip/42.120.160.1 | jq .
```

### 指定提供者查詢

```bash
# 強制使用 IPIP
curl "http://localhost:8080/api/v1/ip/8.8.8.8/provider?provider=ipip" | jq .

# 強制使用 MaxMind
curl "http://localhost:8080/api/v1/ip/42.120.160.1/provider?provider=maxmind" | jq .
```

### 批次查詢

```bash
curl http://localhost:8080/api/v1/ip/batch \
  -H "Content-Type: application/json" \
  -d '{
    "ips": ["8.8.8.8", "114.114.114.114", "42.120.160.1"]
  }' | jq .
```

### 服務統計

```bash
# 查詢統計
curl http://localhost:8080/api/v1/stats | jq .

# 快取統計
curl http://localhost:8080/api/v1/cache/stats | jq .
```

## 🛠️ 管理指令

### 容器管理

```bash
# 啟動服務
docker-compose -f docker-compose.test.yml up -d

# 停止服務
docker-compose -f docker-compose.test.yml down

# 重啟服務
docker-compose -f docker-compose.test.yml restart goip

# 查看日誌
docker-compose -f docker-compose.test.yml logs -f goip

# 進入容器
docker exec -it goip-test sh

# 查看容器狀態
docker-compose -f docker-compose.test.yml ps
```

### 快取管理

```bash
# 清除特定 IP 快取
curl -X POST http://localhost:8080/api/v1/cache/invalidate \
  -H "Content-Type: application/json" \
  -d '{"ips": ["8.8.8.8"]}'

# 查看快取統計
curl http://localhost:8080/api/v1/cache/stats | jq .
```

### 完全清理

```bash
# 停止並刪除容器和卷
docker-compose -f docker-compose.test.yml down -v

# 刪除映像
docker rmi goip:latest
```

## 📊 監控

### 健康檢查

Docker 容器內建健康檢查，每 10 秒檢查一次：

```bash
# 查看健康狀態
docker inspect goip-test | jq '.[0].State.Health'
```

### 日誌監控

```bash
# 即時查看日誌
docker-compose -f docker-compose.test.yml logs -f goip

# 查看最後 50 行
docker-compose -f docker-compose.test.yml logs --tail=50 goip
```

## 🔍 故障排除

### 問題 1: 服務無法啟動

**檢查資料庫檔案：**
```bash
ls -lh data/
# 應該看到 GeoLite2-City.mmdb 和 ipipfree.ipdb
```

**查看錯誤日誌：**
```bash
docker-compose -f docker-compose.test.yml logs goip
```

### 問題 2: 端口被佔用

**修改端口映射：**

編輯 `docker-compose.test.yml`：
```yaml
ports:
  - "8081:8080"  # 改用 8081
```

### 問題 3: Redis 連接失敗

**檢查 Redis 狀態：**
```bash
docker-compose -f docker-compose.test.yml logs redis
docker exec goip-redis-test redis-cli ping
```

### 問題 4: 查詢總是返回錯誤

**檢查資料庫是否載入：**
```bash
curl http://localhost:8080/api/v1/health | jq .
# 應該看到 maxmind: "healthy"
```

**查看可用提供者：**
```bash
curl http://localhost:8080/api/v1/providers | jq .
# 應該看到 ["ipip", "maxmind"]
```

## 📦 生產環境部署

### 使用 Docker Compose

```yaml
version: '3.8'

services:
  goip:
    image: goip:latest
    ports:
      - "8080:8080"
    environment:
      - REDIS_HOST=redis
      - CACHE_ENABLED=true
      - LOG_LEVEL=info
    volumes:
      - ./data:/app/data:ro
      - ./config.yaml:/app/config.yaml:ro
    restart: always
    depends_on:
      - redis

  redis:
    image: redis:7-alpine
    volumes:
      - redis-data:/data
    restart: always

volumes:
  redis-data:
```

### 環境變數

支援的環境變數：

```bash
# 伺服器配置
SERVER_PORT=8080

# Redis 配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=

# 快取配置
CACHE_ENABLED=true
CACHE_TTL=24h

# 限流配置
RATE_LIMIT_ENABLED=false
RATE_LIMIT_RPM=100
RATE_LIMIT_RPH=5000

# 日誌配置
LOG_LEVEL=info
LOG_FORMAT=console
```

## 📝 測試腳本

專案提供了多個測試腳本：

1. **docker-quick-test.sh** - 快速功能測試
2. **docker-build-and-test.sh** - 完整建置和測試流程
3. **test_response_format.sh** - 回應格式測試
4. **test_multi_db.sh** - 多資料庫功能測試

```bash
# 執行快速測試
./docker-quick-test.sh

# 完整測試（包含建置）
./docker-build-and-test.sh
```

## 🎯 效能優化

### 1. 快取設定

調整快取 TTL 以獲得更好的效能：

```yaml
cache:
  enabled: true
  ttl: 24h  # 根據需求調整
```

### 2. Redis 連接池

```yaml
redis:
  pool_size: 10
  min_idle_conns: 5
```

### 3. 資源限制

在生產環境中設定資源限制：

```yaml
services:
  goip:
    deploy:
      resources:
        limits:
          cpus: '1'
          memory: 512M
        reservations:
          cpus: '0.5'
          memory: 256M
```

## 📚 相關文件

- [多資料庫使用指南](MULTI_DB_GUIDE.md)
- [API 回應格式說明](API_RESPONSE_FORMAT.md)
- [Docker 測試報告](DOCKER_TEST_REPORT.md)
- [回應格式總結](RESPONSE_FORMAT_SUMMARY.md)

## 🆘 支援

如遇問題，請檢查：

1. **日誌檔案**: `docker-compose logs goip`
2. **健康檢查**: `curl http://localhost:8080/api/v1/health`
3. **提供者列表**: `curl http://localhost:8080/api/v1/providers`
4. **統計資訊**: `curl http://localhost:8080/api/v1/stats`
