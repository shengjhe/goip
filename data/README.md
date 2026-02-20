# GeoLite2 數據庫

此目錄包含 MaxMind GeoLite2 城市數據庫文件，用於 IP 地理位置查詢。

## 文件說明

- `GeoLite2-City.mmdb` - MaxMind GeoLite2 城市數據庫（使用 Git LFS 管理）
- 當前版本：2026-02-17

## 使用方式

在代碼中使用時，請指定數據庫路徑：

```go
repo, err := repository.NewMaxMindRepository("data/GeoLite2-City.mmdb")
if err != nil {
    log.Fatal(err)
}
defer repo.Close()
```

## 更新數據庫

GeoLite2 數據庫會定期更新。要更新到最新版本：

1. 從 MaxMind 下載最新的 GeoLite2-City.mmdb
2. 替換此目錄中的文件
3. 提交更改（Git LFS 會自動處理大型文件）

## 授權

此數據庫由 MaxMind 提供，遵循 [GeoLite2 End User License Agreement](https://www.maxmind.com/en/geolite2/eula)。
