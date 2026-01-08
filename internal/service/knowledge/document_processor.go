// Package knowledge provides document processing service
// 参考 next-ai/docs/eino-integration-guide.md
// 直接使用 eino/eino-ext 组件，避免冗余封装
package knowledge

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/ashwinyue/next-ai/internal/config"
	"github.com/ashwinyue/next-ai/internal/model"
	"github.com/ashwinyue/next-ai/internal/repository"
	"github.com/cloudwego/eino-ext/components/document/parser/docx"
	"github.com/cloudwego/eino-ext/components/document/parser/html"
	"github.com/cloudwego/eino-ext/components/document/parser/pdf"
	"github.com/cloudwego/eino-ext/components/document/transformer/splitter/recursive"
	einoparser "github.com/cloudwego/eino/components/document/parser"
	"github.com/cloudwego/eino/components/embedding"
	"github.com/cloudwego/eino/schema"
	"github.com/google/uuid"
)

// DocumentProcessor 文档处理服务
type DocumentProcessor struct {
	repo     *repository.Repositories
	cfg      *config.Config
	embedder embedding.Embedder
}

// NewDocumentProcessor 创建文档处理器
func NewDocumentProcessor(repo *repository.Repositories, cfg *config.Config, embedder embedding.Embedder) *DocumentProcessor {
	return &DocumentProcessor{
		repo:     repo,
		cfg:      cfg,
		embedder: embedder,
	}
}

// ProcessRequest 处理请求
type ProcessRequest struct {
	DocumentID      string `json:"document_id"`
	KnowledgeBaseID string `json:"knowledge_base_id"`
}

// ProcessResult 处理结果
type ProcessResult struct {
	DocumentID string        `json:"document_id"`
	Success    bool          `json:"success"`
	ParsedDocs int           `json:"parsed_docs"`
	Chunks     int           `json:"chunks"`
	Duration   time.Duration `json:"duration"`
	Error      string        `json:"error,omitempty"`
}

// Process 处理文档的完整流程
func (p *DocumentProcessor) Process(ctx context.Context, req *ProcessRequest) (*ProcessResult, error) {
	startTime := time.Now()
	result := &ProcessResult{DocumentID: req.DocumentID}

	// 获取文档和知识库
	doc, err := p.repo.Knowledge.GetDocumentByID(req.DocumentID)
	if err != nil {
		result.Error = fmt.Sprintf("文档不存在: %v", err)
		return result, fmt.Errorf("document not found: %w", err)
	}

	kb, err := p.repo.Knowledge.GetKnowledgeBaseByID(req.KnowledgeBaseID)
	if err != nil {
		result.Error = fmt.Sprintf("知识库不存在: %v", err)
		return result, fmt.Errorf("knowledge base not found: %w", err)
	}

	// 解析文档
	parsedDocs, err := p.parseDocument(ctx, doc)
	if err != nil {
		result.Error = fmt.Sprintf("解析失败: %v", err)
		return result, fmt.Errorf("failed to parse document: %w", err)
	}
	result.ParsedDocs = len(parsedDocs)
	if result.ParsedDocs == 0 {
		result.Error = "解析后文档为空"
		return result, fmt.Errorf("no content parsed from document")
	}

	// 分块
	chunks, err := p.splitDocuments(ctx, parsedDocs, doc, kb)
	if err != nil {
		result.Error = fmt.Sprintf("分块失败: %v", err)
		return result, fmt.Errorf("failed to split documents: %w", err)
	}
	result.Chunks = len(chunks)

	// 向量化
	vectors, err := p.embedChunks(ctx, chunks)
	if err != nil {
		result.Error = fmt.Sprintf("向量化失败: %v", err)
		return result, fmt.Errorf("failed to embed chunks: %w", err)
	}

	// 索引
	if err := p.indexChunks(ctx, kb, doc, chunks, vectors); err != nil {
		result.Error = fmt.Sprintf("索引失败: %v", err)
		return result, fmt.Errorf("failed to index chunks: %w", err)
	}

	// 更新文档状态
	doc.Status = "completed"
	doc.ChunkCount = len(chunks)
	if err := p.repo.Knowledge.UpdateDocument(doc); err != nil {
		log.Printf("Warning: failed to update document status: %v", err)
	}

	result.Duration = time.Since(startTime)
	result.Success = true
	return result, nil
}

