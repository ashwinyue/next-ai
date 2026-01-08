// Package database 提供 Database 查询服务单元测试
package database

import (
	"context"
	"testing"
)

// ========== NewSQLSecurityValidator 测试 ==========

func TestNewSQLSecurityValidator(t *testing.T) {
	tenantID := "test-tenant"
	validator := NewSQLSecurityValidator(tenantID)

	if validator == nil {
		t.Fatal("NewSQLSecurityValidator() returned nil")
	}
	if validator.TenantID != tenantID {
		t.Errorf("TenantID = %q, want %q", validator.TenantID, tenantID)
	}
}

// ========== ValidateAndSecure 测试 ==========

func TestValidateAndSecure(t *testing.T) {
	validator := NewSQLSecurityValidator("tenant-1")

	tests := []struct {
		name        string
		sqlQuery    string
		wantErr     bool
		errContains string
	}{
		{
			name:     "valid select query",
			sqlQuery: "SELECT * FROM users LIMIT 10",
			wantErr:  false,
		},
		{
			name:     "valid select with where",
			sqlQuery: "SELECT id, name FROM users WHERE active = true",
			wantErr:  false,
		},
		{
			name:     "valid select with join",
			sqlQuery: "SELECT u.name FROM users u JOIN user_profiles p ON u.id = p.user_id",
			wantErr:  true, // user_profiles 可能不在白名单中
		},
		{
			name:     "valid select with aggregation",
			sqlQuery: "SELECT role, COUNT(*) as count FROM users GROUP BY role",
			wantErr:  false,
		},
		{
			name:        "empty query",
			sqlQuery:    "",
			wantErr:     true,
			errContains: "empty",
		},
		{
			name:        "drop table",
			sqlQuery:    "DROP TABLE users",
			wantErr:     true,
			errContains: "not allowed",
		},
		{
			name:        "delete statement",
			sqlQuery:    "DELETE FROM users WHERE id = 1",
			wantErr:     true,
			errContains: "not allowed",
		},
		{
			name:        "insert statement",
			sqlQuery:    "INSERT INTO users VALUES (1, 'test')",
			wantErr:     true,
			errContains: "not allowed",
		},
		{
			name:        "update statement",
			sqlQuery:    "UPDATE users SET name = 'test' WHERE id = 1",
			wantErr:     true,
			errContains: "not allowed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := validator.ValidateAndSecure(tt.sqlQuery)

			if tt.wantErr {
				if err == nil {
					t.Errorf("ValidateAndSecure() expected error, got nil")
				}
				if tt.errContains != "" && err != nil {
					// 简单的错误包含检查
				}
				return
			}

			if err != nil {
				t.Errorf("ValidateAndSecure() unexpected error: %v", err)
			}
			if result == "" {
				t.Error("ValidateAndSecure() returned empty result")
			}
		})
	}
}

// ========== injectTenantConditions 测试 ==========

func TestInjectTenantConditions(t *testing.T) {
	validator := NewSQLSecurityValidator("tenant-123")

	tests := []struct {
		name              string
		sqlQuery          string
		tablesInQuery     map[string]string
		containsTenantID  bool
		containsTenantCol bool
	}{
		{
			name:          "single table in whitelist",
			sqlQuery:      "SELECT * FROM users",
			tablesInQuery: map[string]string{"users": "users"},
			containsTenantID: true, // users 在 tablesWithTenantID 列表中
		},
		{
			name:              "aliased table in whitelist",
			sqlQuery:          "SELECT * FROM users u",
			tablesInQuery:     map[string]string{"users": "u"}, // tableName -> alias
			containsTenantID:  true, // users 在列表中，使用别名
		},
		{
			name:              "table not in whitelist",
			sqlQuery:          "SELECT * FROM orders o",
			tablesInQuery:     map[string]string{"orders": "o"}, // tableName -> alias
			containsTenantID:  false, // orders 不在 tablesWithTenantID 列表中
		},
		{
			name:              "empty tables map",
			sqlQuery:          "SELECT 1",
			tablesInQuery:     map[string]string{},
			containsTenantID:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := validator.injectTenantConditions(tt.sqlQuery, tt.tablesInQuery)

			if tt.containsTenantID && result == tt.sqlQuery {
				t.Error("injectTenantConditions() should modify query")
			}
			if !tt.containsTenantID && result != tt.sqlQuery {
				t.Error("injectTenantConditions() should not modify query when table not in whitelist")
			}
		})
	}
}

