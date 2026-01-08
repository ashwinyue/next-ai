// Package database 提供数据 Schema 功能单元测试
package database

import (
	"context"
	"strings"
	"testing"

	"github.com/ashwinyue/next-ai/internal/model"
)

// ========== NewSchemaTool 测试 ==========

func TestNewSchemaTool(t *testing.T) {
	// nil repo 允许（仅测试结构）
	tool := NewSchemaTool(nil)

	if tool == nil {
		t.Fatal("NewSchemaTool() returned nil")
	}
}

// ========== ToolInfo 测试 ==========

func TestSchemaTool_ToolInfo(t *testing.T) {
	tool := NewSchemaTool(nil)
	ctx := context.Background()

	info, err := tool.ToolInfo(ctx)

	if err != nil {
		t.Errorf("ToolInfo() unexpected error: %v", err)
	}
	if info == nil {
		t.Fatal("ToolInfo() returned nil")
	}
	if info.Name != ToolDataSchema {
		t.Errorf("ToolInfo().Name = %q, want %q", info.Name, ToolDataSchema)
	}
	if info.Desc == "" {
		t.Error("ToolInfo().Desc should not be empty")
	}
	if info.ParamsOneOf == nil {
		t.Error("ToolInfo().ParamsOneOf should not be nil")
	}
}

// ========== String 测试 ==========

func TestSchemaTool_String(t *testing.T) {
	tool := NewSchemaTool(nil)

	result := tool.String()

	if result != ToolDataSchema {
		t.Errorf("String() = %q, want %q", result, ToolDataSchema)
	}
}

// ========== generateBasicSchemaInfo 测试 ==========

func TestGenerateBasicSchemaInfo_EmptyChunks(t *testing.T) {
	tool := NewSchemaTool(nil)

	document := model.Document{
		ID:       "doc123",
		FileName: "test.csv",
		FilePath: "/path/to/test.csv",
		FileSize: 1024,
	}

	var chunks []*model.DocumentChunk

	result := tool.generateBasicSchemaInfo(&document, chunks)

	if !strings.Contains(result, "test.csv") {
		t.Error("generateBasicSchemaInfo() should contain filename")
	}
	if !strings.Contains(result, "1024 字节") {
		t.Error("generateBasicSchemaInfo() should contain file size")
	}
	if !strings.Contains(result, "0") || !strings.Contains(result, "分块数量") {
		t.Error("generateBasicSchemaInfo() should show 0 chunks")
	}
}

func TestGenerateBasicSchemaInfo_WithChunks(t *testing.T) {
	tool := NewSchemaTool(nil)

	document := model.Document{
		ID:       "doc123",
		FileName: "data.csv",
		FilePath: "/path/to/data.csv",
		FileSize: 2048,
	}

	chunks := []*model.DocumentChunk{
		{Content: "col1,col2,col3\n1,2,3\n4,5,6\n"},
	}

	result := tool.generateBasicSchemaInfo(&document, chunks)

	if !strings.Contains(result, "1") || !strings.Contains(result, "分块数量") {
		t.Error("generateBasicSchemaInfo() should show 1 chunk")
	}
	if !strings.Contains(result, "内容预览") {
		t.Error("generateBasicSchemaInfo() should show content preview")
	}
	if !strings.Contains(result, "col1") {
		t.Error("generateBasicSchemaInfo() should contain chunk content")
	}
	if !strings.Contains(result, "PRAGMA table_info") {
		t.Error("generateBasicSchemaInfo() should suggest PRAGMA query")
	}
}

func TestGenerateBasicSchemaInfo_LongContent(t *testing.T) {
	tool := NewSchemaTool(nil)

	document := model.Document{
		ID:       "doc123",
		FileName: "large.csv",
		FilePath: "/path/to/large.csv",
		FileSize: 9999,
	}

	// 创建超过 500 字符的内容
	longContent := ""
	for i := 0; i < 600; i++ {
		longContent += "x"
	}

	chunks := []*model.DocumentChunk{
		{Content: longContent},
	}

	result := tool.generateBasicSchemaInfo(&document, chunks)

	if !strings.Contains(result, "...") {
		t.Error("generateBasicSchemaInfo() should truncate long content")
	}
}

func TestGenerateBasicSchemaInfo_XLSXFile(t *testing.T) {
	tool := NewSchemaTool(nil)

	document := model.Document{
		ID:       "doc123",
		FileName: "data.xlsx",
		FilePath: "/path/to/data.xlsx",
		FileSize: 4096,
	}

	var chunks []*model.DocumentChunk

	result := tool.generateBasicSchemaInfo(&document, chunks)

	if !strings.Contains(result, ".xlsx") {
		t.Error("generateBasicSchemaInfo() should show file type")
	}
}

func TestGenerateBasicSchemaInfo_HyphenatedID(t *testing.T) {
	tool := NewSchemaTool(nil)

	document := model.Document{
		ID:       "doc-with-hyphens",
		FileName: "test.csv",
		FilePath: "/path/to/test.csv",
		FileSize: 1024,
	}

	// 需要有 chunk 才会输出 PRAGMA 建议
	chunks := []*model.DocumentChunk{
		{Content: "some content"},
	}

	result := tool.generateBasicSchemaInfo(&document, chunks)

	// PRAGMA 查询中的表名应该把横杠替换为下划线
	if !strings.Contains(result, "d_doc_with_hyphens") {
		t.Error("generateBasicSchemaInfo() should convert hyphens to underscores in table name")
	}
}

// ========== Input 结构测试 ==========

func TestDataSchemaInput(t *testing.T) {
	input := DataSchemaInput{
		DocumentID: "doc123",
	}

	if input.DocumentID != "doc123" {
		t.Errorf("DocumentID = %q, want 'doc123'", input.DocumentID)
	}
}
