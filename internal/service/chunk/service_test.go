// Package chunk 提供 Chunk 服务单元测试
package chunk

import (
	"testing"

	"github.com/ashwinyue/next-ai/internal/repository"
)

// ========== 错误常量测试 ==========

func TestErrChunkNotFound(t *testing.T) {
	if ErrChunkNotFound == nil {
		t.Error("ErrChunkNotFound should not be nil")
	}
	if ErrChunkNotFound.Error() == "" {
		t.Error("ErrChunkNotFound.Error() should not be empty")
	}
}

// ========== NewService 测试 ==========

func TestNewService_NilRepo(t *testing.T) {
	service := NewService(nil)
	if service == nil {
		t.Error("NewService(nil) returned nil")
	}
	if service.repo != nil {
		t.Error("service.repo should be nil when created with nil")
	}
}

func TestNewService_WithRepo(t *testing.T) {
	repos := &repository.Repositories{}
	service := NewService(repos)
	if service == nil {
		t.Error("NewService() returned nil")
	}
	if service.repo != repos {
		t.Error("service.repo should match the provided repos")
	}
}

// ========== UpdateChunkRequest 测试 ==========

func TestUpdateChunkRequest_AllFields(t *testing.T) {
	content := "updated content"
	index := 5

	req := &UpdateChunkRequest{
		Content:    content,
		ChunkIndex: &index,
	}

	if req.Content != content {
		t.Errorf("Content = %q, want %q", req.Content, content)
	}
	if req.ChunkIndex == nil {
		t.Error("ChunkIndex should not be nil")
	}
	if *req.ChunkIndex != index {
		t.Errorf("ChunkIndex = %d, want %d", *req.ChunkIndex, index)
	}
}

func TestUpdateChunkRequest_OnlyContent(t *testing.T) {
	req := &UpdateChunkRequest{
		Content: "test content",
	}

	if req.Content != "test content" {
		t.Errorf("Content = %q, want 'test content'", req.Content)
	}
	if req.ChunkIndex != nil {
		t.Error("ChunkIndex should be nil when not set")
	}
}

func TestUpdateChunkRequest_OnlyIndex(t *testing.T) {
	index := 10
	req := &UpdateChunkRequest{
		ChunkIndex: &index,
	}

	if req.Content != "" {
		t.Errorf("Content = %q, want empty string", req.Content)
	}
	if req.ChunkIndex == nil {
		t.Error("ChunkIndex should not be nil")
	}
}

func TestUpdateChunkRequest_Empty(t *testing.T) {
	req := &UpdateChunkRequest{}

	if req.Content != "" {
		t.Errorf("Content = %q, want empty string", req.Content)
	}
	if req.ChunkIndex != nil {
		t.Error("ChunkIndex should be nil when not set")
	}
}

// ========== 边界值测试 ==========

func TestUpdateChunkRequest_ZeroIndex(t *testing.T) {
	index := 0
	req := &UpdateChunkRequest{
		ChunkIndex: &index,
	}

	if req.ChunkIndex == nil {
		t.Error("ChunkIndex should not be nil even when set to 0")
	}
	if *req.ChunkIndex != 0 {
		t.Errorf("ChunkIndex = %d, want 0", *req.ChunkIndex)
	}
}

func TestUpdateChunkRequest_NegativeIndex(t *testing.T) {
	index := -1
	req := &UpdateChunkRequest{
		ChunkIndex: &index,
	}

	if req.ChunkIndex == nil {
		t.Error("ChunkIndex should not be nil even when negative")
	}
	if *req.ChunkIndex != -1 {
		t.Errorf("ChunkIndex = %d, want -1", *req.ChunkIndex)
	}
}

func TestUpdateChunkRequest_LargeIndex(t *testing.T) {
	index := 999999
	req := &UpdateChunkRequest{
		ChunkIndex: &index,
	}

	if req.ChunkIndex == nil {
		t.Error("ChunkIndex should not be nil")
	}
	if *req.ChunkIndex != 999999 {
		t.Errorf("ChunkIndex = %d, want 999999", *req.ChunkIndex)
	}
}

// ========== Service 字段测试 ==========

func TestService_RepoField(t *testing.T) {
	repos := &repository.Repositories{}
	service := NewService(repos)

	// 验证 repo 字段正确设置
	if service.repo == nil {
		t.Error("service.repo should not be nil when created with repos")
	}
}

func TestService_NilRepoField(t *testing.T) {
	service := NewService(nil)

	// 验证 nil repo 不会崩溃
	if service == nil {
		t.Error("NewService(nil) should still return a service")
	}
	if service.repo != nil {
		t.Error("service.repo should be nil when created with nil")
	}
}