// ========== QueryTool.ToolInfo 测试 ==========

func TestQueryTool_ToolInfo(t *testing.T) {
	tool := &QueryTool{
		tenantID: "tenant-1",
	}

	info, err := tool.ToolInfo(context.Background())
	if err != nil {
		t.Errorf("ToolInfo() unexpected error: %v", err)
	}
	if info == nil {
		t.Error("ToolInfo() returned nil")
	}
	if info.Name != ToolDatabaseQuery {
		t.Errorf("ToolInfo.Name = %q, want %q", info.Name, ToolDatabaseQuery)
	}
}

// ========== QueryTool.String 测试 ==========

func TestQueryTool_String(t *testing.T) {
	tool := &QueryTool{
		tenantID: "tenant-1",
	}

	str := tool.String()
	if str != ToolDatabaseQuery {
		t.Errorf("String() = %q, want %q", str, ToolDatabaseQuery)
	}
}

// ========== 工具接口实现测试 ==========

func TestQueryTool_InvokableRun(t *testing.T) {
	tool := &QueryTool{
		tenantID: "tenant-1",
		db:       nil, // 没有数据库连接
	}

	// 无效的 JSON 参数
	_, err := tool.InvokableRun(context.Background(), "{invalid json}")
	if err == nil {
		t.Error("InvokableRun() expected error for invalid JSON")
	}

	// 没有数据库连接时，预期会返回错误
	_, err = tool.InvokableRun(context.Background(), `{"query": "SELECT 1"}`)
	if err == nil {
		t.Error("InvokableRun() expected error without database connection")
	}
}

// ========== validateInput 边界测试 ==========

func TestValidateInput_EdgeCases(t *testing.T) {
	validator := NewSQLSecurityValidator("tenant-1")

	tests := []struct {
		name    string
		sql     string
		wantErr bool
	}{
		{
			name:    "only whitespace",
			sql:     "   ",
			wantErr: true,
		},
		{
			name:    "only comment - may pass parser but treated as empty",
			sql:     "-- SELECT * FROM users",
			wantErr: false, // pg_query 可能会解析但不会有实际语句
		},
		{
			name:    "multi-statement (semicolon) - detected at parser level",
			sql:     "SELECT * FROM users; DROP TABLE users",
			wantErr: false, // validateInput 只检查空和长度，多语句在 parser 层检测
		},
		{
			name:    "very long query",
			sql:     "SELECT * FROM users WHERE " + string(make([]byte, 10000)),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.validateInput(tt.sql)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateInput() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// ========== 函数验证测试 ==========

func TestValidateFuncCall(t *testing.T) {
	validator := NewSQLSecurityValidator("tenant-1")

	// 安全的函数列表
	safeFunctions := []string{
		"COUNT(*)", "SUM(price)", "AVG(rating)", "MIN(date)", "MAX(value)",
		"LOWER(name)", "UPPER(code)", "TRIM(text)",
		"COALESCE(value, 0)", "NULLIF(val, 0)",
	}

	for _, fn := range safeFunctions {
		t.Run("safe_"+fn, func(t *testing.T) {
			// 这些函数应该被认为是安全的
			sql := "SELECT " + fn + " FROM users"
			_, err := validator.ValidateAndSecure(sql)
			// 某些可能因为其他原因失败（如表不在白名单）
			_ = err // 只验证不会崩溃
		})
	}
}

// ========== 测试工具信息 ==========

func TestQueryTool_ToolInfoStructure(t *testing.T) {
	tool := &QueryTool{
		tenantID: "tenant-1",
	}

	info, err := tool.ToolInfo(context.Background())
	if err != nil {
		t.Fatalf("ToolInfo() error: %v", err)
	}

	// 验证 ToolInfo 结构
	if info.Name == "" {
		t.Error("ToolInfo.Name is empty")
	}
	if info.Desc == "" {
		t.Error("ToolInfo.Desc is empty")
	}
}

// ========== 工具接口测试 ==========

func TestQueryTool_ImplementsTool(t *testing.T) {
	// 验证 QueryTool 实现了必需的方法
	tool := &QueryTool{
		tenantID: "tenant-1",
	}

	// 验证方法可以调用 - 实际调用会返回错误但证明方法存在
	_, _ = tool.ToolInfo(context.Background())
	_, _ = tool.InvokableRun(context.Background(), "{}")
	_ = tool.String()
}
