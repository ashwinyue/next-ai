// Package handler 提供网络搜索相关的 HTTP 处理器
// 对齐 WeKnora 的 web_search.go handler
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// WebSearchHandler 网络搜索处理器
type WebSearchHandler struct {
	providers []WebSearchProviderInfo
}

// NewWebSearchHandler 创建网络搜索处理器
func NewWebSearchHandler() *WebSearchHandler {
	return &WebSearchHandler{
		providers: getDefaultWebSearchProviders(),
	}
}

// getDefaultWebSearchProviders 获取默认的网络搜索提供商列表
func getDefaultWebSearchProviders() []WebSearchProviderInfo {
	return []WebSearchProviderInfo{
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
			APIURL:         "https://api.bing.microsoft.com/",
		},
		{
			ID:             "google",
			Name:           "Google Search",
			Free:           false,
			RequiresAPIKey: true,
			Description:    "Google Custom Search API，需要 API Key 和 CX",
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
}

// GetProviders godoc
// @Summary      获取网络搜索提供商列表
// @Description  返回可用的网络搜索提供商列表
// @Tags         网络搜索
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]interface{}  "提供商列表"
// @Security     Bearer
// @Router       /web-search/providers [get]
func (h *WebSearchHandler) GetProviders(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    h.providers,
	})
}

// SearchRequest 网络搜索请求
type SearchRequest struct {
	Query    string `json:"query" binding:"required"`
	Provider string `json:"provider"`
	NumResults int  `json:"num_results"`
}

// SearchResponse 网络搜索响应
type SearchResponse struct {
	Query    string         `json:"query"`
	Provider string         `json:"provider"`
	Results  []SearchResult `json:"results"`
}

// SearchResult 搜索结果
type SearchResult struct {
	Title   string `json:"title"`
	URL     string `json:"url"`
	Snippet string `json:"snippet"`
}

// Search godoc
// @Summary      执行网络搜索
// @Description  使用指定的搜索引擎执行网络搜索（需要集成实际的搜索 API）
// @Tags         网络搜索
// @Accept       json
// @Produce      json
// @Param        request  body      SearchRequest  true  "搜索请求"
// @Success      200      {object}  map[string]interface{}  "搜索结果"
// @Failure      400      {object}  map[string]interface{}  "请求参数错误"
// @Security     Bearer
// @Router       /web-search/search [post]
func (h *WebSearchHandler) Search(c *gin.Context) {
	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "请求参数错误",
		})
		return
	}

	if req.Query == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "搜索查询不能为空",
		})
		return
	}

	// 默认使用 duckduckgo
	if req.Provider == "" {
		req.Provider = "duckduckgo"
	}

	// 设置默认返回结果数量
	if req.NumResults <= 0 {
		req.NumResults = 10
	}
	if req.NumResults > 100 {
		req.NumResults = 100
	}

	// 验证提供商是否存在
	providerExists := false
	for _, p := range h.providers {
		if p.ID == req.Provider {
			providerExists = true
			break
		}
	}

	if !providerExists {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"message": "不支持的搜索引擎",
		})
		return
	}

	// 简化版：返回空结果
	// 实际实现需要集成具体的搜索 API
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": SearchResponse{
			Query:    req.Query,
			Provider: req.Provider,
			Results:  []SearchResult{},
		},
		"message": "搜索功能需要集成具体的搜索 API",
	})
}
