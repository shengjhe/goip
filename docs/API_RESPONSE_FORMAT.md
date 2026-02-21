# GoIP API 回應格式說明

## 統一回應格式

所有 IP 查詢都會返回以下統一的 JSON 格式：

### 必填欄位

| 欄位 | 類型 | 說明 | 範例 |
|------|------|------|------|
| `ip` | string | 查詢的 IP 地址 | "8.8.8.8" |
| `country` | object | 國家資訊 | {"iso_code": "US", "name": "United States"} |
| `city` | object | 城市資訊 | {"name": "Mountain View", "name_zh": "山景城"} |
| `provider` | string | 資料來源 | "maxmind" 或 "ipip" |
| `query_time_ms` | integer | 查詢耗時（毫秒） | 5 |

### 選填欄位

| 欄位 | 類型 | 說明 | 範例 |
|------|------|------|------|
| `continent` | object | 大洲資訊 | {"code": "NA", "name": "North America"} |
| `location` | object | 經緯度資訊 | {"latitude": 37.751, "longitude": -97.822} |
| `cached_at` | string | 快取時間 | "2026-02-21T09:26:39Z" |

## 完整範例

### 範例 1: 海外 IP (MaxMind)

**請求：**
```bash
GET /api/v1/ip/8.8.8.8
```

**回應：**
```json
{
  "ip": "8.8.8.8",
  "country": {
    "iso_code": "US",
    "name": "United States",
    "name_zh": "美国"
  },
  "city": {
    "name": "",
    "name_zh": "",
    "postal_code": ""
  },
  "provider": "maxmind",
  "continent": {
    "code": "NA",
    "name": "North America"
  },
  "location": {
    "latitude": 37.751,
    "longitude": -97.822,
    "time_zone": "America/Chicago"
  },
  "query_time_ms": 2
}
```

### 範例 2: 中國 IP - IPIP 資料庫

**請求：**
```bash
GET /api/v1/ip/42.120.160.1
```

**回應：**
```json
{
  "ip": "42.120.160.1",
  "country": {
    "iso_code": "",
    "name": "中国",
    "name_zh": ""
  },
  "city": {
    "name": "杭州",
    "name_zh": "浙江杭州",
    "postal_code": ""
  },
  "provider": "ipip",
  "continent": {
    "code": "",
    "name": ""
  },
  "query_time_ms": 3
}
```

### 範例 3: 中國 IP - MaxMind 資料庫

**請求：**
```bash
GET /api/v1/ip/42.120.160.1/provider?provider=maxmind
```

**回應：**
```json
{
  "ip": "42.120.160.1",
  "country": {
    "iso_code": "CN",
    "name": "China",
    "name_zh": "中国"
  },
  "city": {
    "name": "Hangzhou",
    "name_zh": "杭州",
    "postal_code": ""
  },
  "provider": "maxmind",
  "continent": {
    "code": "AS",
    "name": "Asia"
  },
  "location": {
    "latitude": 30.2943,
    "longitude": 120.1663,
    "time_zone": "Asia/Shanghai"
  },
  "query_time_ms": 1
}
```

## 欄位詳細說明

### country (國家資訊)

```json
{
  "iso_code": "US",    // ISO 3166-1 alpha-2 國碼，可能為空
  "name": "United States",    // 英文名稱
  "name_zh": "美国"   // 中文名稱（如果有）
}
```

### city (城市資訊)

```json
{
  "name": "Hangzhou",         // 英文名稱，可能為空
  "name_zh": "杭州",          // 中文名稱，可能為空
  "postal_code": "310000"     // 郵遞區號，可能為空
}
```

**注意：**
- IPIP 資料庫的 `name_zh` 可能包含省份+城市（如 "浙江杭州"）
- MaxMind 資料庫的 `name` 是標準英文名稱
- 如果該資料庫沒有城市資訊，所有欄位會是空字串

### continent (大洲資訊)

```json
{
  "code": "AS",    // 大洲代碼 (AF, AN, AS, EU, NA, OC, SA)
  "name": "Asia"   // 大洲名稱
}
```

**注意：** IPIP 免費版通常不提供大洲資訊，欄位會是空字串。

### location (位置資訊)

```json
{
  "latitude": 30.2943,              // 緯度
  "longitude": 120.1663,            // 經度
  "time_zone": "Asia/Shanghai"      // 時區
}
```

**注意：**
- IPIP 免費版通常不提供經緯度資訊
- 如果沒有位置資訊，此欄位不會出現在回應中（omitempty）

## 資料庫特性比較

| 特性 | MaxMind | IPIP |
|------|---------|------|
| 全球覆蓋 | ✅ 優秀 | ⚠️ 一般 |
| 中國地區 | ⚠️ 一般 | ✅ 優秀 |
| 國家資訊 | ✅ | ✅ |
| 城市資訊 | ✅ | ✅ |
| 經緯度 | ✅ | ❌ 免費版無 |
| 時區 | ✅ | ❌ 免費版無 |
| 大洲資訊 | ✅ | ❌ 免費版無 |
| ISO 國碼 | ✅ | ❌ 免費版無 |

## 使用建議

### 智能路由（推薦）

讓系統自動選擇最佳資料庫：

```bash
GET /api/v1/ip/:ip
```

- 中國 IP → 自動使用 IPIP（城市資訊更準確）
- 海外 IP → 自動使用 MaxMind（全球覆蓋更完整）

### 指定資料庫

如果需要特定資料庫的資訊（例如需要經緯度）：

```bash
GET /api/v1/ip/:ip/provider?provider=maxmind
```

可用的 provider：
- `maxmind` - MaxMind GeoLite2 資料庫
- `ipip` - IPIP.NET 資料庫

## 空值處理

如果某個欄位沒有資料，會返回：
- **字串欄位**：空字串 `""`
- **物件欄位**：空物件 `{}`（必填欄位）或不存在（選填欄位）

範例（無城市資訊的 IP）：
```json
{
  "ip": "1.1.1.1",
  "country": {
    "iso_code": "AU",
    "name": "Australia",
    "name_zh": "澳大利亚"
  },
  "city": {
    "name": "",
    "name_zh": "",
    "postal_code": ""
  },
  "provider": "maxmind",
  "continent": {
    "code": "OC",
    "name": "Oceania"
  },
  "query_time_ms": 2
}
```

## 錯誤回應

查詢失敗時的錯誤格式：

```json
{
  "error": "IP not found in database",
  "code": "IP_NOT_FOUND",
  "timestamp": "2026-02-21T09:26:39Z"
}
```

常見錯誤代碼：
- `INVALID_IP` - 無效的 IP 地址格式
- `IP_NOT_FOUND` - IP 在資料庫中找不到
- `PROVIDER_NOT_FOUND` - 指定的提供者不存在
- `DATABASE_ERROR` - 資料庫讀取錯誤
