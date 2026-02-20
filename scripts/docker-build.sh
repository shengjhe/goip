#!/bin/bash

set -e

# 顏色定義
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# 預設值
IMAGE_NAME="goip"
VERSION="latest"
BUILD_CONTEXT="."
DOCKERFILE="deployments/goip/Dockerfile"
PLATFORM=""

# 輸出帶顏色的訊息
info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

error() {
    echo -e "${RED}[ERROR]${NC} $1"
    exit 1
}

# 使用說明
usage() {
    cat << EOF
使用方法: $0 [選項]

選項:
    -n, --name NAME         Docker 映像名稱 (預設: goip)
    -v, --version VERSION   映像版本標籤 (預設: latest)
    -p, --platform PLATFORM 目標平台 (例如: linux/amd64,linux/arm64)
    -h, --help             顯示此幫助訊息

範例:
    # 基本建置
    $0

    # 指定版本
    $0 -v 1.0.0

    # 多平台建置
    $0 -p linux/amd64,linux/arm64 -v 1.0.0

    # 自定義映像名稱
    $0 -n myorg/goip -v dev
EOF
    exit 0
}

# 解析參數
while [[ $# -gt 0 ]]; do
    case $1 in
        -n|--name)
            IMAGE_NAME="$2"
            shift 2
            ;;
        -v|--version)
            VERSION="$2"
            shift 2
            ;;
        -p|--platform)
            PLATFORM="$2"
            shift 2
            ;;
        -h|--help)
            usage
            ;;
        *)
            error "未知選項: $1"
            ;;
    esac
done

# 檢查 Docker
if ! command -v docker &> /dev/null; then
    error "未安裝 Docker，請先安裝 Docker"
fi

# 檢查 Dockerfile 是否存在
if [ ! -f "$DOCKERFILE" ]; then
    error "找不到 Dockerfile: $DOCKERFILE"
fi

# 建置資訊
info "開始建置 Docker 映像..."
info "映像名稱: $IMAGE_NAME:$VERSION"
info "Dockerfile: $DOCKERFILE"
info "建置上下文: $BUILD_CONTEXT"

# 建置命令
BUILD_CMD="docker build"

# 添加平台參數
if [ -n "$PLATFORM" ]; then
    info "目標平台: $PLATFORM"
    BUILD_CMD="$BUILD_CMD --platform $PLATFORM"
fi

# 添加標籤
BUILD_CMD="$BUILD_CMD -t $IMAGE_NAME:$VERSION"

# 如果版本不是 latest，同時打上 latest 標籤
if [ "$VERSION" != "latest" ]; then
    BUILD_CMD="$BUILD_CMD -t $IMAGE_NAME:latest"
fi

# 添加建置參數
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_TIME=$(date -u '+%Y-%m-%d_%H:%M:%S')

BUILD_CMD="$BUILD_CMD --build-arg VERSION=$VERSION"
BUILD_CMD="$BUILD_CMD --build-arg GIT_COMMIT=$GIT_COMMIT"
BUILD_CMD="$BUILD_CMD --build-arg BUILD_TIME=$BUILD_TIME"

# 添加 Dockerfile 和上下文
BUILD_CMD="$BUILD_CMD -f $DOCKERFILE $BUILD_CONTEXT"

# 執行建置
info "執行: $BUILD_CMD"
if eval "$BUILD_CMD"; then
    info "建置成功！"
    echo ""
    info "映像標籤:"
    echo "  - $IMAGE_NAME:$VERSION"
    [ "$VERSION" != "latest" ] && echo "  - $IMAGE_NAME:latest"
    echo ""
    info "下一步:"
    echo "  # 執行容器"
    echo "  docker run -d -p 8080:8080 --env-file .env $IMAGE_NAME:$VERSION"
    echo ""
    echo "  # 推送到 Registry"
    echo "  docker push $IMAGE_NAME:$VERSION"
else
    error "建置失敗"
fi
