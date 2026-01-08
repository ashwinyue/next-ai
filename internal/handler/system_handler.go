package handler

import (
	"fmt"

	"github.com/ashwinyue/next-ai/internal/service"
	"github.com/gin-gonic/gin"
)

// SystemHandler 系统处理器
type SystemHandler struct {
	svc *service.Services
}

// NewSystemHandler 创建系统处理器
func NewSystemHandler(svc *service.Services) *SystemHandler {
	return &SystemHandler{svc: svc}
}

// GetSystemInfo 获取系统信息
// GET /api/v1/system/info
func (h *SystemHandler) GetSystemInfo(c *gin.Context) {
	info, err := h.svc.Initialization.GetSystemInfo(c.Request.Context())
	if err != nil {
		Error(c, err)
		return
	}

	Success(c, info)
}

// MinioBucketInfo MinIO bucket 信息
type MinioBucketInfo struct {
	Name         string `json:"name"`
	CreationDate string `json:"creation_date"`
	Policy       string `json:"policy"` // public, private, etc.
}

// ListMinioBuckets 列出 MinIO buckets（WeKnora API 兼容）
// GET /api/v1/system/minio/buckets
func (h *SystemHandler) ListMinioBuckets(c *gin.Context) {
	buckets, err := h.svc.File.ListBuckets(c.Request.Context())
	if err != nil {
		Error(c, err)
		return
	}

	// 转换为响应格式
	result := make([]MinioBucketInfo, 0, len(buckets))
	for _, bucket := range buckets {
		info := MinioBucketInfo{
			Name: bucket["name"].(string),
		}
		if creationDate, ok := bucket["creation_date"]; ok {
			info.CreationDate = fmt.Sprintf("%v", creationDate)
		}
		result = append(result, info)
	}

	Success(c, gin.H{
		"success": true,
		"data": gin.H{
			"buckets": result,
		},
	})
}

// WebSearchProviderInfo 网络搜索服务提供商信息
type WebSearchProviderInfo struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Free           bool   `json:"free"`
	RequiresAPIKey bool   `json:"requires_api_key"`
	Description    string `json:"description"`
	APIURL         string `json:"api_url"`
}

// GetWebSearchProviders 获取网络搜索服务提供商列表（WeKnora API 兼容）
// GET /api/v1/web-search/providers
// 返回系统内置支持的搜索提供商列表
func (h *SystemHandler) GetWebSearchProviders(c *gin.Context) {
	providers := []WebSearchProviderInfo{
		{
			ID:             "duckduckgo",
			Name:           "DuckDuckGo",
			Free:           true,
			RequiresAPIKey: false,
			Description:    "免费的隐私搜索引擎，无需 API Key",
			APIURL:         "https://api.duckduckgo.com/",
		},
		{
			ID:             "bing",
			Name:           "Bing Search",
			Free:           false,
			RequiresAPIKey: true,
			Description:    "微软必应搜索 API，需要订阅 Azure Cognitive Services",
			APIURL:         "https://api.bing.microsoft.com/v7.0/search",
		},
		{
			ID:             "brave",
			Name:           "Brave Search",
			Free:           true,
			RequiresAPIKey: false,
			Description:    "Brave 私密搜索，提供独立搜索结果",
			APIURL:         "https://api.search.brave.com/app/api/search",
		},
		{
			ID:             "google",
			Name:           "Google Custom Search",
			Free:           false,
			RequiresAPIKey: true,
			Description:    "Google 自定义搜索 API，需要 API Key 和 CX",
			APIURL:         "https://www.googleapis.com/customsearch/v1",
		},
		{
			ID:             "serpapi",
			Name:           "SerpAPI",
			Free:           false,
			RequiresAPIKey: true,
			Description:    "支持 Google、Bing、Yahoo 等多种搜索引擎的聚合 API",
			APIURL:         "https://serpapi.com/",
		},
	}

	Success(c, gin.H{
		"success": true,
		"data":    providers,
	})
}
