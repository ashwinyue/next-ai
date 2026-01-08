// Package database 提供数据库查询工具，允许 AI 安全地执行 SQL 查询
package database

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	pg_query "github.com/pganalyze/pg_query_go/v6"
	"gorm.io/gorm"
)

const (
	ToolDatabaseQuery = "database_query"
)

// DatabaseQueryInput 数据库查询输入参数
type DatabaseQueryInput struct {
	SQL string `json:"sql" jsonschema:"The SELECT SQL query to execute. Only SELECT queries are allowed."`
}

// SQLSecurityValidator SQL 安全验证器，使用 PostgreSQL 官方解析器
type SQLSecurityValidator struct {
	allowedTables    map[string]bool
	allowedFunctions map[string]bool
	// TenantID 租户 ID，用于自动注入租户过滤条件
	TenantID string
}

// NewSQLSecurityValidator 创建 SQL 安全验证器
func NewSQLSecurityValidator(tenantID string) *SQLSecurityValidator {
	return &SQLSecurityValidator{
		allowedTables: map[string]bool{
			// 用户表
			"users": true,
			// 知识库表
			"knowledge_bases": true,
			"knowledges":      true,
			"chunks":          true,
			"chunk_tags":      true,
			// 聊天表
			"chat_sessions": true,
			"chat_messages": true,
			// Agent 表
			"agents": true,
			// 工具表
			"tools": true,
			// FAQ 表
			"faqs":        true,
			"faq_entries": true,
			// 模型表
			"models": true,
		},
		allowedFunctions: map[string]bool{
			// 聚合函数
			"count": true, "sum": true, "avg": true, "min": true, "max": true,
			"array_agg": true, "string_agg": true,
			"json_agg": true, "jsonb_agg": true,
			// 安全标量函数
			"coalesce": true, "nullif": true,
			"greatest": true, "least": true,
			"abs": true, "ceil": true, "floor": true, "round": true,
			"length": true, "lower": true, "upper": true,
			"trim": true, "ltrim": true, "rtrim": true,
			"substring": true, "concat": true, "concat_ws": true,
			"replace": true, "left": true, "right": true,
			"now": true, "current_date": true, "current_timestamp": true,
			"date_trunc": true, "extract": true,
			"to_char": true, "to_date": true, "to_timestamp": true,
		},
		TenantID: tenantID,
	}
}

// ValidateAndSecure 验证并加固 SQL 查询
func (v *SQLSecurityValidator) ValidateAndSecure(sqlQuery string) (string, error) {
	// 阶段 1: 基本输入验证
	if err := v.validateInput(sqlQuery); err != nil {
		return "", err
	}

	// 阶段 2: 使用 PostgreSQL 官方解析器解析 SQL
	result, err := pg_query.Parse(sqlQuery)
	if err != nil {
		return "", fmt.Errorf("SQL 解析错误: %v", err)
	}

	// 阶段 3: 确保只有一个语句
	if len(result.Stmts) == 0 {
		return "", fmt.Errorf("空查询")
	}
	if len(result.Stmts) > 1 {
		return "", fmt.Errorf("不允许执行多条语句")
	}

	stmt := result.Stmts[0].Stmt

	// 阶段 4: 确保是 SELECT 语句
	selectStmt := stmt.GetSelectStmt()
	if selectStmt == nil {
		return "", fmt.Errorf("只允许 SELECT 查询")
	}

	// 阶段 5: 递归验证 SELECT 语句
	tablesInQuery, err := v.validateSelectStmt(selectStmt)
	if err != nil {
		return "", err
	}

	// 阶段 6: 标准化 SQL
	normalizedSQL, err := pg_query.Deparse(result)
	if err != nil {
		return "", fmt.Errorf("SQL 标准化失败: %v", err)
	}

	// 阶段 7: 注入租户过滤条件 (如果设置了 TenantID)
	securedSQL := v.injectTenantConditions(normalizedSQL, tablesInQuery)

	return securedSQL, nil
}

// validateInput 基本输入验证
func (v *SQLSecurityValidator) validateInput(sql string) error {
	if strings.Contains(sql, "\x00") {
		return fmt.Errorf("SQL 查询包含非法字符")
	}
	if len(sql) < 6 {
		return fmt.Errorf("SQL 查询过短")
	}
	if len(sql) > 4096 {
		return fmt.Errorf("SQL 查询过长 (最大 4096 字符)")
	}
	return nil
}

