# GoIP 部署指南

## 使用 GitHub Container Registry

本專案的 Docker image 會自動建置並推送到 GitHub Container Registry (ghcr.io)。

### Image 位置
```
ghcr.io/shengjhe/goip:latest
```

### 可用的 Tags
- `latest` - 最新的 main 分支建置
- `main` - main 分支
- `sha-<commit>` - 特定 commit 的建置
- `v*.*.*` - 版本標籤（如果有打 git tag）

## 快速部署

### 1. 拉取最新 image
```bash
docker pull ghcr.io/shengjhe/goip:latest
```

### 2. 準備資料庫檔案
確保 `../../data/` 目錄包含：
- `GeoLite2-City.mmdb` (MaxMind)
- `ipipfree.ipdb` (IPIP.NET, 選用)

### 3. 啟動服務
```bash
# 確保 Redis 已啟動
cd ../redis && ./start.sh

# 啟動 GoIP
cd ../goip && ./start.sh
```

或直接使用 docker-compose：
```bash
docker-compose up -d
```

## 環境變數

| 變數 | 預設值 | 說明 |
|------|--------|------|
| `SERVER_PORT` | `8080` | HTTP 服務端口 |
| `REDIS_HOST` | `redis` | Redis 主機位址 |
| `REDIS_PORT` | `6379` | Redis 端口 |
| `CACHE_ENABLED` | `true` | 啟用快取 |
| `CACHE_TTL` | `24h` | 快取過期時間 |
| `LOG_LEVEL` | `info` | 日誌級別 |
| `LOG_FORMAT` | `json` | 日誌格式 (json/console) |
| `FLUSH_DNS` | `true` | 啟動時清空快取 |

## 私有倉庫認證

如果 GitHub 倉庫是私有的，需要先登入：

```bash
# 建立 Personal Access Token (PAT) 具有 read:packages 權限
echo $GITHUB_TOKEN | docker login ghcr.io -u USERNAME --password-stdin
```

## 健康檢查

```bash
curl http://localhost:8080/healthz
# 應返回: OK
```

## 更新 Image

```bash
# 停止服務
docker-compose down

# 拉取最新 image
docker-compose pull

# 重新啟動
docker-compose up -d
```

## 故障排除

### Image 拉取失敗
1. 確認網路連線正常
2. 確認 GitHub Container Registry 可訪問
3. 如果是私有倉庫，確認已正確登入

### 服務無法啟動
```bash
# 查看日誌
docker logs goip

# 檢查 Redis 連接
docker exec goip wget -q -O - http://localhost:8080/api/v1/health
```

## CI/CD 自動化

每次推送到 main 分支時，GitHub Actions 會自動：
1. 執行測試
2. 建置 Docker image
3. 推送到 ghcr.io/shengjhe/goip

查看建置狀態：
- https://github.com/shengjhe/goip/actions
- https://github.com/shengjhe/goip/pkgs/container/goip
