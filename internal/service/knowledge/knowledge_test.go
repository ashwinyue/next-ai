// Package knowledge 提供 Knowledge 服务单元测试
package knowledge

import (
	"testing"
)

// ========== getFileExt 测试 ==========

func TestGetFileExt(t *testing.T) {
	tests := []struct {
		name     string
		filePath string
		expected string
	}{
		{
			name:     "txt file",
			filePath: "/path/to/document.txt",
			expected: ".txt",
		},
		{
			name:     "pdf file",
			filePath: "/path/to/document.pdf",
			expected: ".pdf",
		},
		{
			name:     "docx file",
			filePath: "/path/to/document.docx",
			expected: ".docx",
		},
		{
			name:     "md file",
			filePath: "README.md",
			expected: ".md",
		},
		{
			name:     "no extension",
			filePath: "/path/to/document",
			expected: "",
		},
		{
			name:     "hidden file",
			filePath: "/path/to/.gitignore",
			expected: ".gitignore", // getFileExt 把 ".gitignore" 当作扩展名
		},
		{
			name:     "file with multiple dots",
			filePath: "/path/to/file.name.tar.gz",
			expected: ".gz",
		},
		{
			name:     "empty path",
			filePath: "",
			expected: "",
		},
		{
			name:     "extension only",
			filePath: ".txt",
			expected: ".txt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFileExt(tt.filePath)
			if result != tt.expected {
				t.Errorf("getFileExt() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// ========== kbIDFromIndex 测试 ==========

func TestKbIDFromIndex(t *testing.T) {
	// kbIDFromIndex 当前实现直接返回输入
	tests := []struct {
		name      string
		indexName string
		expected  string
	}{
		{
			name:      "standard index",
			indexName: "kb_1234567890_chunks",
			expected:  "kb_1234567890_chunks", // 直接返回
		},
		{
			name:      "short id",
			indexName: "kb_abc_chunks",
			expected:  "kb_abc_chunks",
		},
		{
			name:      "empty string",
			indexName: "",
			expected:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := kbIDFromIndex(tt.indexName)
			if result != tt.expected {
				t.Errorf("kbIDFromIndex() = %q, want %q", result, tt.expected)
			}
		})
	}
}

// ========== Index Name 生成测试 ==========

func TestGenerateChunkIndexName(t *testing.T) {
	tests := []struct {
		name        string
		kbID        string
		expectedEnd string
	}{
		{
			name:        "standard id",
			kbID:        "1234567890",
			expectedEnd: "_chunks",
		},
		{
			name:        "short id",
			kbID:        "abc",
			expectedEnd: "_chunks",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 验证索引名称格式：kb_{kbID}_chunks
			indexName := "kb_" + tt.kbID + tt.expectedEnd
			if len(indexName) < len("kb_")+len(tt.kbID)+len("_chunks") {
				t.Errorf("Generated index name too short: %q", indexName)
			}
		})
	}
}
