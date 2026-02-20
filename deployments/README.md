# 部署配置

GoIP 服務的 Docker 部署配置目錄。

## 目錄結構

```
deployments/
├── goip/                    # GoIP 主服務部署配置
│   ├── Dockerfile           # GoIP Docker 映像建置檔案
│   ├── docker-compose.yml   # GoIP 服務編排配置
│   └── README.md            # GoIP 部署說明
└── dependencies/            # 依賴服務部署配置
    ├── docker-compose.yml   # 依賴服務編排配置（Redis）
    ├── redis.conf           # Redis 配置檔案
    ├── .env.example         # 環境變數範例
    └── README.md            # 依賴服務部署說明
```

## 部署方式

### 方式一：使用 Makefile（推薦）

#### 完整部署（建置 + 啟動所有服務）

```bash
# 建置映像並啟動所有服務
make full-deploy

# 檢查服務狀態
make docker-ps

# 查看日誌
make docker-logs
```

#### 分步部署

```bash
# 1. 啟動依賴服務（Redis）
make docker-deps-up

# 2. 建置 GoIP Docker 映像
make docker-build

# 3. 啟動 GoIP 服務
make docker-goip-up
```

#### 本地開發模式

```bash
# 只啟動依賴服務，本地運行 GoIP
make quick-start

# 在另一個終端運行 GoIP
make run
```

### 方式二：手動使用 Docker Compose

#### 啟動所有服務

```bash
# 1. 啟動依賴服務
docker-compose -f deployments/dependencies/docker-compose.yml up -d

# 2. 建置並啟動 GoIP 服務
docker-compose -f deployments/goip/docker-compose.yml up -d --build
```

#### 停止所有服務

```bash
# 停止 GoIP 服務
docker-compose -f deployments/goip/docker-compose.yml down

# 停止依賴服務
docker-compose -f deployments/dependencies/docker-compose.yml down
```

## 常用命令

### 服務管理

```bash
# 啟動所有服務
make docker-up

# 停止所有服務
make docker-down

# 重啟所有服務
make docker-restart

# 查看服務狀態
make docker-ps
```

### 日誌查看

```bash
# 查看所有日誌
make docker-logs

# 只查看 GoIP 日誌
make docker-goip-logs

# 只查看依賴服務日誌
make docker-deps-logs
```

### 單獨管理依賴服務

```bash
# 啟動依賴服務
make docker-deps-up

# 停止依賴服務
make docker-deps-down

# 查看依賴服務日誌
make docker-deps-logs
```

### 單獨管理 GoIP 服務

```bash
# 啟動 GoIP 服務
make docker-goip-up

# 停止 GoIP 服務
make docker-goip-down

# 重啟 GoIP 服務
make docker-goip-restart

# 查看 GoIP 日誌
make docker-goip-logs
```

## 詳細文檔

- [GoIP 服務部署說明](goip/README.md) - GoIP 主服務的詳細部署配置
- [依賴服務部署說明](dependencies/README.md) - Redis 等依賴服務的配置和管理

## 相關資源

- [主要 README](../README.md)
- [設計文檔](../DESIGN.md)
- [建置腳本說明](../scripts/README.md)
