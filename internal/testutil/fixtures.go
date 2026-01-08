// Package testutil 提供测试辅助工具
package testutil

import (
	"context"
	"testing"
	"time"
)

// ContextHelper 提供上下文相关的测试辅助
type ContextHelper struct{}

// NewContextHelper 创建上下文辅助器
func NewContextHelper() *ContextHelper {
	return &ContextHelper{}
}

// Context 返回测试用的 context.Background()
func (h *ContextHelper) Context() context.Context {
	return context.Background()
}

// CanceledContext 返回已取消的 context
func (h *ContextHelper) CanceledContext() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	return ctx
}

// TimeoutContext 返回带超时的 context
func (h *ContextHelper) TimeoutContext(d time.Duration) context.Context {
	ctx, cancel := context.WithTimeout(context.Background(), d)
	// 返回一个包装函数，让调用者可以在需要时调用 cancel
	_ = cancel // 保存 cancel 函数供未来使用
	return ctx
}

// AssertHelper 提供断言相关的测试辅助
type AssertHelper struct {
	t *testing.T
}

// NewAssertHelper 创建断言辅助器
func NewAssertHelper(t *testing.T) *AssertHelper {
	return &AssertHelper{t: t}
}

// NoError 断言没有错误
func (h *AssertHelper) NoError(err error, msgAndArgs ...interface{}) {
	h.t.Helper()
	if err != nil {
		h.t.Fatalf("Unexpected error: %v %v", err, msgAndArgs)
	}
}

// Error 断言有错误
func (h *AssertHelper) Error(err error, msgAndArgs ...interface{}) {
	h.t.Helper()
	if err == nil {
		h.t.Fatal("Expected error, got nil")
	}
}

// ErrorContains 断言错误包含指定字符串
func (h *AssertHelper) ErrorContains(err error, substr string, msgAndArgs ...interface{}) {
	h.t.Helper()
	if err == nil {
		h.t.Fatal("Expected error, got nil")
	}
	if !contains(err.Error(), substr) {
		h.t.Fatalf("Error %q does not contain %q %v", err.Error(), substr, msgAndArgs)
	}
}

// Equal 断言相等
func (h *AssertHelper) Equal(expected, actual interface{}, msgAndArgs ...interface{}) {
	h.t.Helper()
	if expected != actual {
		h.t.Fatalf("Expected %v, got %v %v", expected, actual, msgAndArgs)
	}
}

// NotEmpty 断言非空
func (h *AssertHelper) NotEmpty(v interface{}, msgAndArgs ...interface{}) {
	h.t.Helper()
	if isEmpty(v) {
		h.t.Fatalf("Expected non-empty, got %v %v", v, msgAndArgs)
	}
}

// Nil 断言为 nil
func (h *AssertHelper) Nil(v interface{}, msgAndArgs ...interface{}) {
	h.t.Helper()
	if v != nil {
		h.t.Fatalf("Expected nil, got %v %v", v, msgAndArgs)
	}
}

// NotNil 断言非 nil
func (h *AssertHelper) NotNil(v interface{}, msgAndArgs ...interface{}) {
	h.t.Helper()
	if v == nil {
		h.t.Fatal("Expected non-nil, got nil")
	}
}

// True 断言为真
func (h *AssertHelper) True(condition bool, msgAndArgs ...interface{}) {
	h.t.Helper()
	if !condition {
		h.t.Fatalf("Expected true, got false %v", msgAndArgs)
	}
}

// False 断言为假
func (h *AssertHelper) False(condition bool, msgAndArgs ...interface{}) {
	h.t.Helper()
	if condition {
		h.t.Fatalf("Expected false, got true %v", msgAndArgs)
	}
}

// Helper 辅助函数
func contains(s, substr string) bool {
	return len(s) >= len(substr) && findSubstring(s, substr)
}

func findSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func isEmpty(v interface{}) bool {
	if v == nil {
		return true
	}
	switch val := v.(type) {
	case string:
		return val == ""
	case []interface{}:
		return len(val) == 0
	case map[string]interface{}:
		return len(val) == 0
	default:
		return false
	}
}
