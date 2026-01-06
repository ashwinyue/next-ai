// Package agent 提供 Agent 工具中间件
// 基于 Eino 官方示例，提供错误处理和 JSON 修复中间件
package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
	"github.com/kaptinlin/jsonrepair"
)

// ========== ErrorRemover 中间件 ==========

// ErrorRemoverConfig 错误移除中间件配置
type ErrorRemoverConfig struct {
	// ErrorHandler 自定义错误处理函数
	// 参数:
	//   - ctx: 上下文
	//   - toolInput: 工具调用输入
	//   - err: 原始错误
	// 返回:
	//   - string: 替换错误的结果字符串
	ErrorHandler func(ctx context.Context, toolInput *compose.ToolInput, err error) string
}

// DefaultErrorHandler 默认错误处理器
func DefaultErrorHandler() func(ctx context.Context, toolInput *compose.ToolInput, err error) string {
	return func(ctx context.Context, in *compose.ToolInput, err error) string {
		return fmt.Sprintf("工具 '%s' 调用失败: %s", in.Name, err.Error())
	}
}

// NewErrorRemoverMiddleware 创建错误移除中间件
// 捕获工具调用错误并返回自定义消息，而不是让整个 Agent 失败
//
// 适用场景：
// - 某些工具失败不应中断 Agent 执行
// - 需要向 LLM 返回友好的错误信息
//
// 使用示例:
//
//	middleware := agent.NewErrorRemoverMiddleware(nil)
//	toolsConfig := compose.ToolsNodeConfig{
//	    Tools: []tool.BaseTool{...},
//	    ToolCallMiddlewares: []compose.ToolMiddleware{middleware},
//	}
func NewErrorRemoverMiddleware(cfg *ErrorRemoverConfig) compose.ToolMiddleware {
	handler := DefaultErrorHandler()
	if cfg != nil && cfg.ErrorHandler != nil {
		handler = cfg.ErrorHandler
	}

	return compose.ToolMiddleware{
		Invokable: func(next compose.InvokableToolEndpoint) compose.InvokableToolEndpoint {
			return func(ctx context.Context, in *compose.ToolInput) (*compose.ToolOutput, error) {
				output, err := next(ctx, in)
				if err != nil {
					// 不中断重试错误
					if _, ok := compose.IsInterruptRerunError(err); ok {
						return nil, err
					}
					// 将错误转换为成功结果
					result := handler(ctx, in, err)
					return &compose.ToolOutput{Result: result}, nil
				}
				return output, nil
			}
		},
		Streamable: func(next compose.StreamableToolEndpoint) compose.StreamableToolEndpoint {
			return func(ctx context.Context, in *compose.ToolInput) (*compose.StreamToolOutput, error) {
				streamOutput, err := next(ctx, in)
				if err != nil {
					// 不中断重试错误
					if _, ok := compose.IsInterruptRerunError(err); ok {
						return nil, err
					}
					// 将错误转换为成功流
					result := handler(ctx, in, err)
					return &compose.StreamToolOutput{
						Result: schema.StreamReaderFromArray([]string{result}),
					}, nil
				}
				return streamOutput, nil
			}
		},
	}
}

// ========== JsonFix 中间件 ==========

// NewJsonFixMiddleware 创建 JSON 修复中间件
// 修复 LLM 生成的格式错误的 JSON 参数
//
// 适用场景：
// - LLM 生成的 JSON 参数格式不规范
// - 参数中有多余的文字、注释或格式错误
//
// 使用示例:
//
//	middleware := agent.NewJsonFixMiddleware()
//	toolsConfig := compose.ToolsNodeConfig{
//	    Tools: []tool.BaseTool{...},
//	    ToolCallMiddlewares: []compose.ToolMiddleware{middleware},
//	}
func NewJsonFixMiddleware() compose.ToolMiddleware {
	return compose.ToolMiddleware{
		Invokable: func(next compose.InvokableToolEndpoint) compose.InvokableToolEndpoint {
			return func(ctx context.Context, in *compose.ToolInput) (*compose.ToolOutput, error) {
				in.Arguments = repairJSON(in.Arguments)
				return next(ctx, in)
			}
		},
		Streamable: func(next compose.StreamableToolEndpoint) compose.StreamableToolEndpoint {
			return func(ctx context.Context, in *compose.ToolInput) (*compose.StreamToolOutput, error) {
				in.Arguments = repairJSON(in.Arguments)
				return next(ctx, in)
			}
		},
	}
}

// repairJSON 修复 JSON 字符串
// 策略：先尝试快速路径（有效 JSON 直接返回），再尝试修复
func repairJSON(input string) string {
	s := strings.TrimSpace(input)

	// 快速路径：已经是有效的 JSON 对象
	if strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}") && json.Valid([]byte(s)) {
		return s
	}

	// 尝试提取 JSON 对象区域
	i := strings.IndexByte(s, '{')
	j := strings.LastIndexByte(s, '}')
	if i >= 0 && j >= i {
		sub := s[i : j+1]
		if json.Valid([]byte(sub)) {
			return sub
		}
		s = sub
	}

	// 移除常见的 LLM 生成伪影
	s = strings.TrimPrefix(s, "<|FunctionCallBegin|>")
	s = strings.TrimSuffix(s, "<|FunctionCallEnd|>")
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")

	// 检查是否有效
	if json.Valid([]byte(s)) {
		return s
	}

	// 启发式：补全缺失的大括号
	if !strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}") {
		s = "{" + s
	} else if strings.HasPrefix(s, "{") && !strings.HasSuffix(s, "}") {
		s = s + "}"
	}

	// 使用 jsonrepair 进行强力修复
	out, err := jsonrepair.JSONRepair(s)
	if err != nil {
		return s // 修复失败，返回原值
	}
	return out
}

// ========== 组合中间件 ==========

// DefaultMiddlewares 返回默认的中间件组合
// 包含 JSON 修复和错误处理
func DefaultMiddlewares() []compose.ToolMiddleware {
	return []compose.ToolMiddleware{
		NewJsonFixMiddleware(),      // 先修复 JSON
		NewErrorRemoverMiddleware(nil), // 再处理错误
	}
}

// CustomMiddlewares 返回自定义中间件组合
func CustomMiddlewares(errorHandler func(ctx context.Context, toolInput *compose.ToolInput, err error) string) []compose.ToolMiddleware {
	return []compose.ToolMiddleware{
		NewJsonFixMiddleware(),
		NewErrorRemoverMiddleware(&ErrorRemoverConfig{
			ErrorHandler: errorHandler,
		}),
	}
}
