package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/axiom/goip/internal/model"
	"github.com/axiom/goip/internal/repository"
	"github.com/axiom/goip/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

// IPHandler HTTP 請求處理器
type IPHandler struct {
	service      service.IPService
	cache        repository.CacheRepository
	geoip        repository.GeoIPRepository
	logger       zerolog.Logger
	batchMaxSize int
}

// NewIPHandler 建立新的 IP Handler
func NewIPHandler(
	service service.IPService,
	cache repository.CacheRepository,
	geoip repository.GeoIPRepository,
	logger zerolog.Logger,
	batchMaxSize int,
) *IPHandler {
	return &IPHandler{
		service:      service,
		cache:        cache,
		geoip:        geoip,
		logger:       logger,
		batchMaxSize: batchMaxSize,
	}
}

// HandleIPLookup 處理單一 IP 查詢
// @Summary 查詢單一 IP 的地理位置
// @Tags IP
// @Accept json
// @Produce json
// @Param ip path string true "IP 地址"
// @Success 200 {object} model.IPInfo
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Router /api/v1/ip/{ip} [get]
func (h *IPHandler) HandleIPLookup(c *gin.Context) {
	ip := c.Param("ip")

	result, err := h.service.LookupIP(c.Request.Context(), ip)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// 將資料來源存入 context，供 logger middleware 使用
	c.Set("source", result.Source)
	c.Set("provider", result.Provider)

	c.JSON(http.StatusOK, result)
}

// HandleIPLookupByProvider 處理使用指定提供者查詢 IP
// @Summary 使用指定的資料庫提供者查詢 IP
// @Tags IP
// @Accept json
// @Produce json
// @Param ip path string true "IP 地址"
// @Param provider query string true "提供者名稱 (maxmind, ipip)"
// @Success 200 {object} model.IPInfo
// @Failure 400 {object} model.ErrorResponse
// @Failure 404 {object} model.ErrorResponse
// @Router /api/v1/ip/{ip}/provider [get]
func (h *IPHandler) HandleIPLookupByProvider(c *gin.Context) {
	ip := c.Param("ip")
	provider := c.Query("provider")

	if provider == "" {
		h.respondError(c, http.StatusBadRequest, "INVALID_REQUEST", "provider parameter is required")
		return
	}

	result, err := h.service.LookupIPByProvider(c.Request.Context(), ip, provider)
	if err != nil {
		h.handleError(c, err)
		return
	}

	// 將資料來源存入 context，供 logger middleware 使用
	c.Set("source", result.Source)
	c.Set("provider", result.Provider)

	c.JSON(http.StatusOK, result)
}

// HandleGetProviders 取得所有可用的提供者
// @Summary 列出所有可用的資料庫提供者
// @Tags System
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Router /api/v1/providers [get]
func (h *IPHandler) HandleGetProviders(c *gin.Context) {
	providers := h.service.GetAvailableProviders()

	c.JSON(http.StatusOK, gin.H{
		"providers": providers,
		"count":     len(providers),
	})
}

// HandleBatchLookup 處理批次 IP 查詢
// @Summary 批次查詢多個 IP 的地理位置
// @Tags IP
// @Accept json
// @Produce json
// @Param request body model.BatchRequest true "批次查詢請求"
// @Success 200 {object} model.BatchResult
// @Failure 400 {object} model.ErrorResponse
// @Router /api/v1/ip/batch [post]
func (h *IPHandler) HandleBatchLookup(c *gin.Context) {
	var req model.BatchRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	// 檢查批次大小限制
	if len(req.IPs) > h.batchMaxSize {
		h.respondError(c, http.StatusBadRequest, "BATCH_TOO_LARGE",
			fmt.Sprintf("批次查詢數量超過限制，最多支援 %d 個 IP", h.batchMaxSize))
		return
	}

	result, err := h.service.BatchLookup(c.Request.Context(), req.IPs)
	if err != nil {
		h.handleError(c, err)
		return
	}

	c.JSON(http.StatusOK, result)
}

