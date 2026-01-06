// Package database 提供数据 schema 查询工具
package database

import (
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/ashwinyue/next-ai/internal/model"
	"github.com/ashwinyue/next-ai/internal/repository"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

const (
	ToolDataSchema = "data_schema"
)

// DataSchemaInput 数据 schema 查询输入参数
type DataSchemaInput struct {
	DocumentID string `json:"document_id" jsonschema:"ID of the document (CSV/Excel file) to get schema info"`
}

// SchemaTool 数据 schema 工具
type SchemaTool struct {
	knowledgeRepo *repository.KnowledgeRepository
}

// NewSchemaTool 创建数据 schema 工具
func NewSchemaTool(knowledgeRepo *repository.KnowledgeRepository) *SchemaTool {
	return &SchemaTool{
		knowledgeRepo: knowledgeRepo,
	}
}

// ToolInfo 返回工具信息
func (t *SchemaTool) ToolInfo(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: ToolDataSchema,
		Desc: `获取 CSV/Excel 文件的表结构信息。

## 功能
- 获取表名和列信息
- 获取行数统计
- 列类型和可空性信息

## 使用场景
在执行数据分析查询前，先了解数据的结构和字段信息。

## 使用示例
{"document_id": "xxx"}`,
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"document_id": {
					Type:     schema.String,
					Desc:     "文档 ID (CSV/Excel 文件)",
					Required: true,
				},
			},
		),
	}, nil
}

// InvokableRun 执行工具
func (t *SchemaTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var input DataSchemaInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	// 获取文档信息
	document, err := t.knowledgeRepo.GetDocumentByID(input.DocumentID)
	if err != nil {
		return "", fmt.Errorf("获取文档失败: %v", err)
	}

	// 检查文件类型 (通过扩展名判断)
	ext := strings.ToLower(filepath.Ext(document.FileName))
	if ext != ".csv" && ext != ".xlsx" && ext != ".xls" {
		return "", fmt.Errorf("不支持的文件类型: %s (仅支持 CSV/Excel)", ext)
	}

	// 获取 chunks 中的 schema 信息
	chunks, err := t.knowledgeRepo.GetChunksByDocumentID(input.DocumentID)
	if err != nil {
		return "", fmt.Errorf("获取分块失败: %v", err)
	}

	// 查找 schema 相关的 chunks
	var summaryContent, columnContent string
	for _, chunk := range chunks {
		content := strings.ToLower(chunk.Content)
		if strings.Contains(content, "表名") || strings.Contains(content, "table name") {
			summaryContent = chunk.Content
		} else if strings.Contains(content, "列") || strings.Contains(content, "column") {
			columnContent = chunk.Content
		}
	}

	// 如果没有找到 schema chunks，返回基本信息
	if summaryContent == "" && columnContent == "" {
		return t.generateBasicSchemaInfo(document, chunks), nil
	}

	// 组合输出
	var output strings.Builder
	if summaryContent != "" {
		output.WriteString(summaryContent)
		output.WriteString("\n\n")
	}
	if columnContent != "" {
		output.WriteString(columnContent)
	}

	return output.String(), nil
}

// generateBasicSchemaInfo 生成基本 schema 信息
func (t *SchemaTool) generateBasicSchemaInfo(document *model.Document, chunks []*model.DocumentChunk) string {
	var output strings.Builder

	output.WriteString("=== 数据文件信息 ===\n\n")
	output.WriteString(fmt.Sprintf("文件名: %s\n", document.FileName))
	output.WriteString(fmt.Sprintf("文件路径: %s\n", document.FilePath))
	output.WriteString(fmt.Sprintf("文件大小: %d 字节\n", document.FileSize))
	output.WriteString(fmt.Sprintf("分块数量: %d\n\n", len(chunks)))

	ext := strings.ToLower(filepath.Ext(document.FileName))
	output.WriteString(fmt.Sprintf("文件类型: %s\n\n", ext))

	// 尝试从第一个 chunk 推断列信息
	if len(chunks) > 0 && chunks[0].Content != "" {
		output.WriteString("=== 内容预览 ===\n\n")
		preview := chunks[0].Content
		if len(preview) > 500 {
			preview = preview[:500] + "..."
		}
		output.WriteString(preview)
		output.WriteString("\n\n提示: 使用 data_analysis 工具执行 SQL 查询以获取详细结构信息\n")
		output.WriteString(fmt.Sprintf("例如: {\"document_id\": \"%s\", \"sql\": \"PRAGMA table_info('d_%s')\"}\n",
			document.ID, strings.ReplaceAll(document.ID, "-", "_")))
	}

	return output.String()
}

// String 类型工具的字符串表示
func (t *SchemaTool) String() string {
	return ToolDataSchema
}
