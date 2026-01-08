// Package database 提供数据库数据分析工具，使用 DuckDB 进行内存数据分析
package database

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ashwinyue/next-ai/internal/model"
	"github.com/ashwinyue/next-ai/internal/repository"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	_ "github.com/duckdb/duckdb-go/v2"
)

const (
	ToolDataAnalysis = "data_analysis"
)

// DataAnalysisInput 数据分析输入参数
type DataAnalysisInput struct {
	DocumentID string `json:"document_id" jsonschema:"ID of the document (CSV/Excel file) to query"`
	SQL        string `json:"sql" jsonschema:"SQL to be executed on the loaded data"`
}

// AnalysisSession 数据分析会话
type AnalysisSession struct {
	db            *sql.DB
	createdTables []string
	mu            sync.Mutex
}

// analysisSessionManager 会话管理器
var analysisSessionManager = struct {
	sessions map[string]*AnalysisSession
	mu       sync.RWMutex
}{
	sessions: make(map[string]*AnalysisSession),
}

// GetAnalysisSession 获取或创建分析会话
func GetAnalysisSession(sessionID string) *AnalysisSession {
	analysisSessionManager.mu.Lock()
	defer analysisSessionManager.mu.Unlock()

	if s, ok := analysisSessionManager.sessions[sessionID]; ok {
		return s
	}

	// 创建新的 DuckDB 连接
	db, err := sql.Open("duckdb", "")
	if err != nil {
		return nil
	}

	s := &AnalysisSession{
		db:            db,
		createdTables: make([]string, 0),
	}
	analysisSessionManager.sessions[sessionID] = s
	return s
}

// Cleanup 清理会话
func (s *AnalysisSession) Cleanup(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, tableName := range s.createdTables {
		dropSQL := fmt.Sprintf("DROP TABLE IF EXISTS \"%s\"", tableName)
		s.db.ExecContext(ctx, dropSQL)
	}
	s.createdTables = nil
}

// Close 关闭连接
func (s *AnalysisSession) Close() error {
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// AnalysisTool 数据分析工具
type AnalysisTool struct {
	knowledgeRepo repository.KnowledgeRepository // 使用接口
	sessionID     string
}

// NewAnalysisTool 创建数据分析工具
func NewAnalysisTool(knowledgeRepo repository.KnowledgeRepository, sessionID string) *AnalysisTool {
	return &AnalysisTool{
		knowledgeRepo: knowledgeRepo,
		sessionID:     sessionID,
	}
}

// ToolInfo 返回工具信息
func (t *AnalysisTool) ToolInfo(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: ToolDataAnalysis,
		Desc: `使用 DuckDB 对 CSV/Excel 文件进行数据分析。

## 功能
- 自动加载 CSV/Excel 文件到内存
- 执行 SQL 查询进行数据分析
- 支持聚合、排序、过滤等操作

## 使用示例

统计列值分布:
{"document_id": "xxx", "sql": "SELECT category, COUNT(*) as count FROM d_xxx GROUP BY category"}

查找最大值:
{"document_id": "xxx", "sql": "SELECT * FROM d_xxx ORDER BY value DESC LIMIT 5"}

## 注意事项
- 只允许只读查询 (SELECT, SHOW, DESCRIBE, EXPLAIN, PRAGMA)
- 表名自动生成为 d_<document_id> (横杠替换为下划线)
- 建议使用 LIMIT 限制结果数量`,
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"document_id": {
					Type:     schema.String,
					Desc:     "文档 ID (CSV/Excel 文件)",
					Required: true,
				},
				"sql": {
					Type:     schema.String,
					Desc:     "要执行的 SQL 查询",
					Required: true,
				},
			},
		),
	}, nil
}

