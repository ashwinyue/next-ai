package knowledge

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"sync"

	"github.com/ashwinyue/next-ai/internal/config"
	"github.com/ashwinyue/next-ai/internal/model"
	"github.com/ashwinyue/next-ai/internal/repository"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/google/uuid"
)

// Service 知识库服务
type Service struct {
	repo              repository.KnowledgeRepository // 使用接口便于测试
	cfg               *config.Config
	embedder          embedding.Embedder
	documentProcessor *DocumentProcessor
	esSearcher        ESSearcher // ES 搜索接口，便于测试
	cloneProgressMap  sync.Map   // 存储 KBCloneProgress，key: taskID, value: *KBCloneProgress
}

// NewService 创建知识库服务
func NewService(repos *repository.Repositories, cfg *config.Config, embedder embedding.Embedder) *Service {
	docProcessor := NewDocumentProcessor(repos, cfg, embedder)

	// 创建 ES 搜索器（用于混合搜索）
	var esSearcher ESSearcher
	if cfg.Elastic.Host != "" {
		esClient, _ := elasticsearch.NewClient(elasticsearch.Config{
			Addresses: []string{cfg.Elastic.Host},
			Username:  cfg.Elastic.Username,
			Password:  cfg.Elastic.Password,
		})
		// 创建适配器包装真实 ES 客户端
		esSearcher = newRealESSearcher(func(ctx context.Context, index string, body io.Reader) (*ESResponseImpl, error) {
			res, err := esClient.Search(
				esClient.Search.WithContext(ctx),
				esClient.Search.WithIndex(index),
				esClient.Search.WithBody(body),
			)
			if err != nil {
				return nil, err
			}
			return &ESResponseImpl{
				isError: res.IsError(),
				body:    res.Body,
				str:     res.String(),
			}, nil
		})
	}

	return &Service{
		repo:              repos.Knowledge, // 使用接口
		cfg:               cfg,
		embedder:          embedder,
		documentProcessor: docProcessor,
		esSearcher:        esSearcher,
	}
}

// CreateKnowledgeBaseRequest 创建知识库请求
type CreateKnowledgeBaseRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description"`
	EmbedModel  string `json:"embed_model"`
}

// CreateKnowledgeBase 创建知识库
func (s *Service) CreateKnowledgeBase(ctx context.Context, req *CreateKnowledgeBaseRequest) (*model.KnowledgeBase, error) {
	kb := &model.KnowledgeBase{
		ID:             uuid.New().String(),
		Name:           req.Name,
		Description:    req.Description,
		EmbeddingModel: req.EmbedModel,
		IndexName:      "kb_" + uuid.New().String(),
		ChunkSize:      512,
		ChunkOverlap:   50,
	}

	if err := s.repo.CreateKnowledgeBase(kb); err != nil {
		return nil, fmt.Errorf("failed to create knowledge base: %w", err)
	}

	return kb, nil
}

// GetKnowledgeBase 获取知识库
func (s *Service) GetKnowledgeBase(ctx context.Context, id string) (*model.KnowledgeBase, error) {
	return s.repo.GetKnowledgeBaseByID(id)
}

// ListKnowledgeBasesRequest 列出知识库请求
type ListKnowledgeBasesRequest struct {
	Page int `json:"page"`
	Size int `json:"size"`
}

// ListKnowledgeBases 列出知识库
func (s *Service) ListKnowledgeBases(ctx context.Context, req *ListKnowledgeBasesRequest) ([]*model.KnowledgeBase, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 || req.Size > 100 {
		req.Size = 20
	}

	offset := (req.Page - 1) * req.Size
	return s.repo.ListKnowledgeBases(offset, req.Size)
}

