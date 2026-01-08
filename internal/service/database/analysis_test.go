// Package database 提供数据分析功能单元测试
package database

import (
	"context"
	"strings"
	"testing"
)

// ========== NewAnalysisTool 测试 ==========

func TestNewAnalysisTool(t *testing.T) {
	// nil repo 允许（仅测试结构）
	tool := NewAnalysisTool(nil, "test-session")

	if tool == nil {
		t.Fatal("NewAnalysisTool() returned nil")
	}
	if tool.sessionID != "test-session" {
		t.Errorf("sessionID = %q, want 'test-session'", tool.sessionID)
	}
}

func TestNewAnalysisTool_EmptySessionID(t *testing.T) {
	tool := NewAnalysisTool(nil, "")

	if tool == nil {
		t.Fatal("NewAnalysisTool() returned nil")
	}
	if tool.sessionID != "" {
		t.Errorf("sessionID should be empty, got %q", tool.sessionID)
	}
}

// ========== ToolInfo 测试 ==========

func TestAnalysisTool_ToolInfo(t *testing.T) {
	tool := NewAnalysisTool(nil, "test-session")
	ctx := context.Background()

	info, err := tool.ToolInfo(ctx)

	if err != nil {
		t.Errorf("ToolInfo() unexpected error: %v", err)
	}
	if info == nil {
		t.Fatal("ToolInfo() returned nil")
	}
	if info.Name != ToolDataAnalysis {
		t.Errorf("ToolInfo().Name = %q, want %q", info.Name, ToolDataAnalysis)
	}
	if info.Desc == "" {
		t.Error("ToolInfo().Desc should not be empty")
	}
	if info.ParamsOneOf == nil {
		t.Error("ToolInfo().ParamsOneOf should not be nil")
	}
}

// ========== String 测试 ==========

func TestAnalysisTool_String(t *testing.T) {
	tool := NewAnalysisTool(nil, "test-session")

	result := tool.String()

	if result != ToolDataAnalysis {
		t.Errorf("String() = %q, want %q", result, ToolDataAnalysis)
	}
}

// ========== AnalysisSession 测试 ==========

func TestAnalysisSession_Close_NilDB(t *testing.T) {
	session := &AnalysisSession{
		db: nil,
	}

	err := session.Close()

	if err != nil {
		t.Errorf("Close() with nil db should not error, got: %v", err)
	}
}

func TestAnalysisSession_Cleanup_EmptyTables(t *testing.T) {
	ctx := context.Background()
	session := &AnalysisSession{
		db:            nil,
		createdTables: []string{},
	}

	// 不应该 panic
	session.Cleanup(ctx)

	if len(session.createdTables) != 0 {
		t.Error("Cleanup() should clear createdTables")
	}
}

// ========== GetAnalysisSession 测试 ==========

func TestGetAnalysisSession_NilDB(t *testing.T) {
	// 创建新会话
	session := GetAnalysisSession("test-new-session")

	if session == nil {
		t.Error("GetAnalysisSession() returned nil")
	}

	// 清理
	analysisSessionManager.mu.Lock()
	delete(analysisSessionManager.sessions, "test-new-session")
	analysisSessionManager.mu.Unlock()
}

func TestGetAnalysisSession_Existing(t *testing.T) {
	// 第一次获取
	session1 := GetAnalysisSession("test-existing")

	// 第二次获取应该返回相同实例
	session2 := GetAnalysisSession("test-existing")

	if session1 != session2 {
		t.Error("GetAnalysisSession() should return same session for same ID")
	}

	// 清理
	analysisSessionManager.mu.Lock()
	delete(analysisSessionManager.sessions, "test-existing")
	analysisSessionManager.mu.Unlock()
}

// ========== formatQueryResults 测试 ==========

func TestFormatQueryResults_Empty(t *testing.T) {
	tool := NewAnalysisTool(nil, "test-session")

	results := []map[string]string{}
	query := "SELECT * FROM test"

	result := tool.formatQueryResults(results, query)

	if !strings.Contains(result, "未找到匹配的记录") {
		t.Error("formatQueryResults() should contain '未找到匹配的记录' for empty results")
	}
}

func TestFormatQueryResults_SingleRecord(t *testing.T) {
	tool := NewAnalysisTool(nil, "test-session")

	results := []map[string]string{
		{"col1": "value1", "col2": "value2"},
	}
	query := "SELECT * FROM test"

	result := tool.formatQueryResults(results, query)

	if !strings.Contains(result, "返回 1 行数据") {
		t.Error("formatQueryResults() should contain row count")
	}
	if !strings.Contains(result, "value1") {
		t.Error("formatQueryResults() should contain data")
	}
}

func TestFormatQueryResults_MultipleRecords(t *testing.T) {
	tool := NewAnalysisTool(nil, "test-session")

	results := make([]map[string]string, 15)
	for i := 0; i < 15; i++ {
		results[i] = map[string]string{"id": string(rune('a' + i))}
	}
	query := "SELECT * FROM test"

	result := tool.formatQueryResults(results, query)

	if !strings.Contains(result, "建议使用 LIMIT") {
		t.Error("formatQueryResults() should suggest LIMIT for large results")
	}
}

func TestFormatQueryResults_WithBytes(t *testing.T) {
	tool := NewAnalysisTool(nil, "test-session")

	results := []map[string]string{
		{"col1": string([]byte("byte_value"))},
	}
	query := "SELECT col1 FROM test"

	result := tool.formatQueryResults(results, query)

	if !strings.Contains(result, "byte_value") {
		t.Error("formatQueryResults() should handle byte values")
	}
}

// ========== Input 结构测试 ==========

func TestDataAnalysisInput(t *testing.T) {
	input := DataAnalysisInput{
		DocumentID: "doc123",
		SQL:        "SELECT * FROM test",
	}

	if input.DocumentID != "doc123" {
		t.Errorf("DocumentID = %q, want 'doc123'", input.DocumentID)
	}
	if input.SQL != "SELECT * FROM test" {
		t.Errorf("SQL = %q, want 'SELECT * FROM test'", input.SQL)
	}
}