// validateSelectStmt 验证 SELECT 语句并提取表信息
func (v *SQLSecurityValidator) validateSelectStmt(stmt *pg_query.SelectStmt) (map[string]string, error) {
	tablesInQuery := make(map[string]string)

	// 检查复合查询 (UNION/INTERSECT/EXCEPT)
	if stmt.Op != pg_query.SetOperation_SETOP_NONE {
		return nil, fmt.Errorf("不允许使用复合查询 (UNION/INTERSECT/EXCEPT)")
	}

	// 检查 WITH 子句 (CTE)
	if stmt.WithClause != nil {
		return nil, fmt.Errorf("不允许使用 WITH 子句 (CTE)")
	}

	// 检查 INTO 子句
	if stmt.IntoClause != nil {
		return nil, fmt.Errorf("不允许使用 SELECT INTO")
	}

	// 检查锁定子句
	if len(stmt.LockingClause) > 0 {
		return nil, fmt.Errorf("不允许使用锁定子句 (FOR UPDATE 等)")
	}

	// 验证 FROM 子句
	for _, fromItem := range stmt.FromClause {
		if err := v.validateFromItem(fromItem, tablesInQuery); err != nil {
			return nil, err
		}
	}

	// 验证 WHERE 子句
	if stmt.WhereClause != nil {
		if err := v.validateNode(stmt.WhereClause); err != nil {
			return nil, err
		}
	}

	// 确保至少有一个有效表
	if len(tablesInQuery) == 0 {
		return nil, fmt.Errorf("查询中未找到有效表")
	}

	return tablesInQuery, nil
}

// validateFromItem 验证 FROM 子句项
func (v *SQLSecurityValidator) validateFromItem(node *pg_query.Node, tables map[string]string) error {
	if node == nil {
		return nil
	}

	// 处理 RangeVar (简单表引用)
	if rv := node.GetRangeVar(); rv != nil {
		tableName := strings.ToLower(rv.Relname)

		// 检查 schema 限定
		if rv.Schemaname != "" {
			schemaName := strings.ToLower(rv.Schemaname)
			if schemaName != "public" {
				return fmt.Errorf("不允许访问 schema '%s'", rv.Schemaname)
			}
		}

		// 验证表名
		if !v.allowedTables[tableName] {
			return fmt.Errorf("表不允许访问: %s", rv.Relname)
		}

		// 获取别名
		alias := tableName
		if rv.Alias != nil && rv.Alias.Aliasname != "" {
			alias = strings.ToLower(rv.Alias.Aliasname)
		}
		tables[tableName] = alias
		return nil
	}

	// 处理 JoinExpr (JOIN)
	if je := node.GetJoinExpr(); je != nil {
		if err := v.validateFromItem(je.Larg, tables); err != nil {
			return err
		}
		if err := v.validateFromItem(je.Rarg, tables); err != nil {
			return err
		}
		if je.Quals != nil {
			if err := v.validateNode(je.Quals); err != nil {
				return err
			}
		}
		return nil
	}

	// 处理子查询 - 不允许
	if node.GetRangeSubselect() != nil {
		return fmt.Errorf("不允许在 FROM 子句中使用子查询")
	}

	return nil
}

// validateNode 递归验证 AST 节点
func (v *SQLSecurityValidator) validateNode(node *pg_query.Node) error {
	if node == nil {
		return nil
	}

	// 检查子查询
	if node.GetSubLink() != nil {
		return fmt.Errorf("不允许使用子查询")
	}

	// 检查函数调用
	if fc := node.GetFuncCall(); fc != nil {
		return v.validateFuncCall(fc)
	}

	// 递归检查表达式
	if ae := node.GetAExpr(); ae != nil {
		if err := v.validateNode(ae.Lexpr); err != nil {
			return err
		}
		if err := v.validateNode(ae.Rexpr); err != nil {
			return err
		}
	}

	// 检查布尔表达式
	if be := node.GetBoolExpr(); be != nil {
		for _, arg := range be.Args {
			if err := v.validateNode(arg); err != nil {
				return err
			}
		}
	}

	return nil
}

