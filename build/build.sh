#!/bin/bash

# GoIP 建置腳本
# 支援跨平台編譯

set -e

# 變數設定
APP_NAME="goip"
BUILD_DIR="build"
MAIN_PATH="cmd/server/main.go"
VERSION=${VERSION:-"1.0.0"}
COMMIT_HASH=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')

# LDFLAGS
LDFLAGS="-w -s -X main.Version=${VERSION} -X main.CommitHash=${COMMIT_HASH} -X main.BuildTime=${BUILD_TIME}"

# 顏色輸出
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== GoIP 建置腳本 ===${NC}"
echo "Version: ${VERSION}"
echo "Commit: ${COMMIT_HASH}"
echo "Build Time: ${BUILD_TIME}"
echo ""

# 建立 build 目錄
mkdir -p ${BUILD_DIR}

# 函數: 建置單一平台
build_platform() {
    local os=$1
    local arch=$2
    local output="${BUILD_DIR}/${APP_NAME}-${os}-${arch}"

    if [ "$os" = "windows" ]; then
        output="${output}.exe"
    fi

    echo -e "${YELLOW}Building for ${os}/${arch}...${NC}"

    CGO_ENABLED=0 GOOS=${os} GOARCH=${arch} go build \
        -ldflags="${LDFLAGS}" \
        -o ${output} \
        ${MAIN_PATH}

    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✓ Built: ${output}${NC}"
    else
        echo "✗ Build failed for ${os}/${arch}"
        return 1
    fi
}

# 判斷建置模式
if [ "$1" = "all" ]; then
    # 建置所有平台
    echo "Building for all platforms..."
    build_platform "linux" "amd64"
    build_platform "linux" "arm64"
    build_platform "darwin" "amd64"
    build_platform "darwin" "arm64"
    build_platform "windows" "amd64"
elif [ "$1" = "linux" ]; then
    build_platform "linux" "amd64"
elif [ "$1" = "darwin" ] || [ "$1" = "mac" ]; then
    build_platform "darwin" "amd64"
    build_platform "darwin" "arm64"
elif [ "$1" = "windows" ]; then
    build_platform "windows" "amd64"
else
    # 預設：建置當前平台
    echo "Building for current platform..."
    go build -ldflags="${LDFLAGS}" -o ${BUILD_DIR}/${APP_NAME} ${MAIN_PATH}
    echo -e "${GREEN}✓ Built: ${BUILD_DIR}/${APP_NAME}${NC}"
fi

echo ""
echo -e "${GREEN}=== Build Complete ===${NC}"
ls -lh ${BUILD_DIR}/
