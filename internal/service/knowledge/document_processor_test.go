// Package knowledge 提供 DocumentProcessor 单元测试
package knowledge

import (
	"context"
	"strings"
	"testing"

	"github.com/cloudwego/eino/schema"
	"github.com/cloudwego/eino/components/document/parser"
)

// ========== textParser 测试 ==========

func TestTextParser_Parse(t *testing.T) {
	p := &textParser{}

	tests := []struct {
		name        string
		content     string
		wantDocs    int
		wantContent string
		wantErr     bool
	}{
		{
			name:        "simple text",
			content:     "Hello, world!",
			wantDocs:    1,
			wantContent: "Hello, world!",
			wantErr:     false,
		},
		{
			name:        "multiline text",
			content:     "Line 1\nLine 2\nLine 3",
			wantDocs:    1,
			wantContent: "Line 1\nLine 2\nLine 3",
			wantErr:     false,
		},
		{
			name:        "empty content",
			content:     "",
			wantDocs:    0,
			wantContent: "",
			wantErr:     false,
		},
		{
			name:        "unicode content",
			content:     "Hello 世界 🌍",
			wantDocs:    1,
			wantContent: "Hello 世界 🌍",
			wantErr:     false,
		},
		{
			name:        "large content",
			content:     strings.Repeat("x", 10000),
			wantDocs:    1,
			wantContent: strings.Repeat("x", 10000),
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			reader := strings.NewReader(tt.content)

			docs, err := p.Parse(ctx, reader)

			if tt.wantErr {
				if err == nil {
					t.Error("Parse() expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("Parse() unexpected error: %v", err)
			}
			if len(docs) != tt.wantDocs {
				t.Errorf("Parse() returned %d docs, want %d", len(docs), tt.wantDocs)
			}
			if tt.wantDocs > 0 && docs[0].Content != tt.wantContent {
				t.Errorf("Parse()[0].Content = %q, want %q", docs[0].Content, tt.wantContent)
			}
			if tt.wantDocs > 0 && docs[0].MetaData == nil {
				t.Error("Parse()[0].MetaData is nil")
			}
		})
	}
}

func TestTextParser_Parse_WithOpts(t *testing.T) {
	p := &textParser{}
	ctx := context.Background()
	content := "Test content"
	reader := strings.NewReader(content)

	// 测试 opts 参数（虽然 textParser 不使用它们）
	// 使用 nil opts 切片
	docs, err := p.Parse(ctx, reader)

	if err != nil {
		t.Errorf("Parse() unexpected error: %v", err)
	}
	if len(docs) != 1 {
		t.Errorf("Parse() returned %d docs, want 1", len(docs))
	}
}

// ========== ProcessRequest/ProcessResult 结构测试 ==========

func TestProcessRequest_Struct(t *testing.T) {
	req := &ProcessRequest{
		DocumentID:      "doc123",
		KnowledgeBaseID: "kb456",
	}

	if req.DocumentID != "doc123" {
		t.Errorf("DocumentID = %q, want 'doc123'", req.DocumentID)
	}
	if req.KnowledgeBaseID != "kb456" {
		t.Errorf("KnowledgeBaseID = %q, want 'kb456'", req.KnowledgeBaseID)
	}
}

func TestProcessResult_Struct(t *testing.T) {
	result := &ProcessResult{
		DocumentID: "doc123",
		Success:    true,
		ParsedDocs: 1,
		Chunks:     5,
		Error:      "",
	}

	if result.DocumentID != "doc123" {
		t.Errorf("DocumentID = %q, want 'doc123'", result.DocumentID)
	}
	if !result.Success {
		t.Error("Success = false, want true")
	}
	if result.ParsedDocs != 1 {
		t.Errorf("ParsedDocs = %d, want 1", result.ParsedDocs)
	}
	if result.Chunks != 5 {
		t.Errorf("Chunks = %d, want 5", result.Chunks)
	}
}

func TestProcessResult_WithError(t *testing.T) {
	result := &ProcessResult{
		DocumentID: "doc123",
		Success:    false,
		Error:      "processing failed",
	}

	if result.Success {
		t.Error("Success = true, want false")
	}
	if result.Error != "processing failed" {
		t.Errorf("Error = %q, want 'processing failed'", result.Error)
	}
}

// ========== getFileExt 在 DocumentProcessor 中的使用测试 ==========

func TestGetFileExt_ForProcessor(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		expected string
	}{
		{
			name:     "pdf file",
			filePath: "/path/to/document.pdf",
			expected: ".pdf",
		},
		{
			name:     "docx file",
			filePath: "/path/to/document.docx",
			expected: ".docx",
		},
		{
			name:     "txt file",
			filePath: "/path/to/document.txt",
			expected: ".txt",
		},
		{
			name:     "md file",
			filePath: "/path/to/README.md",
			expected: ".md",
		},
		{
			name:     "html file",
			filePath: "/path/to/page.html",
			expected: ".html",
		},
		{
			name:     "htm file",
			filePath: "/path/to/page.htm",
			expected: ".htm",
		},
		{
			name:     "no extension",
			filePath: "/path/to/file",
			expected: "",
		},
		{
			name:     "unsupported extension",
			filePath: "/path/to/file.xyz",
			expected: ".xyz",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFileExt(tt.filePath)
			if result != tt.expected {
				t.Errorf("getFileExt() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// ========== DocumentProcessor 结构体初始化测试 ==========

func TestNewDocumentProcessor_Nil(t *testing.T) {
	// 测试 nil 参数不会 panic
	proc := NewDocumentProcessor(nil, nil, nil)

	if proc == nil {
		t.Fatal("NewDocumentProcessor() returned nil")
	}
}

// ========== textParser 接口实现测试 ==========

func TestTextParser_ParserInterface(t *testing.T) {
	// 验证 textParser 实现了 parser.Parser 接口
	var _ parser.Parser = &textParser{}

	p := &textParser{}
	ctx := context.Background()
	reader := strings.NewReader("test content")

	// 简单验证接口实现
	docs, err := p.Parse(ctx, reader)
	if err != nil {
		t.Errorf("Parse() unexpected error: %v", err)
	}
	if docs == nil {
		t.Error("Parse() returned nil docs")
	}
}

func TestTextParser_EmptyReader(t *testing.T) {
	p := &textParser{}
	ctx := context.Background()
	reader := strings.NewReader("")

	docs, err := p.Parse(ctx, reader)

	if err != nil {
		t.Errorf("Parse() unexpected error: %v", err)
	}
	if len(docs) != 0 {
		t.Errorf("Parse() returned %d docs, want 0", len(docs))
	}
}

// ========== schema.Document 相关测试 ==========

func TestSchemaDocument_MetaData(t *testing.T) {
	doc := &schema.Document{
		Content: "test content",
		MetaData: map[string]any{
			"document_id":     "doc123",
			"document_title":  "Test Title",
			"file_name":       "test.txt",
			"custom_field":    "custom value",
		},
	}

	if doc.Content != "test content" {
		t.Errorf("Content = %q, want 'test content'", doc.Content)
	}
	if doc.MetaData == nil {
		t.Fatal("MetaData is nil")
	}
	if doc.MetaData["document_id"] != "doc123" {
		t.Errorf("MetaData['document_id'] = %v, want 'doc123'", doc.MetaData["document_id"])
	}
}

func TestSchemaDocument_EmptyMetaData(t *testing.T) {
	doc := &schema.Document{
		Content:  "test content",
		MetaData: nil,
	}

	// 应该允许 nil MetaData
	if doc.Content != "test content" {
		t.Errorf("Content = %q, want 'test content'", doc.Content)
	}
}

// ========== 文档内容边界测试 ==========

func TestTextParser_SpecialCharacters(t *testing.T) {
	p := &textParser{}
	ctx := context.Background()

	tests := []struct {
		name    string
		content string
	}{
		{
			name:    "tabs",
			content: "\t\tIndented text\t",
		},
		{
			name:    "newlines only",
			content: "\n\n\n\n",
		},
		{
			name:    "mixed whitespace",
			content: " \t\n \t\n ",
		},
		{
			name:    "quotes",
			content: `"quoted text" and 'single quotes'`,
		},
		{
			name:    "backslashes",
			content: "path\\to\\file",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.content)
			docs, err := p.Parse(ctx, reader)

			if err != nil {
				t.Errorf("Parse() unexpected error: %v", err)
			}
			if len(docs) == 0 && tt.content != "" && tt.content != "\n\n\n\n" && tt.content != " \t\n \t\n " {
				t.Errorf("Parse() returned 0 docs for non-empty content")
			}
			if len(docs) > 0 && docs[0].Content != tt.content {
				t.Errorf("Parse()[0].Content = %q, want %q", docs[0].Content, tt.content)
			}
		})
	}
}