// UpdateKnowledgeBase 更新知识库
func (s *Service) UpdateKnowledgeBase(ctx context.Context, id string, req *CreateKnowledgeBaseRequest) (*model.KnowledgeBase, error) {
	kb, err := s.repo.GetKnowledgeBaseByID(id)
	if err != nil {
		return nil, fmt.Errorf("knowledge base not found: %w", err)
	}

	kb.Name = req.Name
	kb.Description = req.Description
	kb.EmbeddingModel = req.EmbedModel

	if err := s.repo.UpdateKnowledgeBase(kb); err != nil {
		return nil, fmt.Errorf("failed to update knowledge base: %w", err)
	}

	return kb, nil
}

// DeleteKnowledgeBase 删除知识库
func (s *Service) DeleteKnowledgeBase(ctx context.Context, id string) error {
	if err := s.repo.DeleteKnowledgeBase(id); err != nil {
		return fmt.Errorf("failed to delete knowledge base: %w", err)
	}
	return nil
}

// UploadDocumentRequest 上传文档请求
type UploadDocumentRequest struct {
	KnowledgeBaseID string `json:"knowledge_base_id" binding:"required"`
	Title           string `json:"title" binding:"required"`
	FileName        string `json:"file_name" binding:"required"`
	FilePath        string `json:"file_path"`
	FileSize        int64  `json:"file_size"`
}

