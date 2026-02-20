.PHONY: help build run test clean docker-build docker-up docker-down docker-logs deps tidy fmt lint vet
.PHONY: docker-deps-up docker-deps-down docker-goip-up docker-goip-down install-tools dev update-db

# 變數
BINARY_NAME=goip
MAIN_PATH=cmd/server/main.go
BUILD_DIR=build
DEPLOY_DIR=deployments

# Docker 相關
DOCKER_IMAGE=$(BINARY_NAME)
DOCKER_TAG=latest

# 預設目標
help: ## 顯示幫助資訊
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

# ============================================================================
# 建置相關
# ============================================================================

build: ## 建置專案
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PATH)
	@echo "Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

build-all: ## 跨平台建置（使用 build.sh）
	@echo "Building for all platforms..."
	@$(BUILD_DIR)/build.sh all

run: ## 運行服務
	@go run $(MAIN_PATH)

clean: ## 清理建置檔案
	@echo "Cleaning..."
	@rm -rf $(BUILD_DIR)/*
	@rm -f coverage.txt coverage.html
	@echo "Clean complete"

# ============================================================================
# 測試相關
# ============================================================================

test: ## 運行測試
	@go test -v -race -coverprofile=coverage.txt -covermode=atomic ./...

test-coverage: test ## 運行測試並生成覆蓋率報告
	@go tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report generated: coverage.html"

# ============================================================================
# 程式碼品質
# ============================================================================

deps: ## 下載依賴
	@echo "Downloading dependencies..."
	@go mod download
	@echo "Dependencies downloaded"

tidy: ## 整理依賴
	@echo "Tidying dependencies..."
	@go mod tidy
	@echo "Dependencies tidied"

fmt: ## 格式化程式碼
	@echo "Formatting code..."
	@go fmt ./...
	@echo "Code formatted"

lint: ## 檢查程式碼
	@echo "Linting code..."
	@golangci-lint run ./...

vet: ## 靜態分析
	@go vet ./...

# ============================================================================
# Docker 建置
# ============================================================================

docker-build: ## 建置 Docker 映像（使用 docker-build.sh）
	@echo "Building Docker image..."
	@$(BUILD_DIR)/docker-build.sh -n $(DOCKER_IMAGE) -v $(DOCKER_TAG)

docker-build-version: ## 建置帶版本的 Docker 映像 (用法: make docker-build-version VERSION=1.0.0)
	@echo "Building Docker image with version..."
	@$(BUILD_DIR)/docker-build.sh -n $(DOCKER_IMAGE) -v $(VERSION)

# ============================================================================
# Docker Compose - Redis
# ============================================================================

docker-redis-up: ## 啟動 Redis 服務
	@echo "Starting Redis..."
	@docker-compose -f $(DEPLOY_DIR)/redis/docker-compose.yml up -d
	@echo "Redis started"

docker-redis-down: ## 停止 Redis 服務
	@echo "Stopping Redis..."
	@docker-compose -f $(DEPLOY_DIR)/redis/docker-compose.yml down
	@echo "Redis stopped"

docker-redis-logs: ## 查看 Redis 日誌
	@docker-compose -f $(DEPLOY_DIR)/redis/docker-compose.yml logs -f

# ============================================================================
# Docker Compose - GoIP
# ============================================================================

docker-goip-up: ## 啟動 GoIP 服務
	@echo "Starting GoIP..."
	@docker-compose -f $(DEPLOY_DIR)/goip/docker-compose.yml up -d
	@echo "GoIP started"

docker-goip-down: ## 停止 GoIP 服務
	@echo "Stopping GoIP..."
	@docker-compose -f $(DEPLOY_DIR)/goip/docker-compose.yml down
	@echo "GoIP stopped"

docker-goip-logs: ## 查看 GoIP 日誌
	@docker-compose -f $(DEPLOY_DIR)/goip/docker-compose.yml logs -f

docker-goip-restart: ## 重啟 GoIP 服務
	@$(MAKE) docker-goip-down
	@$(MAKE) docker-goip-up

# ============================================================================
# Docker Compose - 完整部署
# ============================================================================

docker-up: docker-redis-up docker-goip-up ## 啟動所有服務（Redis + GoIP）

docker-down: docker-goip-down docker-redis-down ## 停止所有服務

docker-restart: docker-down docker-up ## 重啟所有服務

docker-logs: ## 查看所有服務日誌
	@echo "=== Redis Logs ==="
	@docker-compose -f $(DEPLOY_DIR)/redis/docker-compose.yml logs --tail=50
	@echo ""
	@echo "=== GoIP Logs ==="
	@docker-compose -f $(DEPLOY_DIR)/goip/docker-compose.yml logs --tail=50

docker-ps: ## 查看所有容器狀態
	@echo "=== Redis ==="
	@docker-compose -f $(DEPLOY_DIR)/redis/docker-compose.yml ps
	@echo ""
	@echo "=== GoIP ==="
	@docker-compose -f $(DEPLOY_DIR)/goip/docker-compose.yml ps

# ============================================================================
# 開發相關
# ============================================================================

dev: ## 開發模式運行（with hot reload）
	@air -c .air.toml

install-tools: ## 安裝開發工具
	@echo "Installing development tools..."
	@go install github.com/cosmtrek/air@latest
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@echo "Tools installed"

# ============================================================================
# 資料庫相關
# ============================================================================

update-db: ## 更新 MaxMind 資料庫（需手動下載）
	@echo "Please download GeoLite2-City.mmdb from MaxMind and place it in data/"
	@echo "https://dev.maxmind.com/geoip/geolite2-free-geolocation-data"

# ============================================================================
# 快速啟動指令
# ============================================================================

quick-start: docker-redis-up ## 快速啟動（Redis + 本地運行 GoIP）
	@echo "Redis started. Now run 'make run' to start GoIP locally"

full-deploy: docker-build docker-up ## 完整部署（建置 + 啟動所有服務）
	@echo "Full deployment complete!"
	@echo "GoIP API: http://localhost:8080"
	@echo "Health check: curl http://localhost:8080/api/v1/health"
