# GoIP 多資料庫使用指南

## 概述

GoIP 現在支援多個 IP 地理位置資料庫提供者，並具備智能路由功能。

## 支援的資料庫

1. **MaxMind GeoLite2** (.mmdb)
   - 全球覆蓋率高
   - 海外地區資料準確
   - 免費版本

2. **IPIP.NET** (.ipdb)
   - 中國地區資料詳細
   - 包含運營商資訊
   - 免費版本

## 智能路由

系統會自動根據 IP 歸屬地選擇最佳資料庫：

- **中國 IP** → 優先使用 IPIP → 如果查不到，使用 MaxMind
- **海外 IP** → 優先使用 MaxMind → 如果查不到，使用 IPIP

## 配置方式

### 方式一：YAML 配置（推薦）

創建 `config.yaml`：

```yaml
server:
  port: 8080

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

redis:
  host: localhost
  port: 6379

cache:
  enabled: true
  ttl: 24h

log:
  level: info
  format: console
```

### 方式二：環境變數（向後相容）

```bash
# 單一 MaxMind 資料庫（舊方式）
export MAXMIND_DB_PATH=./data/GeoLite2-City.mmdb
```

## API 端點

### 1. 列出可用提供者

```bash
GET /api/v1/providers
```

**回應：**
```json
{
  "count": 2,
  "providers": ["ipip", "maxmind"]
}
```

### 2. 智能路由查詢（推薦）

自動選擇最佳資料庫：

```bash
GET /api/v1/ip/:ip
```

**範例：**
```bash
# 中國 IP - 自動使用 IPIP
curl http://localhost:8080/api/v1/ip/114.114.114.114

# 海外 IP - 自動使用 MaxMind
curl http://localhost:8080/api/v1/ip/8.8.8.8
```

### 3. 指定提供者查詢

手動指定使用哪個資料庫：

```bash
GET /api/v1/ip/:ip/provider?provider={providerType}
```

**範例：**
```bash
# 強制使用 IPIP 查詢海外 IP
curl "http://localhost:8080/api/v1/ip/8.8.8.8/provider?provider=ipip"

# 強制使用 MaxMind 查詢中國 IP
curl "http://localhost:8080/api/v1/ip/114.114.114.114/provider?provider=maxmind"
```

## Docker 部署

### 建置映像

```bash
docker build -t goip:latest -f build/Dockerfile .
```

### 使用 docker-compose 啟動

```bash
# 測試環境（包含 Redis）
docker-compose -f docker-compose.test.yml up -d

# 查看日誌
docker-compose -f docker-compose.test.yml logs -f goip

# 停止服務
docker-compose -f docker-compose.test.yml down
```

### 執行自動化測試

```bash
./docker-build-and-test.sh
```

## 資料庫檔案準備

### MaxMind GeoLite2

1. 註冊 MaxMind 帳號
2. 下載 GeoLite2-City.mmdb
3. 放置到 `./data/GeoLite2-City.mmdb`

### IPIP.NET

1. 下載免費版 ipipfree.ipdb
2. 放置到 `./data/ipipfree.ipdb`

## 回應格式

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
  "city": {
    "name": "Mountain View",
    "name_zh": "山景城"
  },
  "location": {
    "latitude": 37.386,
    "longitude": -122.0838,
    "time_zone": "America/Los_Angeles"
  },
  "provider": "maxmind",
  "query_time_ms": 5
}
```

## 效能建議

### 1. 啟用快取

```yaml
cache:
  enabled: true
  ttl: 24h
```

快取命中可將查詢時間降至 1-2ms。

### 2. 連接池設定

```yaml
redis:
  pool_size: 10
  min_idle_conns: 5
```

### 3. 批次查詢

```bash
POST /api/v1/ip/batch
Content-Type: application/json

{
  "ips": ["8.8.8.8", "114.114.114.114", "1.1.1.1"]
}
```

## 監控

### 健康檢查

```bash
GET /api/v1/health
```

### 統計資訊

```bash
GET /api/v1/stats
```

### 快取統計

```bash
GET /api/v1/cache/stats
```

## 常見問題

### Q: 如何新增其他資料庫？

A: 在 `config.yaml` 中添加新的 provider：

```yaml
geoip:
  providers:
    - type: ipip
      db_path: ./data/ipipfree.ipdb
      priority: 1
      region: cn

    - type: maxmind
      db_path: ./data/GeoLite2-City.mmdb
      priority: 2
      region: all
```

### Q: 如何調整優先級？

A: 修改 `priority` 值，數字越小優先級越高：

```yaml
- type: ipip
  priority: 1  # 優先級最高

- type: maxmind
  priority: 2  # 次優先級
```

### Q: 智能路由如何判斷中國 IP？

A: 系統內建中國主要 IP 段列表，自動判斷。可在 [multi_provider_repository.go](internal/repository/multi_provider_repository.go:119-165) 中查看和修改。

### Q: 可以完全關閉智能路由嗎？

A: 可以，設定所有 provider 的 `region` 為 `all`，系統將按 `priority` 順序查詢。

## 升級指南

### 從單一 MaxMind 升級到多資料庫

1. 保持現有配置不變（向後相容）
2. 創建 `config.yaml` 添加新提供者
3. 重啟服務
4. 測試新端點
5. 逐步遷移到新配置

## 相關連結

- [測試報告](DOCKER_TEST_REPORT.md)
- [配置範例](config/config.example.yaml)
- [API 文檔](README.md)

## 技術支援

如有問題，請查看：
1. Docker 日誌：`docker-compose logs -f goip`
2. 健康檢查：`curl http://localhost:8080/api/v1/health`
3. 提供者列表：`curl http://localhost:8080/api/v1/providers`
