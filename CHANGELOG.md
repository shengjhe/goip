# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- 🌐 多資料庫支援架構
  - 新增 IPIP.NET 資料庫整合
  - 新增統一的 GeoIPRepository 介面
  - 新增 MultiProviderRepository 智能路由管理器
- 🎯 智能路由功能
  - 自動識別 IP 歸屬國家
  - 中國大陸 IP 優先使用 IPIP（中文城市資訊詳細）
  - 其他國家（含台港澳）優先使用 MaxMind（準確性高）
- 📋 API 端點擴充
  - `/api/v1/ip/:ip/provider?provider=xxx` - 指定資料庫查詢
  - `/api/v1/providers` - 列出可用的資料庫提供者
- 📝 統一回應格式
  - 必填欄位：`ip`, `country`, `city`, `provider`
  - 選填欄位：`continent`, `location`（只在有資料時出現）
- 📚 文件更新
  - 新增 CLAUDE.md 專案開發指南
  - 新增 docs/ 目錄存放技術文件

### Changed
- 🔄 配置結構調整
  - 支援多提供者配置（`geoip.providers`）
  - 保留向後相容的單一 MaxMind 配置
- 🏗️ 架構重構
  - Repository 層採用統一介面
  - Service 層透過 MultiProviderRepository 管理多資料庫
- 📊 回應格式優化
  - `continent` 和 `location` 改為指標類型
  - 使用 `omitempty` 標籤避免空物件

### Fixed
- 🐛 修正台灣 IP 被誤判為中國的問題
  - 移除不準確的靜態 CIDR 列表判斷
  - 改用 MaxMind 動態判斷國家歸屬
- 🔧 修正 Docker 建置配置
  - 移除不存在的 config.example.yaml 複製指令

### Security
- ✅ Code Review 通過
  - 無硬編碼密碼或 token
  - 適當的錯誤處理和空值檢查
  - 使用 RWMutex 確保並發安全

## [1.0.0] - 2024-01-XX

### Added
- 初始版本發布
- 基於 MaxMind GeoLite2 的 IP 查詢服務
- Redis 快取支援
- 限流保護
- 批次查詢功能
- Docker Compose 部署