// InvokableRun 执行工具
func (t *AnalysisTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var input DataAnalysisInput
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

	// 获取分析会话
	session := GetAnalysisSession(t.sessionID)
	if session == nil {
		return "", fmt.Errorf("创建分析会话失败")
	}

	// 生成表名
	tableName := fmt.Sprintf("d_%s", strings.ReplaceAll(input.DocumentID, "-", "_"))

	// 加载数据到表
	if err := t.loadDocumentToTable(ctx, session, document, tableName, ext); err != nil {
		return "", fmt.Errorf("加载数据失败: %v", err)
	}

	// 替换 SQL 中的 document_id 为表名
	input.SQL = strings.ReplaceAll(input.SQL, input.DocumentID, tableName)

	// 验证是只读查询
	normalizedSQL := strings.TrimSpace(strings.ToLower(input.SQL))
	isReadOnly := strings.HasPrefix(normalizedSQL, "select") ||
		strings.HasPrefix(normalizedSQL, "show") ||
		strings.HasPrefix(normalizedSQL, "describe") ||
		strings.HasPrefix(normalizedSQL, "explain") ||
		strings.HasPrefix(normalizedSQL, "pragma")

	if !isReadOnly {
		return "", fmt.Errorf("只允许只读查询 (SELECT, SHOW, DESCRIBE, EXPLAIN, PRAGMA)")
	}

	// 执行查询
	results, err := t.executeQuery(ctx, session, input.SQL)
	if err != nil {
		return "", fmt.Errorf("查询执行失败: %v", err)
	}

	return t.formatQueryResults(results, input.SQL), nil
}

// loadDocumentToTable 加载文档数据到表
func (t *AnalysisTool) loadDocumentToTable(ctx context.Context, session *AnalysisSession, document *model.Document, tableName string, ext string) error {
	session.mu.Lock()
	defer session.mu.Unlock()

	// 检查表是否已存在
	for _, name := range session.createdTables {
		if name == tableName {
			return nil // 表已存在
		}
	}

	var createSQL string
	switch ext {
	case ".csv":
		createSQL = fmt.Sprintf("CREATE TABLE \"%s\" AS SELECT * FROM read_csv_auto('%s')", tableName, document.FilePath)
	case ".xlsx", ".xls":
		// 使用 st_read 读取 Excel (需要 spatial 扩展)
		createSQL = fmt.Sprintf("CREATE TABLE \"%s\" AS SELECT * FROM st_read('%s')", tableName, document.FilePath)
	default:
		return fmt.Errorf("不支持的文件类型: %s", ext)
	}

	_, err := session.db.ExecContext(ctx, createSQL)
	if err != nil {
		return fmt.Errorf("创建表失败: %w", err)
	}

	session.createdTables = append(session.createdTables, tableName)
	return nil
}

// executeQuery 执行查询
func (t *AnalysisTool) executeQuery(ctx context.Context, session *AnalysisSession, sqlQuery string) ([]map[string]string, error) {
	rows, err := session.db.QueryContext(ctx, sqlQuery)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	results := make([]map[string]string, 0)
	for rows.Next() {
		columnValues := make([]interface{}, len(columns))
		columnPointers := make([]interface{}, len(columns))
		for i := range columnValues {
			columnPointers[i] = &columnValues[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return nil, err
		}

		rowMap := make(map[string]string)
		for i, colName := range columns {
			val := columnValues[i]
			if b, ok := val.([]byte); ok {
				rowMap[colName] = string(b)
			} else {
				rowMap[colName] = fmt.Sprintf("%v", val)
			}
		}
		results = append(results, rowMap)
	}

	return results, nil
}

// formatQueryResults 格式化查询结果
func (t *AnalysisTool) formatQueryResults(results []map[string]string, query string) string {
	var output strings.Builder

	output.WriteString("=== DuckDB 查询结果 ===\n\n")
	output.WriteString(fmt.Sprintf("执行的 SQL: %s\n\n", query))
	output.WriteString(fmt.Sprintf("返回 %d 行数据\n\n", len(results)))

	if len(results) == 0 {
		output.WriteString("未找到匹配的记录。\n")
		return output.String()
	}

	output.WriteString("=== 数据详情 ===\n\n")
	if len(results) > 10 {
		output.WriteString(fmt.Sprintf("显示了全部 %d 条记录。建议使用 LIMIT 子句限制结果数量。\n\n", len(results)))
	}

	for i, record := range results {
		recordBytes, _ := json.Marshal(record)
		recordStr := strings.Trim(string(recordBytes), "\n")
		output.WriteString(fmt.Sprintf("记录 %d: %s\n", i+1, recordStr))
	}

	return output.String()
}

// String 类型工具的字符串表示
func (t *AnalysisTool) String() string {
	return ToolDataAnalysis
}
