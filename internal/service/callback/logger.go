// Package callback 提供 Eino Callback 日志支持
// 参考 eino-examples，实现日志回调处理器
package callback

import (
	"context"
	"log"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/schema"
)

// Logger 日志回调处理器
// 实现 callbacks.Handler 接口，用于记录 Eino 组件的执行事件
type Logger struct {
	EnableDebug bool // 是否启用调试模式
}

// NewLogger 创建日志回调处理器
func NewLogger(enableDebug bool) *Logger {
	return &Logger{EnableDebug: enableDebug}
}

// OnStart 组件执行开始时调用
func (l *Logger) OnStart(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
	if l.EnableDebug {
		log.Printf("[Eino] OnStart: name=%s type=%s component=%s input=%v",
			info.Name, info.Type, info.Component, l.formatInput(input))
	}
	return ctx
}

// OnEnd 组件执行成功结束时调用
func (l *Logger) OnEnd(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
	if l.EnableDebug {
		log.Printf("[Eino] OnEnd: name=%s type=%s component=%s output=%v",
			info.Name, info.Type, info.Component, l.formatOutput(output))
	}
	return ctx
}

// OnError 组件执行出错时调用
func (l *Logger) OnError(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
	log.Printf("[Eino] Error: name=%s type=%s component=%s error=%v",
		info.Name, info.Type, info.Component, err)
	return ctx
}

// OnStartWithStreamInput 流式输入开始时调用
func (l *Logger) OnStartWithStreamInput(ctx context.Context, info *callbacks.RunInfo, input *schema.StreamReader[callbacks.CallbackInput]) context.Context {
	if l.EnableDebug {
		log.Printf("[Eino] OnStartWithStreamInput: name=%s type=%s component=%s",
			info.Name, info.Type, info.Component)
	}
	return ctx
}

// OnEndWithStreamOutput 流式输出结束时调用
func (l *Logger) OnEndWithStreamOutput(ctx context.Context, info *callbacks.RunInfo, output *schema.StreamReader[callbacks.CallbackOutput]) context.Context {
	if l.EnableDebug {
		log.Printf("[Eino] OnEndWithStreamOutput: name=%s type=%s component=%s",
			info.Name, info.Type, info.Component)
	}
	return ctx
}

// formatInput 格式化输入数据
func (l *Logger) formatInput(input callbacks.CallbackInput) interface{} {
	if input == nil {
		return nil
	}
	// 简化输出，避免日志过大
	if str, ok := input.(string); ok && len(str) > 200 {
		return str[:200] + "..."
	}
	return input
}

// formatOutput 格式化输出数据
func (l *Logger) formatOutput(output callbacks.CallbackOutput) interface{} {
	if output == nil {
		return nil
	}
	// 简化输出，避免日志过大
	if str, ok := output.(string); ok && len(str) > 200 {
		return str[:200] + "..."
	}
	return output
}

// SetupGlobalCallbacks 设置全局回调
func SetupGlobalCallbacks(enableDebug bool) {
	handler := NewLogger(enableDebug)
	callbacks.AppendGlobalHandlers(handler)
	log.Printf("[Eino] Global callbacks registered (debug=%v)", enableDebug)
}
