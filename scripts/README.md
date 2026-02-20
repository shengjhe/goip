# Scripts Directory

輔助腳本目錄，包含建置和部署相關的自動化腳本。

## 現有腳本

### build.sh

跨平台編譯腳本，支援多個作業系統和架構。

**用法：**

```bash
# 編譯當前平台
./scripts/build.sh

# 編譯特定平台
./scripts/build.sh linux amd64

# 編譯所有支援的平台
./scripts/build.sh all

# 顯示幫助
./scripts/build.sh help
```

**支援的平台：**
- Linux (amd64, arm64)
- Darwin/macOS (amd64, arm64)
- Windows (amd64, arm64)

**產出目錄：** `build/`

---

### docker-build.sh

Docker 映像建置腳本。

**用法：**

```bash
# 基本建置
./scripts/docker-build.sh

# 指定版本
./scripts/docker-build.sh -v 1.0.0

# 多平台建置
./scripts/docker-build.sh -p linux/amd64,linux/arm64 -v 1.0.0

# 顯示幫助
./scripts/docker-build.sh -h
```

**選項：**

| 選項 | 說明 | 預設值 |
|------|------|--------|
| -n, --name | Docker 映像名稱 | goip |
| -v, --version | 映像版本標籤 | latest |
| -p, --platform | 目標平台 | - |
| -h, --help | 顯示幫助訊息 | - |

---

## 透過 Makefile 使用

```bash
# 編譯
make build          # 當前平台
make build-all      # 所有平台

# Docker 建置
make docker-build                      # 基本建置
make docker-build-version VERSION=1.0.0  # 指定版本
```

---

## 腳本規範

添加新腳本時請遵循：

1. 使用 `#!/bin/bash` shebang
2. 添加 `set -e` 錯誤處理
3. 提供 `--help` 選項
4. 使用顏色輸出提升可讀性
5. 檢查前置條件（工具是否安裝）
6. 提供清晰的錯誤訊息
7. 給予執行權限 `chmod +x`

---

## 相關資源

- [Makefile](../Makefile)
- [Deployments](../deployments/)
- [主要 README](../README.md)
