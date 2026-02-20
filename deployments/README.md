# Deployments

部署配置目錄，包含 GoIP 服務和 Redis 的 Docker Compose 配置。

## 目錄結構

```
deployments/
├── goip/                   # GoIP 服務部署
│   ├── docker-compose.yml
│   ├── start.sh           # 啟動腳本
│   └── stop.sh            # 停止腳本
└── redis/                  # Redis 服務部署
    ├── docker-compose.yml
    ├── redis.conf
    ├── start.sh           # 啟動腳本
    └── stop.sh            # 停止腳本
```

## 快速啟動

### 方式一：使用 Makefile（推薦）

```bash
# 啟動所有服務
make docker-up

# 停止所有服務
make docker-down
```

### 方式二：使用啟動腳本

```bash
# 1. 啟動 Redis
cd deployments/redis
./start.sh

# 2. 啟動 GoIP
cd ../goip
./start.sh
```

### 方式三：手動啟動

```bash
# 1. 啟動 Redis
docker-compose -f deployments/redis/docker-compose.yml up -d

# 2. 啟動 GoIP
docker-compose -f deployments/goip/docker-compose.yml up -d
```

## 服務說明

### GoIP 服務
- 埠: 8080
- 映像: goip:latest（需先建置，見 [build/](../build/)）
- 網路: goip-network
- 腳本:
  - `./deployments/goip/start.sh` - 啟動服務
  - `./deployments/goip/stop.sh` - 停止服務

### Redis 服務
- 埠: 127.0.0.1:6379（只綁定本地）
- 配置: redis.conf
- 持久化: Docker volume `redis-data`
- 網路: goip-network
- 腳本:
  - `./deployments/redis/start.sh` - 啟動服務
  - `./deployments/redis/stop.sh` - 停止服務

## 建置映像

在部署前需先建置 GoIP 映像：

```bash
# 使用 Makefile
make docker-build

# 或使用建置腳本
./build/docker-build.sh
```

詳見 [build/README.md](../build/README.md)

## 環境變數

可通過環境變數覆蓋配置：

```bash
export REDIS_HOST=redis
export REDIS_PASSWORD=your_password
export LOG_LEVEL=debug
```

## 網路

所有服務使用共享網路 `goip-network`，在啟動 Redis 時會自動建立。
