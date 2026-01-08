// Package rag 提供 Parent 检索器功能单元测试
package rag

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/schema"
)

// ========== NewParentRetriever 测试 ==========

func TestNewParentRetriever(t *testing.T) {
	ctx := context.Background()
	childDoc := newDoc("child content", 0.8)

	tests := []struct {
		name        string
		cfg         *ParentConfig
		wantErr     bool
		errContains string
	}{
		{
			name: "valid config",
			cfg: &ParentConfig{
				Retriever: newMockRetriever([]*schema.Document{childDoc}, nil),
				ParentIDKey: "parent_id",
				OrigDocGetter: func(ctx context.Context, ids []string) ([]*schema.Document, error) {
					return []*schema.Document{newDoc("parent", 1.0)}, nil
				},
			},
			wantErr: false,
		},
		{
			name: "nil retriever",
			cfg: &ParentConfig{
				Retriever: nil,
				ParentIDKey: "parent_id",
				OrigDocGetter: func(ctx context.Context, ids []string) ([]*schema.Document, error) {
					return []*schema.Document{}, nil
				},
			},
			wantErr:     true,
			errContains: "Retriever is required",
		},
		{
			name: "nil orig doc getter",
			cfg: &ParentConfig{
				Retriever: newMockRetriever([]*schema.Document{childDoc}, nil),
				ParentIDKey: "parent_id",
				OrigDocGetter: nil,
			},
			wantErr:     true,
			errContains: "OrigDocGetter is required",
		},
		{
			name: "empty parent id key",
			cfg: &ParentConfig{
				Retriever: newMockRetriever([]*schema.Document{childDoc}, nil),
				ParentIDKey: "",
				OrigDocGetter: func(ctx context.Context, ids []string) ([]*schema.Document, error) {
					return []*schema.Document{}, nil
				},
			},
			wantErr:     true,
			errContains: "ParentIDKey is required",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r, err := NewParentRetriever(ctx, tt.cfg)

			if tt.wantErr {
				if err == nil {
					t.Errorf("NewParentRetriever() expected error, got nil")
				}
				if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("Error = %v, want contain %q", err, tt.errContains)
				}
				return
			}

			if err != nil {
				t.Errorf("NewParentRetriever() unexpected error: %v", err)
			}
			if r == nil {
				t.Error("NewParentRetriever() returned nil retriever")
			}
		})
	}
}

// ========== MapGetter 测试 ==========

func TestMapGetter(t *testing.T) {
	ctx := context.Background()

	// 创建测试文档
	docs := map[string]*schema.Document{
		"doc1": {ID: "doc1", Content: "content 1"},
		"doc2": {ID: "doc2", Content: "content 2"},
		"doc3": {ID: "doc3", Content: "content 3"},
	}

	getter := MapGetter(docs)

	tests := []struct {
		name          string
		ids           []string
		expectedCount int
	}{
		{
			name:          "get all existing docs",
			ids:           []string{"doc1", "doc2", "doc3"},
			expectedCount: 3,
		},
		{
			name:          "get partial docs",
			ids:           []string{"doc1", "doc2"},
			expectedCount: 2,
		},
		{
			name:          "get non-existent docs",
			ids:           []string{"doc4", "doc5"},
			expectedCount: 0,
		},
		{
			name:          "mixed existing and non-existing",
			ids:           []string{"doc1", "doc4", "doc2"},
			expectedCount: 2,
		},
		{
			name:          "empty ids",
			ids:           []string{},
			expectedCount: 0,
		},
		{
			name:          "nil ids",
			ids:           nil,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getter(ctx, tt.ids)

			if err != nil {
				t.Errorf("MapGetter() unexpected error: %v", err)
			}
			if len(result) != tt.expectedCount {
				t.Errorf("MapGetter() returned %d docs, want %d", len(result), tt.expectedCount)
			}
		})
	}
}

// ========== SliceGetter 测试 ==========

func TestSliceGetter(t *testing.T) {
	ctx := context.Background()

	// 创建测试文档
	allDocs := []*schema.Document{
		{ID: "doc1", Content: "content 1"},
		{ID: "doc2", Content: "content 2"},
		{ID: "doc3", Content: "content 3"},
	}

	getter := SliceGetter(allDocs)

	tests := []struct {
		name          string
		ids           []string
		expectedCount int
	}{
		{
			name:          "get all existing docs",
			ids:           []string{"doc1", "doc2", "doc3"},
			expectedCount: 3,
		},
		{
			name:          "get partial docs",
			ids:           []string{"doc1", "doc2"},
			expectedCount: 2,
		},
		{
			name:          "get non-existent docs",
			ids:           []string{"doc4", "doc5"},
			expectedCount: 0,
		},
		{
			name:          "mixed existing and non-existing",
			ids:           []string{"doc1", "doc4", "doc2"},
			expectedCount: 2,
		},
		{
			name:          "empty ids",
			ids:           []string{},
			expectedCount: 0,
		},
		{
			name:          "nil ids",
			ids:           nil,
			expectedCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getter(ctx, tt.ids)

			if err != nil {
				t.Errorf("SliceGetter() unexpected error: %v", err)
			}
			if len(result) != tt.expectedCount {
				t.Errorf("SliceGetter() returned %d docs, want %d", len(result), tt.expectedCount)
			}
		})
	}
}

