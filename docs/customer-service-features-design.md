# 售后客服 Agent 进阶功能设计文档

> Version: 1.0
> Date: 2025-01-09
> Author: Next-AI Team

---

## 目录

- [1. 智能反问](#1-智能反问)
- [2. 多轮确认](#2-多轮确认)
- [3. 情绪识别](#3-情绪识别)
- [4. 人工接管](#4-人工接管)
- [5. 工单系统](#5-工单系统)
- [6. 数据统计](#6-数据统计)

---

## 1. 智能反问

### 1.1 功能说明

当用户请求缺少必要参数时，Agent 自动识别缺失参数并主动追问，而不是直接报错。

### 1.2 场景示例

```
❌ 不智能：
用户：我要退款
Agent：缺少订单号参数，错误

✅ 智能反问：
用户：我要退款
Agent：好的，请问您要退款哪个订单？可以提供订单号或手机号
```

### 1.3 技术方案

#### 方案一：Prompt 工程（推荐）

```go
// 系统提示词中明确要求
const SystemPrompt = `
你是一个售后客服助手。当用户请求缺少必要参数时：

1. 识别缺少哪些参数
2. 用友好的语言向用户询问
3. 一次只问一个问题
4. 确认后再执行操作

示例：
用户：我要退款
你：请问您要退款哪个订单？请提供订单号

用户：ORD123
你：请问退款原因是什么？（如：不想要了、商品质量问题等）
`
```

#### 方案二：工具参数校验

```go
// internal/service/validator/validator.go
package validator

type ParamRequirement struct {
    Name     string
    Required bool
    Prompt   string  // 缺少时的追问话术
}

type ToolValidator struct {
    params []ParamRequirement
}

func (v *ToolValidator) Validate(input map[string]any) (*FollowUp, error) {
    missing := []string{}
    for _, p := range v.params {
        if p.Required && input[p.Name] == nil {
            missing = append(missing, p.Prompt)
        }
    }

    if len(missing) > 0 {
        return &FollowUp{
            NeedInput:  true,
            Questions:  missing,
            ToolCalled: "", // 不调用工具
        }, nil
    }

    return &FollowUp{NeedInput: false}, nil
}

// 使用示例
refundValidator := &ToolValidator{
    params: []ParamRequirement{
        {Name: "order_no", Required: true, Prompt: "请问订单号是多少？"},
        {Name: "reason", Required: true, Prompt: "请问退款原因是什么？"},
    },
}
```

#### 方案三：Eino 内置能力（探索）

Eino 部分组件支持参数校验，可研究是否有原生支持。

### 1.4 数据模型

```go
// internal/model/follow_up.go
package model

// FollowUp 追问记录
type FollowUp struct {
    ID        uint   `gorm:"primaryKey"`
    SessionID string `gorm:"index"` // 关联会话
    ToolName  string                // 原本想调用的工具
    Missing   string                // JSON: 缺失的参数列表
    Question  string                // 向用户追问的问题
    Answer    string                // 用户回答
    Resolved  bool                  // 是否已解决
    CreatedAt time.Time
}
```

### 1.5 接口设计

```go
// internal/service/agent/agent.go
type AgentResponse struct {
    Type     ResponseType `json:"type"`     // answer / follow_up / tool_call
    Content  string       `json:"content"`  // 回复内容
    ToolCall *ToolCall    `json:"tool_call"` // 工具调用信息
}

type ResponseType string

const (
    ResponseTypeAnswer    ResponseType = "answer"     // 直接回答
    ResponseTypeFollowUp  ResponseType = "follow_up"  // 需要追问
    ResponseTypeToolCall  ResponseType = "tool_call"  // 调用工具中
)
```

---

## 2. 多轮确认

### 2.1 功能说明

执行高风险操作（退款、删除、大额转账）前，向用户展示操作详情并要求二次确认。

### 2.2 场景示例

```
用户：我要退款订单 ORD123

Agent：
请确认以下退款信息：
  订单号：ORD123
  商品：iPhone 15 Pro 256G
  金额：¥7999.00
  退款方式：原路退回

回复"确认"继续，或取消操作。

用户：确认

Agent：退款申请已提交，3-5个工作日到账
```

### 2.3 技术方案

```go
// internal/service/confirmation/confirmation.go
package confirmation

// Confirmation 确认状态
type Confirmation struct {
    SessionID    string
    Action       string            // "refund", "cancel_order"
    Summary      string            // 操作摘要（展示给用户）
    ToolCall     *ToolCallPlan     // 待执行的工具调用
    ExpiresAt    time.Time         // 确认过期时间（5分钟）
}

type ConfirmationManager struct {
    store  Store
    logger *zap.Logger
}

// NeedConfirm 判断是否需要确认
func (m *ConfirmationManager) NeedConfirm(toolName string, params map[string]any) bool {
    highRiskTools := map[string]bool{
        "refund_order":  true,
        "cancel_order":  true,
        "delete_address": true,
    }
    return highRiskTools[toolName]
}

// GenerateConfirmation 生成确认信息
func (m *ConfirmationManager) GenerateConfirmation(toolName string, params map[string]any) *Confirmation {
    switch toolName {
    case "refund_order":
        return &Confirmation{
            Action: "退款",
            Summary: m.formatRefundSummary(params),
            ToolCall: &ToolCallPlan{
                Name:   toolName,
                Params: params,
            },
            ExpiresAt: time.Now().Add(5 * time.Minute),
        }
    // ... 其他 case
    }
}

func (m *ConfirmationManager) formatRefundSummary(params map[string]any) string {
    return fmt.Sprintf(`
订单号：%s
金额：¥%.2f
退款方式：原路退回
`, params["order_no"], params["amount"])
}
```

### 2.4 状态机设计

```
                    ┌─────────────┐
                    │   用户请求   │
                    └──────┬──────┘
                           │
                           ▼
                   ┌───────────────┐
                   │ 需要确认？     │
                   └───────┬───────┘
                           │
              ┌────────────┴────────────┐
              │ 是                       │ 否
              ▼                          ▼
     ┌──────────────┐           ┌──────────────┐
     │ 发送确认信息   │           │ 直接执行工具  │
     └──────┬───────┘           └──────────────┘
            │
            ▼
     ┌──────────────┐
     │ 等待用户确认  │ ◄─── 超时取消
     └──────┬───────┘
            │
      ┌─────┴─────┐
      │           │
   确认         取消
      │           │
      ▼           ▼
 ┌─────────┐  ┌─────────┐
 │ 执行操作 │  │ 取消操作 │
 └─────────┘  └─────────┘
```

### 2.5 数据模型

```go
// internal/model/pending_action.go
package model

// PendingAction 待确认的操作
type PendingAction struct {
    ID        uint            `gorm:"primaryKey"`
    SessionID string          `gorm:"index"`
    ToolName  string
    Params    string          `gorm:"type:json"` // 工具参数 JSON
    Summary   string          // 确认摘要
    Status    string          // pending / confirmed / cancelled / expired
    ExpiresAt time.Time       // 过期时间
    CreatedAt time.Time
}
```

---

## 3. 情绪识别

### 3.1 功能说明

分析用户消息中的情绪，当检测到愤怒、极度不满时，自动转人工客服。

### 3.2 场景示例

```
用户：你们这是什么垃圾服务！我都等了一个月了！

系统检测到愤怒情绪 → 自动转人工
→ Agent：非常抱歉给您带来不便，已为您转接人工客服，请稍候...
```

### 3.3 技术方案

#### 方案一：关键词规则（简单快速）

```go
// internal/service/sentiment/rule.go
package sentiment

type RuleEngine struct {
    angryKeywords []string
}

func NewRuleEngine() *RuleEngine {
    return &RuleEngine{
        angryKeywords: []string{
            "垃圾", "骗子", "投诉", "退款", "sb", "傻逼",
            "什么破", "什么鬼", "气的我", "无语", "太差了",
            "再也不买", "曝光你们", "12315", "工商",
        },
    }
}

func (e *RuleEngine) Detect(message string) *Sentiment {
    msg := strings.ToLower(message)

    // 愤怒检测
    for _, kw := range e.angryKeywords {
        if strings.Contains(msg, kw) {
            return &Sentiment{
                Type:     SentimentAngry,
                Score:    0.9,
                Strategy: StrategyTransferToHuman, // 转人工
                Reason:   "检测到愤怒情绪",
            }
        }
    }

    return &Sentiment{
        Type:     SentimentNeutral,
        Score:    0.5,
        Strategy: StrategyContinue, // 继续
    }
}
```

#### 方案二：LLM 情绪分析（更准确）

```go
// internal/service/sentiment/llm.go
package sentiment

type LLMDetector struct {
    chatModel *ChatModel
}

func (d *LLMDetector) Detect(ctx context.Context, message string) (*Sentiment, error) {
    prompt := fmt.Sprintf(`
分析以下用户消息的情绪，只返回 JSON：
{
  "type": "neutral/angry/sad/happy",
  "score": 0.0-1.0,
  "need_human": boolean
}

用户消息：%s
`, message)

    response, err := d.chatModel.Generate(ctx, prompt)
    if err != nil {
        return nil, err
    }

    var result SentimentAnalysis
    json.Unmarshal([]byte(response), &result)

    strategy := StrategyContinue
    if result.NeedHuman || result.Type == "angry" {
        strategy = StrategyTransferToHuman
    }

    return &Sentiment{
        Type:     SentimentType(result.Type),
        Score:    result.Score,
        Strategy: strategy,
    }, nil
}
```

### 3.4 数据模型

```go
// internal/model/sentiment_record.go
package model

// SentimentRecord 情绪记录
type SentimentRecord struct {
    ID        uint           `gorm:"primaryKey"`
    SessionID string         `gorm:"index"`
    Message   string         // 用户消息
    Type      string         // neutral / angry / sad / happy
    Score     float64        // 情绪强度 0-1
    Action    string         // continue / transfer_human
    CreatedAt time.Time
}
```

### 3.5 集成流程

```go
// internal/service/agent/agent.go
func (a *Agent) ProcessMessage(ctx context.Context, sessionID, message string) (*Response, error) {
    // 1. 情绪检测
    sentiment := a.sentimentDetector.Detect(message)

    // 2. 记录情绪（用于后续分析）
    a.sentimentRepo.Record(ctx, &SentimentRecord{
        SessionID: sessionID,
        Message:   message,
        Type:      sentiment.Type,
        Score:     sentiment.Score,
    })

    // 3. 根据情绪决定策略
    if sentiment.Strategy == StrategyTransferToHuman {
        return a.transferToHuman(ctx, sessionID, "检测到用户情绪激动，转人工处理")
    }

    // 4. 正常处理
    return a.chat(ctx, sessionID, message)
}
```

---

## 4. 人工接管

### 4.1 功能说明

当 Agent 无法处理或处理失败时，平滑切换到人工客服。

### 4.2 触发条件

```
1. 情绪检测：用户愤怒
2. 处理失败：工具调用失败 N 次
3. 用户主动：用户说"转人工"、"找客服"
4. 超出能力：Agent 识别无法处理
5. 等待超时：用户长时间未回复
```

### 4.3 技术方案

```go
// internal/service/human_transfer/transfer.go
package human_transfer

type TransferManager struct {
    agentQueue    Queue    // 客服队列
    sessionStore  Store    // 会话存储
    notification  Notifier // 通知系统
}

// TransferToHuman 转人工
func (m *TransferManager) TransferToHuman(ctx context.Context, req *TransferRequest) error {
    // 1. 获取会话历史
    history, err := m.sessionStore.GetHistory(ctx, req.SessionID)
    if err != nil {
        return err
    }

    // 2. 创建工单
    ticket := &Ticket{
        SessionID: req.SessionID,
        UserID:    req.UserID,
        History:   history,
        Reason:    req.Reason,
        Priority:  m.calculatePriority(history),
        Status:    TicketStatusPending,
    }
    m.ticketRepo.Create(ctx, ticket)

    // 3. 通知客服
    m.notification.NotifyHumanAgents(ctx, &Notification{
        Type:    "new_ticket",
        TicketID: ticket.ID,
        Priority: ticket.Priority,
        Summary:  m.formatSummary(ticket),
    })

    // 4. 更新会话状态
    m.sessionStore.UpdateStatus(ctx, req.SessionID, SessionStatusHumanServing)

    return nil
}

// calculatePriority 根据情况计算优先级
func (m *TransferManager) calculatePriority(history []*Message) TicketPriority {
    // 愤怒情绪 → 高优先级
    // 等待时间长 → 高优先级
    // 退款金额大 → 高优先级
    // ...
    return TicketPriorityNormal
}
```

### 4.4 客服工作台接口

```go
// internal/handler/agent_console.go

// GetPendingTickets 获取待处理工单列表
func (h *ConsoleHandler) GetPendingTickets(c *gin.Context) {
    tickets := h.ticketService.GetPending(c)
    c.JSON(200, tickets)
}

// AcceptTicket 接单
func (h *ConsoleHandler) AcceptTicket(c *gin.Context) {
    ticketID := c.Param("id")
    err := h.ticketService.Assign(c, ticketID, h.currentAgentID)
    // ...
}

// SendMessage 客服发送消息
func (h *ConsoleHandler) SendMessage(c *gin.Context) {
    // 客服消息直接发送给用户
    // 同时记录到会话历史
}
```

### 4.5 数据模型

```go
// internal/model/ticket.go
package model

// Ticket 工单
type Ticket struct {
    ID           uint           `gorm:"primaryKey"`
    TicketNo     string         `gorm:"uniqueIndex"` // 工单号
    SessionID    string         `gorm:"index"`
    UserID       string         `gorm:"index"`
    AssignedTo   string         // 分配给的客服 ID
    Reason       string         // 转人工原因
    Priority     string         // low / normal / high / urgent
    Status       string         // pending / assigned / resolved / closed
    History      string         `gorm:"type:json"` // 对话历史
    Resolution   string         // 解决方案
    CreatedAt    time.Time
    AssignedAt   *time.Time
    ResolvedAt   *time.Time
}
```

---

## 5. 工单系统

### 5.1 功能说明

将用户问题记录为工单，支持分类、分配、跟踪、关闭的全流程管理。

### 5.2 工单类型

```
1. 问题反馈：产品问题、使用疑问
2. 售后服务：退款、换货、维修
3. 投诉建议：服务投诉、功能建议
4. 技术支持：技术故障、API 问题
```

### 5.3 工单生命周期

```
创建 → 待分配 → 已分配 → 处理中 → 已解决 → 已关闭
  ↓                           ↓
超时自动关闭              用户不满意 → 重新打开
```

### 5.4 技术方案

```go
// internal/service/ticket/ticket.go
package ticket

type TicketService struct {
    repo       Repository
    dispatcher Dispatcher    // 分配器
    notifier   Notifier      // 通知
}

// CreateTicket 创建工单
func (s *TicketService) CreateTicket(ctx context.Context, req *CreateTicketRequest) (*Ticket, error) {
    // 1. 生成工单号
    ticketNo := s.generateTicketNo()

    // 2. 自动分类（可选 LLM）
    category := s.classifyTicket(req)

    // 3. 创建工单
    ticket := &Ticket{
        TicketNo:  ticketNo,
        UserID:    req.UserID,
        Title:     req.Title,
        Content:   req.Content,
        Category:  category,
        Priority:  s.calculatePriority(req),
        Status:    TicketStatusPending,
        Source:    req.Source, // chat / email / phone
    }
    err := s.repo.Create(ctx, ticket)
    if err != nil {
        return nil, err
    }

    // 4. 自动分配
    agentID, err := s.dispatcher.FindAvailableAgent(ctx, category)
    if err == nil {
        s.Assign(ctx, ticket.ID, agentID)
    }

    // 5. 通知用户
    s.notifier.NotifyUser(ctx, req.UserID, &Notification{
        Type:     "ticket_created",
        TicketNo: ticketNo,
    })

    return ticket, nil
}

// classifyTicket 工单分类
func (s *TicketService) classifyTicket(req *CreateTicketRequest) string {
    // 方案一：规则
    if strings.Contains(req.Content, "退款") || strings.Contains(req.Content, "退货") {
        return CategoryRefund
    }

    // 方案二：LLM 分类
    prompt := fmt.Sprintf("分类以下工单类型：%s", req.Content)
    // ...
}
```

### 5.5 数据模型

```go
// internal/model/ticket.go
package model

// Ticket 工单主表
type Ticket struct {
    ID          uint           `gorm:"primaryKey"`
    TicketNo    string         `gorm:"uniqueIndex"`
    UserID      string         `gorm:"index"`
    AssignedTo  string         `gorm:"index"` // 处理人 ID
    Title       string         // 工单标题
    Content     string         `gorm:"type:text"` // 详细描述
    Category    string         // 分类
    Priority    string         // 优先级
    Status      string         // 状态
    Source      string         // 来源
    Tags        string         `gorm:"type:json"` // 标签
    CreatedAt   time.Time
    UpdatedAt   time.Time
    ResolvedAt  *time.Time
    ClosedAt    *time.Time
}

// TicketComment 工单评论
type TicketComment struct {
    ID        uint      `gorm:"primaryKey"`
    TicketID  uint      `gorm:"index"`
    UserID    string    // 评论人
    Content   string    `gorm:"type:text"`
    IsInternal bool     // 是否内部评论（用户不可见）
    CreatedAt time.Time
}

// TicketAttachment 工单附件
type TicketAttachment struct {
    ID        uint      `gorm:"primaryKey"`
    TicketID  uint      `gorm:"index"`
    FileName  string
    FileURL   string
    FileSize  int64
    CreatedAt time.Time
}
```

### 5.6 接口设计

```go
// internal/handler/ticket.go

// 创建工单
POST /api/tickets
{
  "title": "订单退款",
  "content": "订单 ORD123 申请退款...",
  "category": "refund",
  "priority": "high"
}

// 查询工单列表
GET /api/tickets?status=assigned&priority=high

// 工单详情
GET /api/tickets/:id

// 添加评论
POST /api/tickets/:id/comments
{
  "content": "已处理，退款中",
  "is_internal": false
}

// 更新状态
PUT /api/tickets/:id/status
{
  "status": "resolved",
  "resolution": "退款已提交"
}

// 关闭工单
POST /api/tickets/:id/close
```

---

## 6. 数据统计

### 6.1 功能说明

统计客服系统运行数据，生成报表，辅助决策。

### 6.2 统计维度

```
1. 会话统计
   - 日均会话量
   - 平均会话时长
   - 解决率

2. 消息统计
   - 消息总量
   - Agent 处理比例
   - 人工接管比例

3. 工单统计
   - 工单数量趋势
   - 分类分布
   - 平均处理时长
   - 超时率

4. 满意度统计
   - 用户评分
   - NPS（净推荐值）
   - 投诉率

5. 情绪统计
   - 情绪分布
   - 愤怒率趋势
   - 情绪与问题类型关联
```

### 6.3 技术方案

```go
// internal/service/analytics/analytics.go
package analytics

type AnalyticsService struct {
    db       *gorm.DB
    cache    Cache
    logger   *zap.Logger
}

// SessionMetrics 会话指标
type SessionMetrics struct {
    Date              string
    TotalSessions     int64
    AvgDuration       float64 // 平均时长（分钟）
    ResolvedCount     int64
    ResolvedRate      float64
    AgentResolved     int64
    HumanResolved     int64
}

// GetSessionMetrics 获取会话统计
func (s *AnalyticsService) GetSessionMetrics(ctx context.Context, start, end time.Time) (*SessionMetrics, error) {
    cacheKey := fmt.Sprintf("metrics:session:%s:%s", start.Format("2006-01-02"), end.Format("2006-01-02"))

    // 尝试从缓存获取
    if cached := s.cache.Get(ctx, cacheKey); cached != nil {
        return cached.(*SessionMetrics), nil
    }

    // 查询数据库
    var metrics SessionMetrics
    err := s.db.WithContext(ctx).
        Model(&ChatSession{}).
        Where("created_at BETWEEN ? AND ?", start, end).
        Select(`
            DATE(created_at) as date,
            COUNT(*) as total_sessions,
            AVG(TIMEDIFF(updated_at, created_at)) as avg_duration,
            SUM(CASE WHEN status = 'resolved' THEN 1 ELSE 0 END) as resolved_count
        `).
        Scan(&metrics).Error

    if err != nil {
        return nil, err
    }

    metrics.ResolvedRate = float64(metrics.ResolvedCount) / float64(metrics.TotalSessions)

    // 缓存结果
    s.cache.Set(ctx, cacheKey, &metrics, 5*time.Minute)

    return &metrics, nil
}

// TicketMetrics 工单指标
type TicketMetrics struct {
    Date           string
    TotalTickets   int64
    ByCategory     map[string]int64
    ByPriority     map[string]int64
    AvgProcessTime float64 // 平均处理时长（小时）
    OverdueCount   int64
    OverdueRate    float64
}

// GetCategoryTrend 分类趋势（用于发现问题）
func (s *AnalyticsService) GetCategoryTrend(ctx context.Context, days int) ([]*CategoryTrend, error) {
    var trends []*CategoryTrend

    err := s.db.WithContext(ctx).
        Model(&Ticket{}).
        Where("created_at > ?", Now().AddDate(0, 0, -days)).
        Select(`
            category,
            DATE(created_at) as date,
            COUNT(*) as count
        `).
        Group("category, DATE(created_at)").
        Order("date ASC, count DESC").
        Scan(&trends).Error

    return trends, err
}

// SentimentMetrics 情绪指标
type SentimentMetrics struct {
    Date         string
    TotalMessages int64
    AngryCount   int64
    AngryRate    float64
    ByType       map[string]int64
}
```

### 6.4 报表展示

```go
// internal/handler/analytics.go

// DashboardData 仪表盘数据
type DashboardData struct {
    // 今日数据
    Today struct {
        Sessions   int64   `json:"sessions"`
        Messages   int64   `json:"messages"`
        Tickets    int64   `json:"tickets"`
        ResolvedRate float64 `json:"resolved_rate"`
    } `json:"today"`

    // 趋势（最近7天）
    Trend struct {
        Sessions  []Point `json:"sessions"`
        Tickets   []Point `json:"tickets"`
        Satisfied []Point `json:"satisfied"`
    } `json:"trend"`

    // 问题分类
    TopCategories []CategoryCount `json:"top_categories"`

    // 人工接管原因
    TransferReasons []ReasonCount `json:"transfer_reasons"`
}

type Point struct {
    Date  string `json:"date"`
    Value int64  `json:"value"`
}

func (h *AnalyticsHandler) Dashboard(c *gin.Context) {
    data := &DashboardData{}

    // 并行查询
    var wg sync.WaitGroup
    wg.Add(4)

    go func() {
        defer wg.Done()
        data.Today.Sessions = h.analytics.TodaySessions(c)
    }()
    go func() {
        defer wg.Done()
        data.Trend.Sessions = h.analytics.SessionTrend(c, 7)
    }()
    go func() {
        defer wg.Done()
        data.TopCategories = h.analytics.TopCategories(c, 10)
    }()
    go func() {
        defer wg.Done()
        data.TransferReasons = h.analytics.TransferReasons(c, 7)
    }()

    wg.Wait()
    c.JSON(200, data)
}
```

### 6.5 数据模型

```go
// internal/model/metrics.go
package model

// DailyMetrics 每日指标（预计算表）
type DailyMetrics struct {
    Date          string  `gorm:"primaryKey"`
    Sessions      int64
    Messages      int64
    AgentResolved int64
    HumanResolved int64
    Tickets       int64
    ResolvedRate  float64
    AngryRate     float64
    AvgDuration   float64
    CreatedAt     time.Time
}

// MetricSnapshot 实时指标快照（用于 Redis）
type MetricSnapshot struct {
    Key       string
    Value     float64
    Timestamp time.Time
}
```

---

## 附录：整体架构图

```
┌─────────────────────────────────────────────────────────────────┐
│                           用户层                                 │
│  微信小程序 │ H5 网页 │ App │ 电话 │ 邮件                          │
└────────────────────────────┬────────────────────────────────────┘
                             │
┌────────────────────────────┴────────────────────────────────────┐
│                         API 网关层                               │
│  认证 │ 限流 │ 日志 │ 路由                                        │
└────────────────────────────┬────────────────────────────────────┘
                             │
┌────────────────────────────┴────────────────────────────────────┐
│                      Agent 服务层                                │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐        │
│  │意图识别   │  │工具调度   │  │情绪检测   │  │人工接管   │        │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘        │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐        │
│  │智能反问   │  │多轮确认   │  │工单系统   │  │数据统计   │        │
│  └──────────┘  └──────────┘  └──────────┘  └──────────┘        │
└────────────────────────────┬────────────────────────────────────┘
                             │
┌────────────────────────────┴────────────────────────────────────┐
│                      业务系统集成层                              │
│  订单系统 │ 退款系统 │ 物流系统 │ CRM │ 库存系统                   │
└────────────────────────────┬────────────────────────────────────┘
                             │
┌────────────────────────────┴────────────────────────────────────┐
│                      数据存储层                                  │
│  PostgreSQL │ Redis │ Elasticsearch │ ClickHouse（可选，用于分析） │
└─────────────────────────────────────────────────────────────────┘
```

---

## 实现优先级建议

```
P0 (必须有)：
  ✅ 智能反问（通过 Prompt 实现）
  ✅ 人工接管（基本流程）
  ✅ 工单系统（基本 CRUD）

P1 (重要)：
  ⭐ 情绪识别（关键词规则版）
  ⭐ 多轮确认（高风险操作）
  ⭐ 数据统计（基础指标）

P2 (增强)：
  🔸 情绪识别（LLM 版本）
  🔸 高级报表
  🔸 客服工作台
```
