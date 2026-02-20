# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- 初始化專案結構與設計文檔
- 完整的 API 設計規格（DESIGN.md）
- 環境變數配置範本（.env.example）
- Git 配置檔案（.gitignore）
- MaxMind GeoLite2-City 資料庫
- 專案說明文檔（README.md）

### Features
- RESTful API 端點設計
  - 單一 IP 查詢 (GET /api/v1/ip/{ip})
  - 批次 IP 查詢 (POST /api/v1/ip/batch)
  - 健康檢查 (GET /api/v1/health)
- Redis 分散式快取架構
  - L1 本地快取（可選）
  - L2 Redis 快取
  - Cache-Aside 模式
  - Pipeline 批次操作優化
- 分散式限流機制
  - 基於 Redis Sorted Set 的滑動窗口
  - 支援每分鐘/每小時限流
  - 多實例支援
- 完整的錯誤處理與降級策略
  - Redis 故障自動降級
  - 快取操作錯誤不影響主流程

### Technical Specifications
- Repository 層設計
  - MaxMind DB 存取層
  - Redis Cache 存取層
- Service 層業務邏輯
- Handler 層 HTTP 處理
- Middleware 層
  - Logger 日誌中間件
  - Rate Limiter 限流中間件
  - Recovery 錯誤恢復中間件

### Documentation
- 完整的架構設計文檔（54+ 章節）
- API 規格與回應格式
- Redis 整合詳細設計
- 部署架構與建議
- 效能指標與監控方案