// validateFuncCall 验证函数调用
func (v *SQLSecurityValidator) validateFuncCall(fc *pg_query.FuncCall) error {
	// 获取函数名
	funcName := ""
	for _, namePart := range fc.Funcname {
		if s := namePart.GetString_(); s != nil {
			funcName = strings.ToLower(s.Sval)
		}
	}

	// 检查 schema 限定的函数调用
	if len(fc.Funcname) > 1 {
		schemaName := ""
		if s := fc.Funcname[0].GetString_(); s != nil {
			schemaName = strings.ToLower(s.Sval)
		}
		if schemaName != "" && schemaName != "pg_catalog" {
			return fmt.Errorf("不允许使用 schema 限定的函数调用: %s", schemaName)
		}
	}

	// 阻止危险函数前缀
	dangerousPrefixes := []string{"pg_", "lo_", "dblink", "file_", "copy_"}
	for _, prefix := range dangerousPrefixes {
		if strings.HasPrefix(funcName, prefix) {
			return fmt.Errorf("不允许使用函数 '%s' (危险前缀)", funcName)
		}
	}

	// 检查白名单
	if !v.allowedFunctions[funcName] {
		return fmt.Errorf("函数不允许: %s", funcName)
	}

	// 递归验证参数
	for _, arg := range fc.Args {
		if err := v.validateNode(arg); err != nil {
			return err
		}
	}

	return nil
}

// injectTenantConditions 注入租户过滤条件
func (v *SQLSecurityValidator) injectTenantConditions(sql string, tablesInQuery map[string]string) string {
	if v.TenantID == "" {
		return sql
	}

	// 需要 tenant_id 过滤的表
	tablesWithTenantID := map[string]bool{
		"users":           true,
		"knowledge_bases": true,
		"knowledges":      true,
		"chunks":          true,
		"chat_sessions":   true,
		"chat_messages":   true,
		"agents":          true,
		"tools":           true,
		"faqs":            true,
		"faq_entries":     true,
		"models":          true,
	}

	var conditions []string
	for tableName, alias := range tablesInQuery {
		if tablesWithTenantID[tableName] {
			conditions = append(conditions, fmt.Sprintf("%s.tenant_id = '%s'", alias, v.TenantID))
		}
	}

	if len(conditions) == 0 {
		return sql
	}

	tenantFilter := strings.Join(conditions, " AND ")

	// 检查 WHERE 是否存在
	wherePattern := regexp.MustCompile(`(?i)\bWHERE\b`)
	if wherePattern.MatchString(sql) {
		return wherePattern.ReplaceAllString(sql, fmt.Sprintf("WHERE %s AND ", tenantFilter))
	}

	// 在 ORDER BY, GROUP BY, LIMIT 等子句前添加
	clausePattern := regexp.MustCompile(`(?i)\b(GROUP BY|ORDER BY|LIMIT|OFFSET|HAVING|FETCH)\b`)
	if loc := clausePattern.FindStringIndex(sql); loc != nil {
		return sql[:loc[0]] + fmt.Sprintf(" WHERE %s ", tenantFilter) + sql[loc[0]:]
	}

	return fmt.Sprintf("%s WHERE %s", sql, tenantFilter)
}

// QueryTool 数据库查询工具
type QueryTool struct {
	db       *gorm.DB
	tenantID string
}

// NewQueryTool 创建数据库查询工具
func NewQueryTool(db *gorm.DB, tenantID string) *QueryTool {
	return &QueryTool{
		db:       db,
		tenantID: tenantID,
	}
}

// ToolInfo 返回工具信息
func (t *QueryTool) ToolInfo(ctx context.Context) (*schema.ToolInfo, error) {
	return &schema.ToolInfo{
		Name: ToolDatabaseQuery,
		Desc: `执行 SQL 查询以从数据库中获取信息。

## 安全特性
- 只读查询: 只允许 SELECT 语句
- 表白名单: 只允许查询授权的表
- 自动租户过滤: 自动添加 tenant_id 过滤条件

## 可用表

### 用户表 (users)
- id: 用户 ID
- username: 用户名
- email: 邮箱
- role: 角色
- created_at, updated_at: 时间戳

### 知识库表 (knowledge_bases)
- id: 知识库 ID
- name: 名称
- description: 描述
- tenant_id: 租户 ID
- created_at, updated_at: 时间戳

### 文档表 (knowledges)
- id: 文档 ID
- knowledge_base_id: 所属知识库 ID
- title: 标题
- description: 描述
- parse_status: 处理状态
- file_name, file_type: 文件信息
- created_at, updated_at: 时间戳

### 分块表 (chunks)
- id: 分块 ID
- knowledge_base_id: 所属知识库 ID
- knowledge_id: 所属文档 ID
- content: 内容
- chunk_type: 类型 (text/image/table)
- created_at, updated_at: 时间戳

### 聊天会话表 (chat_sessions)
- id: 会话 ID
- title: 标题
- agent_id: 关联的 Agent ID
- created_at, updated_at: 时间戳

### 聊天消息表 (chat_messages)
- id: 消息 ID
- session_id: 会话 ID
- role: 角色 (user/assistant/system)
- content: 内容
- created_at: 时间戳

### Agent 表 (agents)
- id: Agent ID
- name: 名称
- description: 描述
- system_prompt: 系统提示词
- is_active: 是否激活
- created_at, updated_at: 时间戳

## 使用示例

查询知识库列表:
{"sql": "SELECT id, name, description FROM knowledge_bases ORDER BY created_at DESC LIMIT 10"}

统计文档数量:
{"sql": "SELECT parse_status, COUNT(*) as count FROM knowledges GROUP BY parse_status"}

查询最近会话:
{"sql": "SELECT id, title, created_at FROM chat_sessions ORDER BY created_at DESC LIMIT 5"}

## 注意事项
- 只允许 SELECT 查询
- 建议使用 LIMIT 子句限制结果数量
- 支持多表 JOIN`,
		ParamsOneOf: schema.NewParamsOneOfByParams(
			map[string]*schema.ParameterInfo{
				"sql": {
					Type:     schema.String,
					Desc:     "要执行的 SELECT SQL 查询",
					Required: true,
				},
			},
		),
	}, nil
}

