# 多階段建置
# Stage 1: 建置階段
FROM golang:1.23-alpine AS builder

# 安裝必要工具
RUN apk add --no-cache git make

# 設定工作目錄
WORKDIR /build

# 複製 go mod 檔案
COPY go.mod go.sum ./

# 下載依賴
RUN go mod download

# 複製原始碼
COPY . .

# 建置二進制檔案
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -o /build/bin/goip \
    ./cmd/server/main.go

# Stage 2: 運行階段
FROM alpine:latest

# 安裝 CA 證書和 wget（用於健康檢查）
RUN apk --no-cache add ca-certificates wget

# 建立非 root 用戶
RUN addgroup -g 1000 goip && \
    adduser -D -u 1000 -G goip goip

# 設定工作目錄
WORKDIR /app

# 從建置階段複製二進制檔案
COPY --from=builder /build/bin/goip /app/goip

# 複製資料目錄（如果需要）
COPY --chown=goip:goip data /app/data

# 切換到非 root 用戶
USER goip

# 暴露端口
EXPOSE 8080

# 健康檢查
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget -q --spider http://localhost:8080/api/v1/health || exit 1

# 運行服務
ENTRYPOINT ["/app/goip"]
