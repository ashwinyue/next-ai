// Package evaluation 提供评估服务
package evaluation

import (
	"math"
)

// MetricInput 指标计算输入
type MetricInput struct {
	// RetrievalGT 是 Ground Truth 标注的相关文档集合
	// 每个内层 []int 代表一个查询对应的相关文档ID列表
	RetrievalGT [][]int

	// RetrievalIDs 是检索系统返回的文档ID列表
	RetrievalIDs []int
}

// Metric 指标接口
type Metric interface {
	Compute(input *MetricInput) float64
	Name() string
}

// ========== Precision 精确率 ==========

// PrecisionMetric 精确率指标
// Precision = 命中的相关文档数 / 返回的文档总数
type PrecisionMetric struct{}

// NewPrecisionMetric 创建精确率指标
func NewPrecisionMetric() *PrecisionMetric {
	return &PrecisionMetric{}
}

// Compute 计算精确率
func (m *PrecisionMetric) Compute(input *MetricInput) float64 {
	if len(input.RetrievalIDs) == 0 {
		return 0.0
	}

	// 计算命中的文档数
	hitCount := 0
	for _, id := range input.RetrievalIDs {
		if m.isInGroundTruth(id, input.RetrievalGT) {
			hitCount++
		}
	}

	return float64(hitCount) / float64(len(input.RetrievalIDs))
}

// Name 返回指标名称
func (m *PrecisionMetric) Name() string {
	return "precision"
}

// isInGroundTruth 检查文档ID是否在Ground Truth中
func (m *PrecisionMetric) isInGroundTruth(id int, groundTruth [][]int) bool {
	for _, gtSet := range groundTruth {
		for _, gtID := range gtSet {
			if id == gtID {
				return true
			}
		}
	}
	return false
}

// ========== Recall 召回率 ==========

// RecallMetric 召回率指标
// Recall = 命中的相关文档数 / Ground Truth 中的相关文档总数
type RecallMetric struct{}

// NewRecallMetric 创建召回率指标
func NewRecallMetric() *RecallMetric {
	return &RecallMetric{}
}

// Compute 计算召回率
func (m *RecallMetric) Compute(input *MetricInput) float64 {
	if len(input.RetrievalGT) == 0 {
		return 0.0
	}

	// 收集所有 Ground Truth 文档ID（去重）
	gtSet := make(map[int]bool)
	for _, gtList := range input.RetrievalGT {
		for _, id := range gtList {
			gtSet[id] = true
		}
	}

	totalGT := len(gtSet)
	if totalGT == 0 {
		return 0.0
	}

	// 计算命中的唯一文档数（去重）
	hitSet := make(map[int]bool)
	for _, id := range input.RetrievalIDs {
		if gtSet[id] {
			hitSet[id] = true
		}
	}

	return float64(len(hitSet)) / float64(totalGT)
}

// Name 返回指标名称
func (m *RecallMetric) Name() string {
	return "recall"
}

// ========== MRR 平均倒数排名 ==========

// MRRMetric 平均倒数排名指标
// MRR = 1/n * Σ(1/rank_i)，其中 rank_i 是第i个查询的第一个相关文档的排名
type MRRMetric struct{}

// NewMRRMetric 创建 MRR 指标
func NewMRRMetric() *MRRMetric {
	return &MRRMetric{}
}

// Compute 计算 MRR
func (m *MRRMetric) Compute(input *MetricInput) float64 {
	if len(input.RetrievalGT) == 0 {
		return 0.0
	}

	// 收集所有 Ground Truth 文档ID
	gtSet := make(map[int]bool)
	for _, gtList := range input.RetrievalGT {
		for _, id := range gtList {
			gtSet[id] = true
		}
	}

	if len(gtSet) == 0 {
		return 0.0
	}

	// 找到第一个相关文档的排名
	reciprocalRank := 0.0
	for rank, id := range input.RetrievalIDs {
		if gtSet[id] {
			reciprocalRank = 1.0 / float64(rank+1)
			break
		}
	}

	return reciprocalRank
}

// Name 返回指标名称
func (m *MRRMetric) Name() string {
	return "mrr"
}

// ========== F1 Score ==========

// F1Metric F1 分数指标
// F1 = 2 * Precision * Recall / (Precision + Recall)
type F1Metric struct {
	precision *PrecisionMetric
	recall    *RecallMetric
}

// NewF1Metric 创建 F1 指标
func NewF1Metric() *F1Metric {
	return &F1Metric{
		precision: NewPrecisionMetric(),
		recall:    NewRecallMetric(),
	}
}

// Compute 计算 F1 分数
func (m *F1Metric) Compute(input *MetricInput) float64 {
	precision := m.precision.Compute(input)
	recall := m.recall.Compute(input)

	if precision+recall == 0 {
		return 0.0
	}

	return 2 * precision * recall / (precision + recall)
}

// Name 返回指标名称
func (m *F1Metric) Name() string {
	return "f1"
}

// ========== NDCG (Normalized Discounted Cumulative Gain) ==========

// NDCGMetric NDCG 指标
type NDCGMetric struct {
	k int // 计算@k的NDCG
}

// NewNDCGMetric 创建 NDCG 指标
func NewNDCGMetric(k int) *NDCGMetric {
	return &NDCGMetric{k: k}
}

// Compute 计算 NDCG
func (m *NDCGMetric) Compute(input *MetricInput) float64 {
	if m.k <= 0 {
		m.k = len(input.RetrievalIDs)
	}

	// 计算 DCG
	dcg := m.calculateDCG(input.RetrievalIDs, input.RetrievalGT, m.k)

	// 计算 IDCG (理想情况下的DCG)
	idcg := m.calculateIDCG(input.RetrievalGT, m.k)

	if idcg == 0 {
		return 0.0
	}

	return dcg / idcg
}

// calculateDCG 计算折扣累积增益
func (m *NDCGMetric) calculateDCG(retrievalIDs []int, groundTruth [][]int, k int) float64 {
	dcg := 0.0
	gtSet := m.buildGTSet(groundTruth)

	for i := 0; i < min(k, len(retrievalIDs)); i++ {
		relevance := 0.0
		if gtSet[retrievalIDs[i]] {
			relevance = 1.0
		}
		dcg += relevance / math.Log2(float64(i+2))
	}

	return dcg
}

// calculateIDCG 计算理想DCG
func (m *NDCGMetric) calculateIDCG(groundTruth [][]int, k int) float64 {
	gtSet := m.buildGTSet(groundTruth)
	idcg := 0.0

	// 假设所有相关文档都排在前面
	count := 0
	for i := 0; i < k; i++ {
		if count >= len(gtSet) {
			break
		}
		idcg += 1.0 / math.Log2(float64(i+2))
		count++
	}

	return idcg
}

// buildGTSet 构建 Ground Truth 集合
func (m *NDCGMetric) buildGTSet(groundTruth [][]int) map[int]bool {
	gtSet := make(map[int]bool)
	for _, gtList := range groundTruth {
		for _, id := range gtList {
			gtSet[id] = true
		}
	}
	return gtSet
}

// Name 返回指标名称
func (m *NDCGMetric) Name() string {
	return "ndcg"
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