func TestSliceGetter_EmptySlice(t *testing.T) {
	ctx := context.Background()
	getter := SliceGetter([]*schema.Document{})

	result, err := getter(ctx, []string{"doc1"})

	if err != nil {
		t.Errorf("SliceGetter() unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("SliceGetter() returned %d docs, want 0", len(result))
	}
}

func TestSliceGetter_NilSlice(t *testing.T) {
	ctx := context.Background()
	getter := SliceGetter(nil)

	result, err := getter(ctx, []string{"doc1"})

	if err != nil {
		t.Errorf("SliceGetter() unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("SliceGetter() returned %d docs, want 0", len(result))
	}
}

// ========== AddParentID 测试 ==========

func TestAddParentID(t *testing.T) {
	tests := []struct {
		name     string
		doc      *schema.Document
		parentID string
	}{
		{
			name: "doc with existing metadata",
			doc: &schema.Document{
				ID: "child1",
				MetaData: map[string]any{
					"existing_key": "existing_value",
				},
			},
			parentID: "parent1",
		},
		{
			name: "doc without metadata",
			doc: &schema.Document{
				ID: "child2",
				MetaData: nil,
			},
			parentID: "parent2",
		},
		{
			name: "doc with empty metadata",
			doc: &schema.Document{
				ID: "child3",
				MetaData: map[string]any{},
			},
			parentID: "parent3",
		},
		{
			name: "doc with existing parent_id",
			doc: &schema.Document{
				ID: "child4",
				MetaData: map[string]any{
					"parent_id": "old_parent",
				},
			},
			parentID: "new_parent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := AddParentID(tt.doc, tt.parentID)

			if result == nil {
				t.Fatal("AddParentID() returned nil")
			}
			if result.MetaData == nil {
				t.Error("AddParentID() result has nil MetaData")
			} else if result.MetaData["parent_id"] != tt.parentID {
				t.Errorf("parent_id = %v, want %q", result.MetaData["parent_id"], tt.parentID)
			}
		})
	}
}

func TestAddParentID_PreservesOtherMetadata(t *testing.T) {
	doc := &schema.Document{
		ID: "child1",
		MetaData: map[string]any{
			"existing_key": "existing_value",
			"another_key":  123,
		},
	}

	result := AddParentID(doc, "parent1")

	if result.MetaData["existing_key"] != "existing_value" {
		t.Error("AddParentID() did not preserve existing metadata")
	}
	if result.MetaData["another_key"] != 123 {
		t.Error("AddParentID() did not preserve all existing metadata")
	}
}

// ========== WithParentMetadata 测试 ==========

func TestWithParentMetadata(t *testing.T) {
	tests := []struct {
		name           string
		id             string
		content        string
		parentID       string
		additionalMeta map[string]any
		expectedMeta   map[string]any
	}{
		{
			name:           "doc without additional metadata",
			id:             "child1",
			content:        "child content",
			parentID:       "parent1",
			additionalMeta: nil,
			expectedMeta: map[string]any{
				"parent_id": "parent1",
			},
		},
		{
			name:     "doc with additional metadata",
			id:       "child2",
			content:  "child content 2",
			parentID: "parent2",
			additionalMeta: map[string]any{
				"chunk_index": 0,
				"source":     "test",
			},
			expectedMeta: map[string]any{
				"parent_id":   "parent2",
				"chunk_index": 0,
				"source":      "test",
			},
		},
		{
			name:           "doc with empty additional metadata",
			id:             "child3",
			content:        "child content 3",
			parentID:       "parent3",
			additionalMeta: map[string]any{},
			expectedMeta: map[string]any{
				"parent_id": "parent3",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WithParentMetadata(tt.id, tt.content, tt.parentID, tt.additionalMeta)

			if result == nil {
				t.Fatal("WithParentMetadata() returned nil")
			}
			if result.ID != tt.id {
				t.Errorf("ID = %q, want %q", result.ID, tt.id)
			}
			if result.Content != tt.content {
				t.Errorf("Content = %q, want %q", result.Content, tt.content)
			}
			if result.MetaData == nil {
				t.Fatal("MetaData is nil")
			}

			for k, v := range tt.expectedMeta {
				if result.MetaData[k] != v {
					t.Errorf("MetaData[%q] = %v, want %v", k, result.MetaData[k], v)
				}
			}
		})
	}
}

func TestWithParentMetadata_ConflictingKey(t *testing.T) {
	// 如果 additionalMeta 中有 parent_id，additionalMeta 会覆盖设置的值
	additionalMeta := map[string]any{
		"parent_id": "additional_parent", // 这个会覆盖后面设置的 "new_parent"
		"other_key": "other_value",
	}

	result := WithParentMetadata("child1", "content", "new_parent", additionalMeta)

	// 实现：先设置 parent_id，然后用 additionalMeta 覆盖
	// 所以结果是 additional_parent
	if result.MetaData["parent_id"] != "additional_parent" {
		t.Errorf("parent_id should be 'additional_parent' (from additionalMeta), got %v", result.MetaData["parent_id"])
	}
	if result.MetaData["other_key"] != "other_value" {
		t.Errorf("other_key should be preserved")
	}
}