// InvokableRun 执行工具
func (t *QueryTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	var input DatabaseQueryInput
	if err := json.Unmarshal([]byte(argumentsInJSON), &input); err != nil {
		return "", fmt.Errorf("参数解析失败: %v", err)
	}

	if input.SQL == "" {
		return "", fmt.Errorf("缺少 'sql' 参数")
	}

	// 验证并加固 SQL
	validator := NewSQLSecurityValidator(t.tenantID)
	securedSQL, err := validator.ValidateAndSecure(input.SQL)
	if err != nil {
		return "", fmt.Errorf("SQL 验证失败: %v", err)
	}

	// 执行查询
	rows, err := t.db.WithContext(ctx).Raw(securedSQL).Rows()
	if err != nil {
		return "", fmt.Errorf("查询执行失败: %v", err)
	}
	defer rows.Close()

	// 获取列名
	columns, err := rows.Columns()
	if err != nil {
		return "", fmt.Errorf("获取列名失败: %v", err)
	}

	// 处理结果
	results := make([]map[string]interface{}, 0)
	for rows.Next() {
		columnValues := make([]interface{}, len(columns))
		columnPointers := make([]interface{}, len(columns))
		for i := range columnValues {
			columnPointers[i] = &columnValues[i]
		}

		if err := rows.Scan(columnPointers...); err != nil {
			return "", fmt.Errorf("读取行数据失败: %v", err)
		}

		rowMap := make(map[string]interface{})
		for i, colName := range columns {
			val := columnValues[i]
			if b, ok := val.([]byte); ok {
				rowMap[colName] = string(b)
			} else {
				rowMap[colName] = val
			}
		}
		results = append(results, rowMap)
	}

	return t.formatQueryResults(columns, results, securedSQL), nil
}

// formatQueryResults 格式化查询结果
func (t *QueryTool) formatQueryResults(columns []string, results []map[string]interface{}, query string) string {
	output := "=== 查询结果 ===\n\n"
	output += fmt.Sprintf("执行的 SQL: %s\n\n", query)
	output += fmt.Sprintf("返回 %d 行数据\n\n", len(results))

	if len(results) == 0 {
		output += "未找到匹配的记录。\n"
		return output
	}

	output += "=== 数据详情 ===\n\n"

	// 格式化每一行
	for i, row := range results {
		output += fmt.Sprintf("--- 记录 #%d ---\n", i+1)
		for _, col := range columns {
			value := row[col]
			var formattedValue string
			if value == nil {
				formattedValue = "<NULL>"
			} else if jsonData, err := json.Marshal(value); err == nil {
				switch v := value.(type) {
				case string:
					formattedValue = v
				case []byte:
					formattedValue = string(v)
				default:
					formattedValue = string(jsonData)
				}
			} else {
				formattedValue = fmt.Sprintf("%v", value)
			}
			output += fmt.Sprintf("  %s: %s\n", col, formattedValue)
		}
		output += "\n"
	}

	if len(results) > 10 {
		output += fmt.Sprintf("注意: 显示了全部 %d 条记录。建议使用 LIMIT 子句限制结果数量。\n", len(results))
	}

	return output
}

// String 类型工具的字符串表示
func (t *QueryTool) String() string {
	return ToolDatabaseQuery
}
