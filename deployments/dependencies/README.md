# GoIP 依賴服務部署

GoIP 服務所需的外部依賴服務部署配置。

## 包含的服務

- **Redis**: 分散式快取和限流儲存

## 檔案說明

- **docker-compose.yml**: 依賴服務的 Docker Compose 編排配置
- **redis.conf**: Redis 伺服器配置檔案
- **.env.example**: 環境變數範例（可選）

## 快速啟動

### 1. 啟動所有依賴服務

```bash
# 在此目錄下
docker-compose up -d

# 或從專案根目錄
docker-compose -f deployments/dependencies/docker-compose.yml up -d
```

### 2. 驗證服務狀態

```bash
# 檢查容器狀態
docker-compose ps

# 檢查 Redis
docker exec -it goip-redis redis-cli ping
# 應返回: PONG
```

### 3. 查看日誌

```bash
docker-compose logs -f redis
```

## Redis 配置

### 基本配置

Redis 配置檔案位於 `redis.conf`，主要配置項：

| 配置項 | 值 | 說明 |
|--------|-----|------|
| maxmemory | 512mb | 最大記憶體使用量 |
| maxmemory-policy | allkeys-lru | 記憶體淘汰策略 |
| appendonly | yes | 啟用 AOF 持久化 |
| save | 3600 1 300 100 60 10000 | RDB 快照策略 |

### 安全性配置

**重要：生產環境必須設置密碼！**

1. 編輯 `redis.conf`，取消註解並設置強密碼：
   ```conf
   requirepass your_strong_password_here
   ```

2. 更新 `.env` 檔案：
   ```bash
   REDIS_PASSWORD=your_strong_password_here
   ```

3. 重啟 Redis：
   ```bash
   docker-compose restart redis
   ```

### 禁用的危險命令

為提高安全性，以下命令已被禁用：
- `FLUSHDB` - 刪除當前資料庫
- `FLUSHALL` - 刪除所有資料庫
- `CONFIG` - 修改配置
- `KEYS` - 列出所有鍵（影響效能）

### 持久化策略

Redis 使用雙重持久化：

1. **RDB (快照)**
   - 每小時至少 1 次變更時保存
   - 每 5 分鐘至少 100 次變更時保存
   - 每 1 分鐘至少 10000 次變更時保存

2. **AOF (追加日誌)**
   - 每秒同步一次
   - 提供更好的資料安全性

## 網路配置

所有依賴服務使用 `goip-network` 橋接網路，與 GoIP 主服務通訊。

### 埠映射

- Redis: `127.0.0.1:6379:6379` (只綁定本地，提高安全性)

如需從外部訪問（開發環境），修改 `docker-compose.yml`：
```yaml
ports:
  - "6379:6379"  # 暴露到所有介面（不建議用於生產環境）
```

## 資料備份

### 手動備份

```bash
# 建立 RDB 快照
docker exec goip-redis redis-cli BGSAVE

# 複製資料檔案
docker cp goip-redis:/data/dump.rdb ./backup/redis-$(date +%Y%m%d).rdb
```

### 自動備份

建議使用 cron job 定期備份：

```bash
# 添加到 crontab
0 2 * * * docker exec goip-redis redis-cli BGSAVE
0 3 * * * docker cp goip-redis:/data/dump.rdb /backup/redis-$(date +\%Y\%m\%d).rdb
```

## 效能監控

### 監控指令

```bash
# 即時監控
docker exec -it goip-redis redis-cli MONITOR

# 查看統計資訊
docker exec -it goip-redis redis-cli INFO stats

# 查看記憶體使用
docker exec -it goip-redis redis-cli INFO memory

# 查看慢查詢日誌
docker exec -it goip-redis redis-cli SLOWLOG GET 10
```

### 效能調校

如遇效能問題，可調整以下配置：

1. **增加記憶體**（`redis.conf`）：
   ```conf
   maxmemory 1gb
   ```

2. **調整持久化策略**（降低 I/O）：
   ```conf
   save 3600 1
   appendfsync no
   ```

3. **調整連接數**（`redis.conf`）：
   ```conf
   maxclients 10000
   ```

## 故障排除

### Redis 無法啟動

1. 檢查日誌：
   ```bash
   docker-compose logs redis
   ```

2. 檢查配置檔案語法：
   ```bash
   docker exec goip-redis redis-server /usr/local/etc/redis/redis.conf --test-memory 1
   ```

### 記憶體不足

1. 查看記憶體使用：
   ```bash
   docker exec goip-redis redis-cli INFO memory
   ```

2. 清理過期鍵：
   ```bash
   docker exec goip-redis redis-cli --scan --pattern "cache:*" | xargs -L 100 docker exec -i goip-redis redis-cli DEL
   ```

### 連接被拒絕

1. 檢查網路：
   ```bash
   docker network inspect goip-network
   ```

2. 檢查防火牆規則

3. 驗證密碼設置（如有）

## 清理資料

```bash
# 停止並移除容器（保留資料）
docker-compose down

# 停止並移除容器和資料卷
docker-compose down -v

# 清理網路
docker network rm goip-network
```

## 擴展配置

### 添加更多依賴服務

在 `docker-compose.yml` 中添加新服務：

```yaml
services:
  redis:
    # ... existing config ...

  # 例如：添加 PostgreSQL
  postgres:
    image: postgres:15-alpine
    container_name: goip-postgres
    environment:
      POSTGRES_DB: goip
      POSTGRES_USER: goip
      POSTGRES_PASSWORD: ${POSTGRES_PASSWORD}
    volumes:
      - postgres-data:/var/lib/postgresql/data
    networks:
      - goip-network

volumes:
  redis-data:
  postgres-data:
```

## 相關資源

- [Redis 官方文檔](https://redis.io/docs/)
- [Redis 配置參考](https://redis.io/docs/management/config/)
- [Redis 安全性指南](https://redis.io/docs/management/security/)
