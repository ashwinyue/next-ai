// Package testutil 提供测试辅助工具
package testutil

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"
)

// HTTPRoundTripper 重写 HTTP 请求到测试服务器
// 用于将真实 API 请求重定向到 mock 服务器
type HTTPRoundTripper struct {
	base *url.URL          // 测试服务器 URL
	next http.RoundTripper // 下一个 Transport
}

// NewHTTPRoundTripper 创建 HTTP 请求重定向器
func NewHTTPRoundTripper(baseURL string) *HTTPRoundTripper {
	u, _ := url.Parse(baseURL)
	return &HTTPRoundTripper{
		base: u,
		next: http.DefaultTransport,
	}
}

// NewHTTPRoundTripperWithTransport 创建带有自定义 Transport 的重定向器
func NewHTTPRoundTripperWithTransport(baseURL string, next http.RoundTripper) *HTTPRoundTripper {
	u, _ := url.Parse(baseURL)
	if next == nil {
		next = http.DefaultTransport
	}
	return &HTTPRoundTripper{
		base: u,
		next: next,
	}
}

// RoundTrip 实现 http.RoundTripper 接口
func (t *HTTPRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	// 只有当请求匹配目标主机时才重写
	if t.shouldRewrite(req) {
		cloned := *req
		u := *req.URL
		u.Scheme = t.base.Scheme
		u.Host = t.base.Host
		cloned.URL = &u
		req = &cloned
	}
	return t.next.RoundTrip(req)
}

// shouldRewrite 判断是否应该重写请求
// 子类可以覆盖此方法来实现自定义的重写逻辑
func (t *HTTPRoundTripper) shouldRewrite(req *http.Request) bool {
	return true
}

// NewTestClient 创建测试用 HTTP 客户端
// 自动将请求重定向到测试服务器
func NewTestClient(ts *httptest.Server) *http.Client {
	return NewTestClientWithTimeout(ts, 5*time.Second)
}

// NewTestClientWithTimeout 创建带超时的测试 HTTP 客户端
func NewTestClientWithTimeout(ts *httptest.Server, timeout time.Duration) *http.Client {
	u, _ := url.Parse(ts.URL)
	return &http.Client{
		Timeout: timeout,
		Transport: &HTTPRoundTripper{
			base: u,
			next: http.DefaultTransport,
		},
	}
}

// NewTestClientFromURL 从 URL 创建测试 HTTP 客户端
func NewTestClientFromURL(baseURL string) *http.Client {
	u, err := url.Parse(baseURL)
	if err != nil {
		u = &url.URL{}
	}
	return &http.Client{
		Timeout: 5 * time.Second,
		Transport: &HTTPRoundTripper{
			base: u,
			next: http.DefaultTransport,
		},
	}
}

// HostSpecificRoundTripper 只重写特定主机的请求
type HostSpecificRoundTripper struct {
	*HTTPRoundTripper
	hosts []string // 需要重写的主机列表
}

// NewHostSpecificRoundTripper 创建主机特定的重定向器
func NewHostSpecificRoundTripper(baseURL string, hosts []string) *HostSpecificRoundTripper {
	return &HostSpecificRoundTripper{
		HTTPRoundTripper: NewHTTPRoundTripper(baseURL),
		hosts:            hosts,
	}
}

// shouldRewrite 只重写指定主机的请求
func (t *HostSpecificRoundTripper) shouldRewrite(req *http.Request) bool {
	for _, host := range t.hosts {
		if req.URL.Host == host {
			return true
		}
	}
	return false
}

func parseURL(s string) *url.URL {
	u, _ := url.Parse(s)
	return u
}