// HandleHealth 健康檢查
// @Summary 健康檢查
// @Tags System
// @Produce json
// @Success 200 {object} model.HealthResponse
// @Failure 503 {object} model.ErrorResponse
// @Router /api/v1/health [get]
func (h *IPHandler) HandleHealth(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), 5*time.Second)
	defer cancel()

	services := make(map[string]string)

	// 檢查 Redis
	if err := h.cache.HealthCheck(ctx); err != nil {
		services["redis"] = "unhealthy: " + err.Error()
	} else {
		services["redis"] = "healthy"
	}

	// 檢查 MaxMind DB（嘗試查詢一個 IP）
	if _, err := h.geoip.LookupCountry("8.8.8.8"); err != nil {
		services["maxmind"] = "unhealthy: " + err.Error()
	} else {
		services["maxmind"] = "healthy"
	}

	// 判斷整體狀態
	status := "healthy"
	httpStatus := http.StatusOK

	for _, svcStatus := range services {
		if svcStatus != "healthy" {
			status = "unhealthy"
			httpStatus = http.StatusServiceUnavailable
			break
		}
	}

	c.JSON(httpStatus, model.HealthResponse{
		Status:   status,
		Services: services,
	})
}

// HandleStats 獲取服務統計
// @Summary 獲取服務統計資訊
// @Tags System
// @Produce json
// @Success 200 {object} model.ServiceStats
// @Router /api/v1/stats [get]
func (h *IPHandler) HandleStats(c *gin.Context) {
	stats := h.service.GetStats()
	c.JSON(http.StatusOK, stats)
}

// HandleCacheStats 獲取快取統計
// @Summary 獲取 Redis 快取統計
// @Tags System
// @Produce json
// @Success 200 {object} model.CacheStats
// @Failure 500 {object} model.ErrorResponse
// @Router /api/v1/cache/stats [get]
func (h *IPHandler) HandleCacheStats(c *gin.Context) {
	stats, err := h.cache.GetStats(c.Request.Context())
	if err != nil {
		h.respondError(c, http.StatusInternalServerError, "CACHE_ERROR", err.Error())
		return
	}

	c.JSON(http.StatusOK, stats)
}

// HandleInvalidateCache 清除快取
// @Summary 清除指定 IP 的快取
// @Tags Cache
// @Accept json
// @Produce json
// @Param request body model.BatchRequest true "要清除的 IP 列表"
// @Success 200 {object} map[string]string
// @Failure 400 {object} model.ErrorResponse
// @Router /api/v1/cache/invalidate [post]
func (h *IPHandler) HandleInvalidateCache(c *gin.Context) {
	var req model.BatchRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		h.respondError(c, http.StatusBadRequest, "INVALID_REQUEST", err.Error())
		return
	}

	if err := h.service.InvalidateCache(c.Request.Context(), req.IPs...); err != nil {
		h.respondError(c, http.StatusInternalServerError, "CACHE_ERROR", err.Error())
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "快取已清除",
		"count":   len(req.IPs),
	})
}

// handleError 統一錯誤處理
func (h *IPHandler) handleError(c *gin.Context, err error) {
	switch err {
	case repository.ErrInvalidIP:
		h.respondError(c, http.StatusBadRequest, "INVALID_IP", "IP 地址格式無效")
	case repository.ErrIPNotFound:
		h.respondError(c, http.StatusNotFound, "IP_NOT_FOUND", "IP 不在資料庫中")
	case repository.ErrDatabaseClosed:
		h.respondError(c, http.StatusServiceUnavailable, "DB_ERROR", "資料庫連接已關閉")
	default:
		h.respondError(c, http.StatusInternalServerError, "INTERNAL_ERROR", err.Error())
	}
}

// respondError 回應錯誤
func (h *IPHandler) respondError(c *gin.Context, httpStatus int, code, message string) {
	h.logger.Error().
		Str("code", code).
		Str("message", message).
		Str("path", c.Request.URL.Path).
		Msg("Request error")

	c.JSON(httpStatus, model.ErrorResponse{
		Error:     message,
		Code:      code,
		Timestamp: time.Now(),
	})
}
