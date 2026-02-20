# Build Directory

建置相關檔案和產物輸出目錄。

## 目錄內容

```
build/
├── Dockerfile          # Docker 映像建置檔案
├── .dockerignore      # Docker 建置忽略檔案
├── build.sh           # 跨平台編譯腳本
├── docker-build.sh    # Docker 映像建置腳本
├── .gitkeep           # Git 目錄占位符
├── README.md          # 本檔案
└── goip*              # 編譯產物（被 gitignore）
```

## 編譯二進制檔案

### 使用 Makefile

```bash
# 編譯當前平台
make build

# 跨平台編譯
make build-all
```

### 使用建置腳本

```bash
# 編譯當前平台
./build/build.sh

# 編譯特定平台
./build/build.sh linux amd64

# 編譯所有平台
./build/build.sh all
```

### 產出檔案

編譯後會產生以下檔案：

```
build/
├── goip                    # 預設編譯（當前平台）
├── goip-linux-amd64        # Linux AMD64
├── goip-linux-arm64        # Linux ARM64
├── goip-darwin-amd64       # macOS AMD64 (Intel)
├── goip-darwin-arm64       # macOS ARM64 (Apple Silicon)
├── goip-windows-amd64.exe  # Windows AMD64
└── goip-windows-arm64.exe  # Windows ARM64
```

## 建置 Docker 映像

### 使用 Makefile（推薦）

```bash
# 基本建置
make docker-build

# 建置指定版本
make docker-build-version VERSION=1.0.0
```

### 使用建置腳本

```bash
# 基本建置
./build/docker-build.sh

# 指定版本和映像名稱
./build/docker-build.sh -n myorg/goip -v 1.0.0

# 多平台建置
./build/docker-build.sh -p linux/amd64,linux/arm64 -v 1.0.0
```

### Dockerfile 說明

位於 `build/Dockerfile`，特點：

- **多階段建置**: 減少最終映像大小
- **Alpine Linux**: 輕量級基礎映像
- **非 root 用戶**: 安全性最佳實踐
- **健康檢查**: 內建健康檢查端點
- **靜態連結**: 無外部依賴

## 執行編譯產物

### 本地執行

```bash
# Linux/macOS
./build/goip

# Windows
build\goip-windows-amd64.exe
```

### 配置環境變數

複製 `.env.example` 為 `.env`：

```bash
cp .env.example .env
# 編輯 .env 設定必要的環境變數
```

## 清理

```bash
# 清理編譯產物
make clean

# 清理 Docker 映像
docker rmi goip:latest
```

## 相關資源

- [部署配置](../deployments/) - Docker Compose 部署
- [主要 README](../README.md)
