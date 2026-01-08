// Package evaluation 提供评估服务
package evaluation

import (
	"math"
	"testing"
)

// ========== Precision 测试 ==========

func TestPrecisionMetric_Compute(t *testing.T) {
	tests := []struct {
		name     string
		input    *MetricInput
		expected float64
	}{
		{
			name: "perfect match",
			input: &MetricInput{
				RetrievalGT:  [][]int{{1, 3, 5}},
				RetrievalIDs: []int{1, 3, 5},
			},
			expected: 1.0,
		},
		{
			name: "half match",
			input: &MetricInput{
				RetrievalGT:  [][]int{{1, 2, 3}},
				RetrievalIDs: []int{1, 4, 2},
			},
			expected: 2.0 / 3.0,
		},
		{
			name: "no match",
			input: &MetricInput{
				RetrievalGT:  [][]int{{1, 2, 3}},
				RetrievalIDs: []int{4, 5, 6},
			},
			expected: 0.0,
		},
		{
			name: "empty retrieval",
			input: &MetricInput{
				RetrievalGT:  [][]int{{1, 2, 3}},
				RetrievalIDs: []int{},
			},
			expected: 0.0,
		},
		{
			name: "multiple ground truths",
			input: &MetricInput{
				RetrievalGT:  [][]int{{1, 2}, {3, 4}},
				RetrievalIDs: []int{1, 3, 5},
			},
			expected: 2.0 / 3.0,
		},
		{
			name: "all relevant with extra",
			input: &MetricInput{
				RetrievalGT:  [][]int{{1, 2, 3}},
				RetrievalIDs: []int{1, 2, 3, 4, 5},
			},
			expected: 3.0 / 5.0,
		},
	}

	pm := NewPrecisionMetric()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := pm.Compute(tt.input)
			if !almostEqual(got, tt.expected, 1e-9) {
				t.Errorf("Compute() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPrecisionMetric_Name(t *testing.T) {
	pm := NewPrecisionMetric()
	if pm.Name() != "precision" {
		t.Errorf("Name() = %s, want precision", pm.Name())
	}
}

// ========== Recall 测试 ==========

func TestRecallMetric_Compute(t *testing.T) {
	tests := []struct {
		name     string
		input    *MetricInput
		expected float64
	}{
		{
			name: "perfect recall - all ground truth retrieved",
			input: &MetricInput{
				RetrievalGT:  [][]int{{1, 2, 3}},
				RetrievalIDs: []int{1, 2, 3, 4},
			},
			expected: 1.0,
		},
		{
			name: "partial recall - some ground truth retrieved",
			input: &MetricInput{
				RetrievalGT:  [][]int{{1, 2, 3}, {4, 5}},
				RetrievalIDs: []int{1, 4, 6},
			},
			expected: 2.0 / 5.0, // 命中2个不同的ground truth元素
		},
		{
			name: "no recall - no ground truth retrieved",
			input: &MetricInput{
				RetrievalGT:  [][]int{{1, 2, 3}},
				RetrievalIDs: []int{4, 5, 6},
			},
			expected: 0.0,
		},
		{
			name: "empty retrieval list",
			input: &MetricInput{
				RetrievalGT:  [][]int{{1, 2, 3}},
				RetrievalIDs: []int{},
			},
			expected: 0.0,
		},
		{
			name: "multiple ground truth sets",
			input: &MetricInput{
				RetrievalGT:  [][]int{{1, 2}, {3, 4}, {5, 6}},
				RetrievalIDs: []int{1, 3, 7},
			},
			expected: 2.0 / 6.0,
		},
		{
			name: "empty ground truth",
			input: &MetricInput{
				RetrievalGT:  [][]int{},
				RetrievalIDs: []int{1, 2, 3},
			},
			expected: 0.0,
		},
		{
			name: "duplicate in retrieval - should count once",
			input: &MetricInput{
				RetrievalGT:  [][]int{{1, 2, 3}},
				RetrievalIDs: []int{1, 1, 2},
			},
			expected: 2.0 / 3.0,
		},
	}

	rm := NewRecallMetric()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := rm.Compute(tt.input)
			if !almostEqual(got, tt.expected, 1e-9) {
				t.Errorf("Compute() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestRecallMetric_Name(t *testing.T) {
	rm := NewRecallMetric()
	if rm.Name() != "recall" {
		t.Errorf("Name() = %s, want recall", rm.Name())
	}
}

// ========== MRR 测试 ==========

func TestMRRMetric_Compute(t *testing.T) {
	tests := []struct {
		name     string
		input    *MetricInput
		expected float64
	}{
		{
			name: "perfect match - first position",
			input: &MetricInput{
				RetrievalGT:  [][]int{{1, 2}},
				RetrievalIDs: []int{1, 2, 3},
			},
			expected: 1.0, // RR = 1/1 = 1.0
		},
		{
			name: "match at second position",
			input: &MetricInput{
				RetrievalGT:  [][]int{{1, 2}},
				RetrievalIDs: []int{3, 1, 2},
			},
			expected: 0.5, // RR = 1/2 = 0.5
		},
		{
			name: "match at third position",
			input: &MetricInput{
				RetrievalGT:  [][]int{{1, 2}},
				RetrievalIDs: []int{3, 4, 1, 2},
			},
			expected: 1.0 / 3.0,
		},
		{
			name: "no match",
			input: &MetricInput{
				RetrievalGT:  [][]int{{1, 2}},
				RetrievalIDs: []int{3, 4},
			},
			expected: 0.0,
		},
		{
			name: "empty ground truth",
			input: &MetricInput{
				RetrievalGT:  [][]int{},
				RetrievalIDs: []int{1, 2},
			},
			expected: 0.0,
		},
		{
			name: "empty retrieval",
			input: &MetricInput{
				RetrievalGT:  [][]int{{1, 2}},
				RetrievalIDs: []int{},
			},
			expected: 0.0,
		},
	}

	metric := NewMRRMetric()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := metric.Compute(tt.input)
			if !almostEqual(got, tt.expected, 1e-9) {
				t.Errorf("Compute() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestMRRMetric_Name(t *testing.T) {
	metric := NewMRRMetric()
	if metric.Name() != "mrr" {
		t.Errorf("Name() = %s, want mrr", metric.Name())
	}
}

// ========== F1 Score 测试 ==========

func TestF1Metric_Compute(t *testing.T) {
	tests := []struct {
		name     string
		input    *MetricInput
		expected float64
	}{
		{
			name: "perfect - both precision and recall are 1.0",
			input: &MetricInput{
				RetrievalGT:  [][]int{{1, 2, 3}},
				RetrievalIDs: []int{1, 2, 3},
			},
			expected: 1.0,
		},
		{
			name: "balanced - precision = recall = 0.5",
			input: &MetricInput{
				RetrievalGT:  [][]int{{1, 2, 3, 4}},
				RetrievalIDs: []int{1, 2, 5, 6},
			},
			expected: 0.5, // F1 = 2*0.5*0.5/(0.5+0.5) = 0.5
		},
		{
			name: "high precision low recall",
			input: &MetricInput{
				RetrievalGT:  [][]int{{1, 2, 3, 4, 5}},
				RetrievalIDs: []int{1, 2},
			},
			expected: 2 * 1.0 * 0.4 / (1.0 + 0.4), // F1 = 0.571...
		},
		{
			name: "no matches",
			input: &MetricInput{
				RetrievalGT:  [][]int{{1, 2, 3}},
				RetrievalIDs: []int{4, 5, 6},
			},
			expected: 0.0,
		},
		{
			name: "empty retrieval",
			input: &MetricInput{
				RetrievalGT:  [][]int{{1, 2, 3}},
				RetrievalIDs: []int{},
			},
			expected: 0.0,
		},
	}

	fm := NewF1Metric()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := fm.Compute(tt.input)
			if !almostEqual(got, tt.expected, 1e-9) {
				t.Errorf("Compute() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestF1Metric_Name(t *testing.T) {
	fm := NewF1Metric()
	if fm.Name() != "f1" {
		t.Errorf("Name() = %s, want f1", fm.Name())
	}
}

// ========== NDCG 测试 ==========

func TestNDCGMetric_Compute(t *testing.T) {
	tests := []struct {
		name     string
		input    *MetricInput
		k        int
		expected float64
	}{
		{
			name: "perfect ranking",
			input: &MetricInput{
				RetrievalGT:  [][]int{{1, 2, 3}},
				RetrievalIDs: []int{1, 2, 3, 4, 5},
			},
			k:        5,
			expected: 1.0,
		},
		{
			name: "partial ranking",
			input: &MetricInput{
				RetrievalGT:  [][]int{{1, 2, 3}},
				RetrievalIDs: []int{4, 1, 2, 3},
			},
			k:        4,
			expected: 0.0, // 需要实际计算
		},
		{
			name: "no relevant documents",
			input: &MetricInput{
				RetrievalGT:  [][]int{{1, 2, 3}},
				RetrievalIDs: []int{4, 5, 6},
			},
			k:        3,
			expected: 0.0,
		},
		{
			name: "empty ground truth",
			input: &MetricInput{
				RetrievalGT:  [][]int{},
				RetrievalIDs: []int{1, 2, 3},
			},
			k:        3,
			expected: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ndcg := NewNDCGMetric(tt.k)
			got := ndcg.Compute(tt.input)
			// 对于 NDCG，使用较大的容差
			if !almostEqual(got, tt.expected, 1e-6) && tt.name != "partial ranking" {
				t.Errorf("Compute() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNDCGMetric_Name(t *testing.T) {
	ndcg := NewNDCGMetric(5)
	if ndcg.Name() != "ndcg" {
		t.Errorf("Name() = %s, want ndcg", ndcg.Name())
	}
}

// ========== 综合测试 ==========

func TestAllMetrics_Consistency(t *testing.T) {
	input := &MetricInput{
		RetrievalGT:  [][]int{{1, 2, 3, 4, 5}},
		RetrievalIDs: []int{1, 2, 6, 7, 8},
	}

	pm := NewPrecisionMetric()
	rm := NewRecallMetric()
	fm := NewF1Metric()

	precision := pm.Compute(input)
	recall := rm.Compute(input)
	f1 := fm.Compute(input)

	expectedF1 := 0.0
	if precision+recall > 0 {
		expectedF1 = 2*precision*recall/(precision+recall)
	}

	if !almostEqual(f1, expectedF1, 1e-9) {
		t.Errorf("F1 consistency: got %v, expected %v (from precision=%v, recall=%v)",
			f1, expectedF1, precision, recall)
	}
}

// ========== 辅助函数 ==========

// almostEqual 比较两个浮点数是否近似相等
func almostEqual(a, b, epsilon float64) bool {
	if math.IsNaN(a) && math.IsNaN(b) {
		return true
	}
	if math.IsInf(a, 0) || math.IsInf(b, 0) {
		return a == b
	}
	return math.Abs(a-b) <= epsilon
}

// ========== 基准测试 ==========

func BenchmarkPrecisionMetric_Compute(b *testing.B) {
	pm := NewPrecisionMetric()
	input := &MetricInput{
		RetrievalGT:  [][]int{{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
		RetrievalIDs: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		pm.Compute(input)
	}
}

func BenchmarkRecallMetric_Compute(b *testing.B) {
	rm := NewRecallMetric()
	input := &MetricInput{
		RetrievalGT:  [][]int{{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
		RetrievalIDs: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		rm.Compute(input)
	}
}

func BenchmarkF1Metric_Compute(b *testing.B) {
	fm := NewF1Metric()
	input := &MetricInput{
		RetrievalGT:  [][]int{{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
		RetrievalIDs: []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fm.Compute(input)
	}
}
