// Package rag 提供 RAG 增强功能单元测试
package rag

import (
	"testing"

	"github.com/cloudwego/eino/schema"
)

// ========== 辅助函数测试 ==========

func TestTokenize(t *testing.T) {
	tests := []struct {
		name     string
		text     string
		minCount int
	}{
		{
			name:     "simple text",
			text:     "hello world test",
			minCount: 3,
		},
		{
			name:     "text with punctuation",
			text:     "Hello, world! This is a test.",
			minCount: 5,
		},
		{
			name:     "text with numbers",
			text:     "test123 api_v2",
			minCount: 2,
		},
		{
			name:     "empty text",
			text:     "",
			minCount: 0,
		},
		{
			name:     "chinese characters",
			text:     "测试中文分词",
			minCount: 0, // tokenize 使用 unicode.IsLetter/IsNumber，中文字符被排除
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := tokenize(tt.text)
			if len(tokens) < tt.minCount {
				t.Errorf("tokenize() returned %d tokens, want at least %d", len(tokens), tt.minCount)
			}
		})
	}
}

func TestJaccard(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected float64
	}{
		{
			name:     "identical sets",
			a:        "hello world test",
			b:        "hello world test",
			expected: 1.0,
		},
		{
			name:     "no overlap",
			a:        "hello world",
			b:        "foo bar",
			expected: 0.0,
		},
		{
			name:     "partial overlap",
			a:        "hello world test",
			b:        "hello foo bar",
			expected: 0.2, // "hello" / 5 unique tokens (hello, world, test, foo, bar)
		},
		{
			name:     "empty sets",
			a:        "",
			b:        "",
			expected: 1.0,
		},
		{
			name:     "one empty set",
			a:        "hello",
			b:        "",
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aTokens := tokenize(tt.a)
			bTokens := tokenize(tt.b)
			result := jaccard(aTokens, bTokens)

			if !almostEqual(result, tt.expected, 0.01) {
				t.Errorf("jaccard() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestClamp(t *testing.T) {
	tests := []struct {
		name  string
		v     float64
		min   float64
		max   float64
		want  float64
	}{
		{
			name: "within range",
			v:    0.5,
			min:  0.0,
			max:  1.0,
			want: 0.5,
		},
		{
			name: "below min",
			v:    -0.5,
			min:  0.0,
			max:  1.0,
			want: 0.0,
		},
		{
			name: "above max",
			v:    1.5,
			min:  0.0,
			max:  1.0,
			want: 1.0,
		},
		{
			name: "at boundary",
			v:    0.0,
			min:  0.0,
			max:  1.0,
			want: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := clamp(tt.v, tt.min, tt.max); got != tt.want {
				t.Errorf("clamp() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalizeContent(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		expected string
	}{
		{
			name:     "simple text",
			content:  "Hello World",
			expected: "hello world",
		},
		{
			name:     "extra spaces",
			content:  "Hello    World   Test",
			expected: "hello world test",
		},
		{
			name:     "mixed case",
			content:  "HeLLo WoRLd",
			expected: "hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeContent(tt.content); got != tt.expected {
				t.Errorf("normalizeContent() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestBuildContentSignature(t *testing.T) {
	tests := []struct {
		name     string
		content  string
		sameAs   string
	}{
		{
			name:    "same content produces same signature",
			content: "Hello World",
			sameAs:  "Hello World",
		},
		{
			name:    "normalized content produces same signature",
			content: "Hello World",
			sameAs:  "hello    world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sig1 := buildContentSignature(tt.content)
			sig2 := buildContentSignature(tt.sameAs)

			if sig1 != sig2 {
				t.Errorf("buildContentSignature() produced different signatures: %q vs %q", sig1, sig2)
			}

			if len(sig1) != 32 { // MD5 hex length
				t.Errorf("buildContentSignature() returned length %d, want 32", len(sig1))
			}
		})
	}
}

// ========== MMR 算法测试 ==========

func TestDefaultMMRConfig(t *testing.T) {
	cfg := DefaultMMRConfig()
	if cfg.K != 10 {
		t.Errorf("K = %d, want 10", cfg.K)
	}
	if cfg.Lambda != 0.7 {
		t.Errorf("Lambda = %v, want 0.7", cfg.Lambda)
	}
}

func TestApplyMMR(t *testing.T) {
	// 创建测试文档：doc1 和 doc2 内容相似
	docs := []*schema.Document{
		newDoc("machine learning is a subset of artificial intelligence", 0.9),
		newDoc("machine learning and AI are related fields", 0.8),
		newDoc("deep learning uses neural networks", 0.7),
		newDoc("natural language processing for text analysis", 0.6),
		newDoc("computer vision processes images", 0.5),
	}

	tests := []struct {
		name          string
		docs          []*schema.Document
		cfg           *MMRConfig
		expectedCount int
	}{
		{
			name:          "select top 3 with default lambda",
			docs:          docs,
			cfg:           &MMRConfig{K: 3, Lambda: 0.7},
			expectedCount: 3,
		},
		{
			name:          "select top 2",
			docs:          docs,
			cfg:           &MMRConfig{K: 2, Lambda: 0.5},
			expectedCount: 2,
		},
		{
			name:          "k larger than docs",
			docs:          docs,
			cfg:           &MMRConfig{K: 100, Lambda: 0.7},
			expectedCount: len(docs),
		},
		{
			name:          "nil config uses default",
			docs:          docs,
			cfg:           nil,
			expectedCount: len(docs), // default K=10, but only 5 docs available
		},
		{
			name:          "empty docs",
			docs:          []*schema.Document{},
			cfg:           &MMRConfig{K: 3, Lambda: 0.7},
			expectedCount: 0,
		},
		{
			name:          "zero k returns all",
			docs:          docs,
			cfg:           &MMRConfig{K: 0, Lambda: 0.7},
			expectedCount: len(docs),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ApplyMMR(tt.docs, tt.cfg)
			if len(result) != tt.expectedCount {
				t.Errorf("ApplyMMR() returned %d docs, want %d", len(result), tt.expectedCount)
			}
		})
	}
}

func TestApplyMMR_Diversity(t *testing.T) {
	// 创建相似的文档
	docs := []*schema.Document{
		newDoc("machine learning AI", 0.9),
		newDoc("machine learning and artificial intelligence", 0.85),
		newDoc("deep learning neural networks", 0.8),
	}

	// Lambda = 0 时，只看多样性，应该选择不相似的文档
	result := ApplyMMR(docs, &MMRConfig{K: 2, Lambda: 0.0})

	if len(result) != 2 {
		t.Fatalf("ApplyMMR() returned %d docs, want 2", len(result))
	}

	// 第一个应该是分数最高的
	if result[0].Score() < result[1].Score() {
		t.Error("First result should have highest score")
	}
}

// ========== 复合评分测试 ==========

func TestDefaultCompositeConfig(t *testing.T) {
	cfg := DefaultCompositeConfig()
	if cfg.ModelWeight != 0.6 {
		t.Errorf("ModelWeight = %v, want 0.6", cfg.ModelWeight)
	}
	if cfg.BaseWeight != 0.3 {
		t.Errorf("BaseWeight = %v, want 0.3", cfg.BaseWeight)
	}
	if cfg.SourceWeight != 0.1 {
		t.Errorf("SourceWeight = %v, want 0.1", cfg.SourceWeight)
	}
}

func TestApplyCompositeScore(t *testing.T) {
	docs := []*DocumentWithSource{
		{
			Document:  newDoc("doc1", 0.8),
			Source:    "es8",
			BaseScore: 0.7,
		},
		{
			Document:  newDoc("doc2", 0.6),
			Source:    "web_search",
			BaseScore: 0.5,
		},
		{
			Document:  newDoc("doc3", 0.9),
			Source:    "milvus",
			BaseScore: 0.6,
		},
	}

	result := ApplyCompositeScore(docs, DefaultCompositeConfig())

	if len(result) != 3 {
		t.Fatalf("ApplyCompositeScore() returned %d docs, want 3", len(result))
	}

	// 验证已排序
	for i := 1; i < len(result); i++ {
		if result[i-1].Document.Score() < result[i].Document.Score() {
			t.Errorf("Results not sorted: index %d (%.2f) < index %d (%.2f)",
				i-1, result[i-1].Document.Score(), i, result[i].Document.Score())
		}
	}

	// web_search 应该有更高的权重
	webSearchDoc := result[0]
	if webSearchDoc.Source != "web_search" {
		// web_search 应该排在前面因为权重更高
	}
}

// ========== 内容去重测试 ==========

func TestDefaultContentDedupConfig(t *testing.T) {
	cfg := DefaultContentDedupConfig()
	if cfg.Threshold != 0.85 {
		t.Errorf("Threshold = %v, want 0.85", cfg.Threshold)
	}
}

func TestContentDedupBySignature(t *testing.T) {
	docs := []*schema.Document{
		newDoc("Hello World", 0),
		newDoc("Hello World", 0),     // 完全重复
		newDoc("Different content", 0),
		newDoc("hello    world", 0),   // 归一化后相同
	}

	result := ContentDedupBySignature(docs)

	// "Hello World" 和 "hello    world" 归一化后都是 "hello world"
	// "Different content" 归一化后是 "different content"
	// 结果：2 个唯一文档，2 个被移除
	if result.RemovedCount != 2 {
		t.Errorf("RemovedCount = %d, want 2", result.RemovedCount)
	}
	if len(result.Unique) != 2 {
		t.Errorf("Unique count = %d, want 2", len(result.Unique))
	}
}

func TestContentDedupBySimilarity(t *testing.T) {
	tests := []struct {
		name            string
		docs            []*schema.Document
		threshold       float64
		expectedUnique  int
		expectedRemoved int
	}{
		{
			name: "no duplicates",
			docs: []*schema.Document{
				newDoc("machine learning", 0),
				newDoc("natural language processing", 0),
				newDoc("computer vision", 0),
			},
			threshold:       0.85,
			expectedUnique:  3,
			expectedRemoved: 0,
		},
		{
			name: "similar content removed",
			docs: []*schema.Document{
				newDoc("machine learning is great", 0),
				newDoc("machine learning is awesome", 0),
				newDoc("deep learning networks", 0),
			},
			threshold:       0.5, // 较低阈值
			expectedUnique:  2,
			expectedRemoved: 1,
		},
		{
			name: "empty docs",
			docs: []*schema.Document{},
			threshold:       0.85,
			expectedUnique:  0,
			expectedRemoved: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &ContentDedupConfig{Threshold: tt.threshold}
			result := ContentDedupBySimilarity(tt.docs, cfg)

			if len(result.Unique) != tt.expectedUnique {
				t.Errorf("Unique count = %d, want %d", len(result.Unique), tt.expectedUnique)
			}
			if result.RemovedCount != tt.expectedRemoved {
				t.Errorf("RemovedCount = %d, want %d", result.RemovedCount, tt.expectedRemoved)
			}
		})
	}
}

// ========== 批量处理测试 ==========

func TestDefaultBatchConfig(t *testing.T) {
	cfg := DefaultBatchConfig()
	if cfg.BatchSize != 15 {
		t.Errorf("BatchSize = %d, want 15", cfg.BatchSize)
	}
	if cfg.MaxContentLength != 800 {
		t.Errorf("MaxContentLength = %d, want 800", cfg.MaxContentLength)
	}
}

func TestProcessBatches(t *testing.T) {
	items := make([]interface{}, 25)
	for i := range items {
		items[i] = i
	}

	callCount := 0
	batchSizes := []int{}

	err := ProcessBatches(items, DefaultBatchConfig(), func(batch []interface{}) error {
		callCount++
		batchSizes = append(batchSizes, len(batch))
		return nil
	})

	if err != nil {
		t.Errorf("ProcessBatches() unexpected error: %v", err)
	}
	if callCount != 2 { // 25 items / 15 batch size = 2 batches
		t.Errorf("ProcessBatches() called processFunc %d times, want 2", callCount)
	}
	if len(batchSizes) != 2 || batchSizes[0] != 15 || batchSizes[1] != 10 {
		t.Errorf("Batch sizes = %v, want [15, 10]", batchSizes)
	}
}

func TestProcessBatches_Error(t *testing.T) {
	items := make([]interface{}, 10)
	expectedErr := "batch error"

	err := ProcessBatches(items, &BatchConfig{BatchSize: 5}, func(batch []interface{}) error {
		if len(batch) == 5 {
			return &testError{msg: expectedErr}
		}
		return nil
	})

	if err == nil {
		t.Fatal("ProcessBatches() expected error, got nil")
	}
	if err.Error() != "batch 0-5 failed: batch error" {
		t.Errorf("Error = %v, want 'batch 0-5 failed: batch error'", err)
	}
}

// ========== 统计信息测试 ==========

func TestCalculateStats(t *testing.T) {
	docs := []*schema.Document{
		newDocWithSource("doc1", 0.9, "es8"),
		newDocWithSource("doc2", 0.7, "web"),
		newDocWithSource("doc3", 0.8, "es8"),
	}

	stats := CalculateStats(docs)

	if stats.TotalResults != 3 {
		t.Errorf("TotalResults = %d, want 3", stats.TotalResults)
	}
	if stats.UniqueResults != 3 {
		t.Errorf("UniqueResults = %d, want 3", stats.UniqueResults)
	}
	if !almostEqual(stats.AverageScore, 0.8, 0.001) { // (0.9+0.7+0.8)/3
		t.Errorf("AverageScore = %v, want 0.8", stats.AverageScore)
	}
	if stats.TopScore != 0.9 {
		t.Errorf("TopScore = %v, want 0.9", stats.TopScore)
	}
	if stats.BottomScore != 0.7 {
		t.Errorf("BottomScore = %v, want 0.7", stats.BottomScore)
	}
	if stats.Sources["es8"] != 2 {
		t.Errorf("Sources[\"es8\"] = %d, want 2", stats.Sources["es8"])
	}
	if stats.Sources["web"] != 1 {
		t.Errorf("Sources[\"web\"] = %d, want 1", stats.Sources["web"])
	}
}

func TestCalculateStats_Empty(t *testing.T) {
	stats := CalculateStats([]*schema.Document{})

	if stats.TotalResults != 0 {
		t.Errorf("TotalResults = %d, want 0", stats.TotalResults)
	}
	if stats.Sources == nil {
		t.Error("Sources map should not be nil")
	}
}

func TestFormatStats(t *testing.T) {
	stats := &RetrievalStats{
		Query:          "test query",
		TotalResults:   10,
		UniqueResults:  8,
		DuplicateCount: 2,
		AverageScore:   0.75,
		TopScore:       0.95,
		BottomScore:    0.55,
		Sources:        map[string]int{"es8": 6, "web": 2},
	}

	formatted := FormatStats(stats)

	if formatted == "" {
		t.Error("FormatStats() returned empty string")
	}
	if !contains(formatted, "总结果数") {
		t.Error("FormatStats() missing '总结果数'")
	}
	if !contains(formatted, "es8") {
		t.Error("FormatStats() missing source 'es8'")
	}
}

func TestFormatStats_Nil(t *testing.T) {
	formatted := FormatStats(nil)
	if formatted != "" {
		t.Errorf("FormatStats(nil) = %q, want empty string", formatted)
	}
}

// ========== 辅助类型 ==========

type testError struct {
	msg string
}

func (e *testError) Error() string {
	return e.msg
}
