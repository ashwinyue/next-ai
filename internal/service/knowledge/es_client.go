package knowledge

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
)

// ESSearcher Elasticsearch 搜索接口，用于抽象 ES 客户端
type ESSearcher interface {
	// Search 执行搜索请求并返回响应
	DoSearch(ctx context.Context, index string, queryJSON []byte) (*ESResponse, error)
}

// ESResponse Elasticsearch 搜索响应
type ESResponse struct {
	IsError bool
	Body    io.ReadCloser
	String  string
}

// realESSearcher 真实 ES 客户端的适配器
type realESSearcher struct {
	doSearch func(ctx context.Context, index string, body io.Reader) (*ESResponseImpl, error)
}

// ESResponseImpl ES 响应实现（兼容 go-elasticsearch）
type ESResponseImpl struct {
	isError bool
	body    io.ReadCloser
	str     string
}

func (r *ESResponseImpl) IsError() bool       { return r.isError }
func (r *ESResponseImpl) Body() io.ReadCloser { return r.body }
func (r *ESResponseImpl) String() string      { return r.str }

// newRealESSearcher 创建真实 ES 搜索器
func newRealESSearcher(doSearch func(ctx context.Context, index string, body io.Reader) (*ESResponseImpl, error)) ESSearcher {
	return &realESSearcher{doSearch: doSearch}
}

func (r *realESSearcher) DoSearch(ctx context.Context, index string, queryJSON []byte) (*ESResponse, error) {
	resp, err := r.doSearch(ctx, index, bytes.NewReader(queryJSON))
	if err != nil {
		return nil, err
	}
	return &ESResponse{
		IsError: resp.IsError(),
		Body:    resp.Body(),
		String:  resp.String(),
	}, nil
}

// mockESSearcher 用于测试的 mock ES 搜索器
type mockESSearcher struct {
	searchFunc func(ctx context.Context, index string, queryJSON []byte) (*ESResponse, error)
}

// newMockESSearcher 创建 mock ES 搜索器
func newMockESSearcher(searchFunc func(ctx context.Context, index string, queryJSON []byte) (*ESResponse, error)) ESSearcher {
	return &mockESSearcher{searchFunc: searchFunc}
}

func (m *mockESSearcher) DoSearch(ctx context.Context, index string, queryJSON []byte) (*ESResponse, error) {
	if m.searchFunc != nil {
		return m.searchFunc(ctx, index, queryJSON)
	}
	// 默认返回空结果
	return &ESResponse{
		IsError: false,
		Body:    io.NopCloser(bytes.NewReader([]byte("{}"))),
		String:  "{}",
	}, nil
}

// helper: 创建空搜索响应
func createEmptySearchResponse() []byte {
	resp := map[string]interface{}{
		"hits": map[string]interface{}{
			"hits": []interface{}{},
		},
	}
	data, _ := json.Marshal(resp)
	return data
}

// helper: 创建带有结果的搜索响应
func createSearchResponse(results []map[string]interface{}) []byte {
	hits := make([]map[string]interface{}, len(results))
	for i, r := range results {
		hits[i] = map[string]interface{}{
			"_id":    r["id"],
			"_score": r["score"],
			"_source": r["source"],
		}
	}
	resp := map[string]interface{}{
		"hits": map[string]interface{}{
			"hits": hits,
		},
	}
	data, _ := json.Marshal(resp)
	return data
}

// helper: 创建 ES 错误响应
func createErrorResponse(message string) []byte {
	resp := map[string]interface{}{
		"error": map[string]interface{}{
			"reason": message,
		},
	}
	data, _ := json.Marshal(resp)
	return data
}