// parseDocument 解析文档
func (p *DocumentProcessor) parseDocument(ctx context.Context, doc *model.Document) ([]*schema.Document, error) {
	if doc.FilePath == "" {
		return nil, fmt.Errorf("file path is empty")
	}

	fileParser, err := p.newParser(ctx, doc.FilePath)
	if err != nil {
		return nil, err
	}

	file, err := os.Open(doc.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	docs, err := fileParser.Parse(ctx, file)
	if err != nil {
		return nil, fmt.Errorf("parser failed: %w", err)
	}

	// 添加元数据
	for _, d := range docs {
		if d.MetaData == nil {
			d.MetaData = make(map[string]any)
		}
		d.MetaData["document_id"] = doc.ID
		d.MetaData["document_title"] = doc.Title
		d.MetaData["file_name"] = doc.FileName
	}

	return docs, nil
}

// newParser 创建解析器
func (p *DocumentProcessor) newParser(ctx context.Context, filePath string) (einoparser.Parser, error) {
	ext := getFileExt(filePath)

	switch ext {
	case ".pdf":
		return pdf.NewPDFParser(ctx, &pdf.Config{ToPages: false})
	case ".docx":
		return docx.NewDocxParser(ctx, &docx.Config{
			ToSections:      false,
			IncludeComments: false,
			IncludeHeaders:  true,
			IncludeFooters:  false,
			IncludeTables:   true,
		})
	case ".html", ".htm":
		// 使用 body 选择器提取正文内容
		bodySelector := "body"
		return html.NewParser(ctx, &html.Config{
			Selector: &bodySelector,
		})
	case ".txt", ".md":
		return &textParser{}, nil
	default:
		return nil, fmt.Errorf("unsupported file type: %s", ext)
	}
}

// textParser 纯文本解析器
type textParser struct{}

func (p *textParser) Parse(_ context.Context, reader io.Reader, opts ...einoparser.Option) ([]*schema.Document, error) {
	content, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read: %w", err)
	}

	text := string(content)
	if text == "" {
		return []*schema.Document{}, nil
	}

	return []*schema.Document{
		{
			Content:  text,
			MetaData: make(map[string]any),
		},
	}, nil
}

// splitDocuments 分块文档
func (p *DocumentProcessor) splitDocuments(ctx context.Context, docs []*schema.Document, doc *model.Document, kb *model.KnowledgeBase) ([]*model.DocumentChunk, error) {
	splitter, err := recursive.NewSplitter(ctx, &recursive.Config{
		ChunkSize:   512,
		OverlapSize: 50,
		Separators:  []string{"\n\n", "\n", ". ", "。", "? ", "？", "! ", "！", ", ", "，", " ", ""},
		KeepType:    recursive.KeepTypeNone,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create splitter: %w", err)
	}

	splitDocs, err := splitter.Transform(ctx, docs)
	if err != nil {
		return nil, fmt.Errorf("splitter failed: %w", err)
	}

	chunks := make([]*model.DocumentChunk, 0, len(splitDocs))
	for i, splitDoc := range splitDocs {
		metadata := model.JSON{
			"chunk_index":         i,
			"document_id":         doc.ID,
			"document_title":      doc.Title,
			"knowledge_base_id":   kb.ID,
			"knowledge_base_name": kb.Name,
		}

		chunk := &model.DocumentChunk{
			ID:              uuid.New().String(),
			DocumentID:      doc.ID,
			KnowledgeBaseID: kb.ID,
			ChunkIndex:      i,
			Content:         splitDoc.Content,
			Metadata:        metadata,
		}
		chunks = append(chunks, chunk)
	}

	return chunks, nil
}

// embedChunks 向量化文档块
func (p *DocumentProcessor) embedChunks(ctx context.Context, chunks []*model.DocumentChunk) ([][]float64, error) {
	if p.embedder == nil {
		return nil, fmt.Errorf("embedder not configured")
	}

	texts := make([]string, len(chunks))
	for i, chunk := range chunks {
		texts[i] = chunk.Content
	}

	vectors, err := p.embedder.EmbedStrings(ctx, texts)
	if err != nil {
		return nil, fmt.Errorf("embed strings failed: %w", err)
	}

	if len(vectors) != len(chunks) {
		return nil, fmt.Errorf("vector count mismatch: expected %d, got %d", len(chunks), len(vectors))
	}

	return vectors, nil
}

// indexChunks 索引文档块（使用 Eino ES8 Indexer）
func (p *DocumentProcessor) indexChunks(ctx context.Context, kb *model.KnowledgeBase, doc *model.Document, chunks []*model.DocumentChunk, vectors [][]float64) error {
	// 创建 Eino ES8 Indexer
	indexer, err := NewES8Indexer(ctx, p.cfg, p.embedder)
	if err != nil {
		return fmt.Errorf("failed to create ES8 indexer: %w", err)
	}
	defer indexer.Close()

	// 确保索引存在
	if err := indexer.EnsureIndex(ctx); err != nil {
		return fmt.Errorf("failed to ensure index: %w", err)
	}

	// 保存到数据库
	if err := p.repo.Knowledge.CreateChunks(chunks); err != nil {
		return fmt.Errorf("failed to save chunks to database: %w", err)
	}

	// 转换为 Eino Document 并索引
	einoDocs := ChunksToEinoDocuments(chunks)
	ids, err := indexer.Store(ctx, einoDocs)
	if err != nil {
		return fmt.Errorf("failed to store documents: %w", err)
	}

	log.Printf("Indexed %d chunks to ES", len(ids))
	return nil
}

// 辅助函数
func getFileExt(filePath string) string {
	for i := len(filePath) - 1; i >= 0; i-- {
		if filePath[i] == '.' {
			return filePath[i:]
		}
	}
	return ""
}

// CreateChunkIndex 创建文档块索引（使用 Eino ES8 Indexer）
func CreateChunkIndex(ctx context.Context, cfg *config.Config, embedder embedding.Embedder) error {
	indexer, err := NewES8Indexer(ctx, cfg, embedder)
	if err != nil {
		return err
	}
	defer indexer.Close()

	return indexer.EnsureIndex(ctx)
}