// UploadDocument 上传文档
func (s *Service) UploadDocument(ctx context.Context, req *UploadDocumentRequest) (*model.Document, error) {
	// 检查知识库是否存在
	_, err := s.repo.GetKnowledgeBaseByID(req.KnowledgeBaseID)
	if err != nil {
		return nil, fmt.Errorf("knowledge base not found: %w", err)
	}

	doc := &model.Document{
		ID:              uuid.New().String(),
		KnowledgeBaseID: req.KnowledgeBaseID,
		Title:           req.Title,
		FileName:        req.FileName,
		FilePath:        req.FilePath,
		FileSize:        req.FileSize,
		Status:          "pending",
	}

	if err := s.repo.CreateDocument(doc); err != nil {
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	return doc, nil
}

// GetDocument 获取文档
func (s *Service) GetDocument(ctx context.Context, id string) (*model.Document, error) {
	return s.repo.GetDocumentByID(id)
}

// ListDocumentsRequest 列出文档请求
type ListDocumentsRequest struct {
	KnowledgeBaseID string `json:"knowledge_base_id"`
	Page            int    `json:"page"`
	Size            int    `json:"size"`
}

// ListDocuments 列出文档
func (s *Service) ListDocuments(ctx context.Context, req *ListDocumentsRequest) ([]*model.Document, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.Size <= 0 || req.Size > 100 {
		req.Size = 20
	}

	offset := (req.Page - 1) * req.Size
	return s.repo.ListDocuments(req.KnowledgeBaseID, offset, req.Size)
}

// DeleteDocument 删除文档
func (s *Service) DeleteDocument(ctx context.Context, id string) error {
	if err := s.repo.DeleteDocument(id); err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	return nil
}

// UpdateDocumentStatus 更新文档状态
func (s *Service) UpdateDocumentStatus(ctx context.Context, id, status string, chunkCount int) error {
	doc, err := s.repo.GetDocumentByID(id)
	if err != nil {
		return fmt.Errorf("document not found: %w", err)
	}

	doc.Status = status
	if chunkCount > 0 {
		doc.ChunkCount = chunkCount
	}

	return s.repo.UpdateDocument(doc)
}

// ProcessDocument 处理文档（解析、分块、向量化、索引）
// 直接调用 DocumentProcessor
func (s *Service) ProcessDocument(ctx context.Context, documentID, knowledgeBaseID string) (*ProcessResult, error) {
	return s.documentProcessor.Process(ctx, &ProcessRequest{
		DocumentID:      documentID,
		KnowledgeBaseID: knowledgeBaseID,
	})
}

// SearchKnowledgeRequest 搜索知识库请求
type SearchKnowledgeRequest struct {
	KnowledgeBaseID string `json:"knowledge_base_id"`
	Query           string `json:"query" binding:"required"`
	TopK            int    `json:"top_k"`
}

// SearchResult 搜索结果
type SearchResult struct {
	Content  string                 `json:"content"`
	Score    float64                `json:"score"`
	Metadata map[string]interface{} `json:"metadata,omitempty"`
}

// Search 搜索知识库（使用 Elasticsearch）
func (s *Service) Search(ctx context.Context, req *SearchKnowledgeRequest) ([]*SearchResult, error) {
	if s.esSearcher == nil {
		return nil, fmt.Errorf("elasticsearch client not configured")
	}

	// 设置默认 topK
	topK := req.TopK
	if topK <= 0 {
		topK = 10
	}

	// 获取知识库信息以获取索引名
	kb, err := s.repo.GetKnowledgeBaseByID(req.KnowledgeBaseID)
	if err != nil {
		return nil, fmt.Errorf("knowledge base not found: %w", err)
	}

	// 构建 ES 查询
	query := map[string]interface{}{
		"size": topK,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must": []interface{}{
					map[string]interface{}{
						"term": map[string]interface{}{
							"knowledge_base_id": req.KnowledgeBaseID,
						},
					},
				},
				"should": []interface{}{
					map[string]interface{}{
						"match": map[string]interface{}{
							"content": map[string]interface{}{
								"query": req.Query,
							},
						},
					},
				},
			},
		},
	}

	// 序列化查询
	queryJSON, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	// 执行搜索
	res, err := s.esSearcher.DoSearch(ctx, kb.IndexName+"_chunks", queryJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError {
		return nil, fmt.Errorf("elasticsearch error: %s", res.String)
	}

	// 解析响应
	var response struct {
		Hits struct {
			Hits []struct {
				ID     string                 `json:"_id"`
				Score  float64                `json:"_score"`
				Source map[string]interface{} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// 转换结果
	results := make([]*SearchResult, 0, len(response.Hits.Hits))
	for _, hit := range response.Hits.Hits {
		content := ""
		if c, ok := hit.Source["content"].(string); ok {
			content = c
		}

		result := &SearchResult{
			Content:  content,
			Score:    hit.Score,
			Metadata: hit.Source,
		}
		results = append(results, result)
	}

	return results, nil
}

// ========== 混合搜索 (Hybrid Search) ==========

// HybridSearchParams 混合搜索请求参数
// 兼容 WeKnora API 格式
type HybridSearchParams struct {
	QueryText            string   `json:"query_text" binding:"required"`
	VectorThreshold      float64  `json:"vector_threshold"`
	KeywordThreshold     float64  `json:"keyword_threshold"`
	MatchCount           int      `json:"match_count"`
	DisableKeywordsMatch bool     `json:"disable_keywords_match"`
	DisableVectorMatch   bool     `json:"disable_vector_match"`
	KnowledgeIDs         []string `json:"knowledge_ids"`
	TagIDs               []string `json:"tag_ids"`
}

// HybridSearchResult 混合搜索结果
// 兼容 WeKnora API 格式
type HybridSearchResult struct {
	ID                string            `json:"id"`
	Content           string            `json:"content"`
	KnowledgeID       string            `json:"knowledge_id"`
	ChunkIndex        int               `json:"chunk_index"`
	KnowledgeTitle    string            `json:"knowledge_title"`
	StartAt           int               `json:"start_at"`
	EndAt             int               `json:"end_at"`
	Seq               int               `json:"seq"`
	Score             float64           `json:"score"`
	ChunkType         string            `json:"chunk_type"`
	ImageInfo         string            `json:"image_info,omitempty"`
	Metadata          map[string]string `json:"metadata,omitempty"`
	KnowledgeFilename string            `json:"knowledge_filename,omitempty"`
	KnowledgeSource   string            `json:"knowledge_source,omitempty"`
}

// HybridSearch 执行混合搜索（向量 + 关键词）
// 使用 Elasticsearch 的 bool 查询结合向量相似度和 BM25
func (s *Service) HybridSearch(ctx context.Context, kbID string, params *HybridSearchParams) ([]*HybridSearchResult, error) {
	// 获取知识库信息
	kb, err := s.repo.GetKnowledgeBaseByID(kbID)
	if err != nil {
		return nil, fmt.Errorf("knowledge base not found: %w", err)
	}

	// 设置默认参数
	topK := params.MatchCount
	if topK <= 0 {
		topK = 10
	}

	vectorThreshold := params.VectorThreshold
	if vectorThreshold <= 0 {
		vectorThreshold = 0.7
	}

	keywordThreshold := params.KeywordThreshold
	if keywordThreshold <= 0 {
		keywordThreshold = 0.1
	}

	// 如果启用了向量搜索，获取查询向量
	var queryVector []float64
	if !params.DisableVectorMatch && s.embedder != nil {
		vectors, err := s.embedder.EmbedStrings(ctx, []string{params.QueryText})
		if err == nil && len(vectors) > 0 && len(vectors[0]) > 0 {
			queryVector = vectors[0]
		}
		// 如果获取向量失败，继续使用仅关键词搜索
	}

	// 构建混合搜索查询
	query := s.buildHybridSearchQuery(kb.IndexName, params, queryVector, vectorThreshold, keywordThreshold, topK)

	// 执行搜索
	results, err := s.executeHybridSearch(ctx, kb.IndexName, query, topK)
	if err != nil {
		return nil, fmt.Errorf("failed to execute hybrid search: %w", err)
	}

	// 填充文档元数据
	if err := s.fillSearchResultsMetadata(ctx, results); err != nil {
		// 元数据填充失败不影响返回结果
		return results, nil
	}

	return results, nil
}

// buildHybridSearchQuery 构建 Elasticsearch 混合搜索查询
func (s *Service) buildHybridSearchQuery(indexName string, params *HybridSearchParams,
	queryVector []float64, vectorThreshold, keywordThreshold float64, topK int) map[string]interface{} {

	// 构建查询子句
	mustClauses := []interface{}{}
	shouldClauses := []interface{}{}

	// 关键词匹配 (BM25) - 如果未禁用
	if !params.DisableKeywordsMatch {
		shouldClauses = append(shouldClauses, map[string]interface{}{
			"match": map[string]interface{}{
				"content": map[string]interface{}{
					"query": params.QueryText,
				},
			},
		})
	}

	// 向量相似度搜索 - 如果未禁用且有向量
	if !params.DisableVectorMatch && len(queryVector) > 0 {
		shouldClauses = append(shouldClauses, map[string]interface{}{
			"script_score": map[string]interface{}{
				"query": map[string]interface{}{
					"match_all": map[string]interface{}{},
				},
				"script": map[string]interface{}{
					"source": "cosineSimilarity(params.query_vector, 'content_vector') + 1.0",
					"params": map[string]interface{}{
						"query_vector": queryVector,
					},
				},
				"min_score": vectorThreshold,
			},
		})
	}

	// 添加知识库过滤
	mustClauses = append(mustClauses, map[string]interface{}{
		"term": map[string]interface{}{
			"knowledge_base_id": kbIDFromIndex(indexName),
		},
	})

	// 添加知识 ID 过滤
	if len(params.KnowledgeIDs) > 0 {
		mustClauses = append(mustClauses, map[string]interface{}{
			"terms": map[string]interface{}{
				"knowledge_id": params.KnowledgeIDs,
			},
		})
	}

	// 构建最终查询
	query := map[string]interface{}{
		"size": topK,
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must":   mustClauses,
				"should": shouldClauses,
			},
		},
	}

	return query
}

// executeHybridSearch 执行 Elasticsearch 查询
func (s *Service) executeHybridSearch(ctx context.Context, indexName string,
	query map[string]interface{}, topK int) ([]*HybridSearchResult, error) {

	if s.esSearcher == nil {
		return nil, fmt.Errorf("elasticsearch client not configured")
	}

	// 序列化查询
	queryJSON, err := json.Marshal(query)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal query: %w", err)
	}

	// 执行搜索
	res, err := s.esSearcher.DoSearch(ctx, indexName+"_chunks", queryJSON)
	if err != nil {
		return nil, fmt.Errorf("failed to execute search: %w", err)
	}
	defer res.Body.Close()

	if res.IsError {
		return nil, fmt.Errorf("elasticsearch error: %s", res.String)
	}

	// 解析响应
	var esResponse struct {
		Hits struct {
			Hits []struct {
				ID     string                 `json:"_id"`
				Score  float64                `json:"_score"`
				Source map[string]interface{} `json:"_source"`
			} `json:"hits"`
		} `json:"hits"`
	}

	if err := json.NewDecoder(res.Body).Decode(&esResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// 转换为搜索结果
	results := make([]*HybridSearchResult, 0, len(esResponse.Hits.Hits))
	for _, hit := range esResponse.Hits.Hits {
		result := &HybridSearchResult{
			ID:    hit.ID,
			Score: hit.Score,
		}

		// 从 _source 提取字段
		if content, ok := hit.Source["content"].(string); ok {
			result.Content = content
		}
		if knowledgeID, ok := hit.Source["knowledge_id"].(string); ok {
			result.KnowledgeID = knowledgeID
		}
		if chunkIndex, ok := hit.Source["chunk_index"].(float64); ok {
			result.ChunkIndex = int(chunkIndex)
		}

		results = append(results, result)
	}

	return results, nil
}

// fillSearchResultsMetadata 填充搜索结果的文档元数据
func (s *Service) fillSearchResultsMetadata(ctx context.Context, results []*HybridSearchResult) error {
	// 收集需要查询的知识 ID
	knowledgeIDs := make([]string, 0, len(results))
	for _, r := range results {
		if r.KnowledgeID != "" {
			knowledgeIDs = append(knowledgeIDs, r.KnowledgeID)
		}
	}

	if len(knowledgeIDs) == 0 {
		return nil
	}

	// 批量获取文档信息
	for _, result := range results {
		if result.KnowledgeID == "" {
			continue
		}

		// 从数据库获取文档详情
		doc, err := s.repo.GetDocumentByID(result.KnowledgeID)
		if err != nil {
			continue
		}

		// 填充元数据
		result.KnowledgeFilename = doc.FileName
		result.KnowledgeSource = doc.ContentType

		// 获取文档标题
		if doc.Title != "" {
			result.KnowledgeTitle = doc.Title
		} else {
			result.KnowledgeTitle = doc.FileName
		}
	}

	return nil
}

// kbIDFromIndex 从索引名称提取知识库 ID
func kbIDFromIndex(indexName string) string {
	return indexName
}

// ========== 知识库复制 (Copy Knowledge Base) ==========

// CopyKnowledgeBaseRequest 复制知识库请求
type CopyKnowledgeBaseRequest struct {
	SourceID string `json:"source_id" binding:"required"`
	TargetID string `json:"target_id"` // 可选，不提供则自动创建
	Name     string `json:"name"`      // 目标知识库名称（仅当 TargetID 为空时使用）
}

// CopyKnowledgeBaseResponse 复制知识库响应
type CopyKnowledgeBaseResponse struct {
	TaskID   string `json:"task_id"`
	SourceID string `json:"source_id"`
	TargetID string `json:"target_id"`
	Message  string `json:"message"`
}

// KBCloneProgress 知识库复制进度
type KBCloneProgress struct {
	TaskID     string  `json:"task_id"`
	SourceID   string  `json:"source_id"`
	TargetID   string  `json:"target_id"`
	Status     string  `json:"status"` // pending, processing, completed, failed
	Progress   float64 `json:"progress"`
	TotalDocs  int     `json:"total_docs"`
	CopiedDocs int     `json:"copied_docs"`
	Error      string  `json:"error,omitempty"`
}

// CopyKnowledgeBase 复制知识库（异步）
func (s *Service) CopyKnowledgeBase(ctx context.Context, req *CopyKnowledgeBaseRequest) (*CopyKnowledgeBaseResponse, error) {
	// 验证源知识库存在
	sourceKB, err := s.repo.GetKnowledgeBaseByID(req.SourceID)
	if err != nil {
		return nil, fmt.Errorf("source knowledge base not found: %w", err)
	}

	var targetID string
	var targetKB *model.KnowledgeBase

	// 如果提供了 target_id，验证目标知识库存在
	if req.TargetID != "" {
		targetKB, err = s.repo.GetKnowledgeBaseByID(req.TargetID)
		if err != nil {
			return nil, fmt.Errorf("target knowledge base not found: %w", err)
		}
		targetID = req.TargetID
	} else {
		// 创建新的目标知识库
		name := req.Name
		if name == "" {
			name = sourceKB.Name + " (副本)"
		}

		targetKB = &model.KnowledgeBase{
			ID:             uuid.New().String(),
			Name:           name,
			Description:    sourceKB.Description,
			EmbeddingModel: sourceKB.EmbeddingModel,
			IndexName:      "kb_" + uuid.New().String(),
			ChunkSize:      sourceKB.ChunkSize,
			ChunkOverlap:   sourceKB.ChunkOverlap,
		}

		if err := s.repo.CreateKnowledgeBase(targetKB); err != nil {
			return nil, fmt.Errorf("failed to create target knowledge base: %w", err)
		}
		targetID = targetKB.ID
	}

	// 生成任务 ID
	taskID := uuid.New().String()

	// 创建初始进度记录（如果需要 Redis 存储，可以在这里实现）
	progress := &KBCloneProgress{
		TaskID:     taskID,
		SourceID:   req.SourceID,
		TargetID:   targetID,
		Status:     "pending",
		Progress:   0,
		TotalDocs:  0,
		CopiedDocs: 0,
	}

	// 异步执行复制（简化版：使用 goroutine）
	// 注意：生产环境建议使用专门的任务队列（如 Asynq）
	go s.executeCopy(context.Background(), taskID, req.SourceID, targetID, progress)

	return &CopyKnowledgeBaseResponse{
		TaskID:   taskID,
		SourceID: req.SourceID,
		TargetID: targetID,
		Message:  "Knowledge base copy task started",
	}, nil
}

// executeCopy 执行知识库复制
func (s *Service) executeCopy(ctx context.Context, taskID, sourceID, targetID string, progress *KBCloneProgress) {
	// 更新状态为处理中并保存到 map
	progress.Status = "processing"
	s.cloneProgressMap.Store(taskID, progress)

	// 获取源知识库的所有文档
	docs, err := s.repo.ListDocuments(sourceID, 0, 1000)
	if err != nil {
		progress.Status = "failed"
		progress.Error = err.Error()
		s.cloneProgressMap.Store(taskID, progress)
		return
	}

	progress.TotalDocs = len(docs)
	s.cloneProgressMap.Store(taskID, progress)

	// 收集所有要复制的分块
	allChunks := make([]*model.DocumentChunk, 0)

	// 复制每个文档
	for i, doc := range docs {
		// 创建新文档
		newDoc := &model.Document{
			ID:              uuid.New().String(),
			KnowledgeBaseID: targetID,
			Title:           doc.Title,
			FileName:        doc.FileName,
			FilePath:        doc.FilePath,
			FileSize:        doc.FileSize,
			ContentType:     doc.ContentType,
			Status:          doc.Status,
			ChunkCount:      doc.ChunkCount,
		}

		if err := s.repo.CreateDocument(newDoc); err != nil {
			continue // 跳过失败的文档
		}

		// 获取源文档的分块
		chunks, _, err := s.repo.ListChunksByKnowledgeBaseID(sourceID, 0, 1000)
		if err == nil {
			for _, chunk := range chunks {
				if chunk.DocumentID == doc.ID {
					newChunk := &model.DocumentChunk{
						ID:              uuid.New().String(),
						DocumentID:      newDoc.ID,
						KnowledgeBaseID: targetID,
						ChunkIndex:      chunk.ChunkIndex,
						Content:         chunk.Content,
						Metadata:        chunk.Metadata,
					}
					allChunks = append(allChunks, newChunk)
				}
			}
		}

		progress.CopiedDocs = i + 1
		progress.Progress = float64(progress.CopiedDocs) / float64(progress.TotalDocs) * 100
		// 更新进度到 map
		s.cloneProgressMap.Store(taskID, progress)
	}

	// 批量创建分块
	if len(allChunks) > 0 {
		s.repo.CreateChunks(allChunks)
	}

	progress.Status = "completed"
	progress.Progress = 100
	s.cloneProgressMap.Store(taskID, progress)
}

// GetKBCloneProgress 获取知识库复制进度
func (s *Service) GetKBCloneProgress(ctx context.Context, taskID string) (*KBCloneProgress, error) {
	// 从内存 map 获取进度
	value, exists := s.cloneProgressMap.Load(taskID)
	if !exists {
		return &KBCloneProgress{
			TaskID:     taskID,
			Status:     "not_found",
			Progress:   0,
			TotalDocs:  0,
			CopiedDocs: 0,
			Error:      "Task not found",
		}, nil
	}

	progress, ok := value.(*KBCloneProgress)
	if !ok {
		return &KBCloneProgress{
			TaskID:     taskID,
			Status:     "error",
			Progress:   0,
			TotalDocs:  0,
			CopiedDocs: 0,
			Error:      "Invalid progress data",
		}, nil
	}

	return progress, nil
}

// ========== 父子 Chunk 功能（WeKnora 对齐）==========

// ListChunksByParentID 获取指定父 Chunk 的所有子 Chunk
// 对应 WeKnora 的 ChunkService.ListChunkByParentID
func (s *Service) ListChunksByParentID(ctx context.Context, parentID string) ([]*model.DocumentChunk, error) {
	return s.repo.ListChunksByParentID(parentID)
}

// GetParentChunk 获取指定 Chunk 的父 Chunk
// 用于追溯某个 Chunk 的父分块
func (s *Service) GetParentChunk(ctx context.Context, chunkID string) (*model.DocumentChunk, error) {
	return s.repo.GetParentChunk(chunkID)
}

// UpdateChunkParent 更新 Chunk 的父 Chunk ID
// 用于建立父子关系
func (s *Service) UpdateChunkParent(ctx context.Context, chunkID, parentID string) error {
	chunk, err := s.repo.GetChunkByID(chunkID)
	if err != nil {
		return fmt.Errorf("chunk not found: %w", err)
	}

	chunk.ParentChunkID = parentID
	return s.repo.UpdateChunk(chunk)
}

// UpdateChunkImageInfoRequest 更新分块图像信息请求
type UpdateChunkImageInfoRequest struct {
	ChunkID   string `json:"chunk_id" binding:"required"`
	ImageInfo string `json:"image_info" binding:"required"`
}

// UpdateChunkImageInfo 更新分块的图像信息（WeKnora API 兼容）
// 用于存储图像识别结果、OCR 文本等信息
func (s *Service) UpdateChunkImageInfo(ctx context.Context, req *UpdateChunkImageInfoRequest) error {
	chunk, err := s.repo.GetChunkByID(req.ChunkID)
	if err != nil {
		return fmt.Errorf("chunk not found: %w", err)
	}

	// 初始化 Metadata
	if chunk.Metadata == nil {
		chunk.Metadata = make(model.JSON)
	}

	// 更新图像信息
	chunk.Metadata["image_info"] = req.ImageInfo

	return s.repo.UpdateChunkMetadata(req.ChunkID, chunk.Metadata)
}

// DeleteQuestionsByChunk 删除分块关联的问题（WeKnora API 兼容）
// 从分块元数据中删除 questions 字段
func (s *Service) DeleteQuestionsByChunk(ctx context.Context, chunkID string) error {
	return s.repo.DeleteQuestionsFromChunkMetadata(chunkID)
}

// GetChunk 获取单个分块（用于 handler 验证）
func (s *Service) GetChunk(ctx context.Context, chunkID string) (*model.DocumentChunk, error) {
	return s.repo.GetChunkByID(chunkID)
}
