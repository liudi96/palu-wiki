package testutils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

// HTTPTestHelper HTTP测试助手
type HTTPTestHelper struct {
	Router *gin.Engine
}

// NewHTTPTestHelper 创建HTTP测试助手
func NewHTTPTestHelper() *HTTPTestHelper {
	gin.SetMode(gin.TestMode)
	return &HTTPTestHelper{
		Router: gin.New(),
	}
}

// POST 执行POST请求测试
func (h *HTTPTestHelper) POST(path string, body interface{}, headers ...map[string]string) *httptest.ResponseRecorder {
	var req *http.Request
	
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		req = httptest.NewRequest("POST", path, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest("POST", path, nil)
	}

	// 添加自定义头部
	for _, headerMap := range headers {
		for key, value := range headerMap {
			req.Header.Set(key, value)
		}
	}

	recorder := httptest.NewRecorder()
	h.Router.ServeHTTP(recorder, req)
	return recorder
}

// GET 执行GET请求测试
func (h *HTTPTestHelper) GET(path string, headers ...map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("GET", path, nil)

	// 添加自定义头部
	for _, headerMap := range headers {
		for key, value := range headerMap {
			req.Header.Set(key, value)
		}
	}

	recorder := httptest.NewRecorder()
	h.Router.ServeHTTP(recorder, req)
	return recorder
}

// PUT 执行PUT请求测试
func (h *HTTPTestHelper) PUT(path string, body interface{}, headers ...map[string]string) *httptest.ResponseRecorder {
	var req *http.Request
	
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		req = httptest.NewRequest("PUT", path, bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest("PUT", path, nil)
	}

	// 添加自定义头部
	for _, headerMap := range headers {
		for key, value := range headerMap {
			req.Header.Set(key, value)
		}
	}

	recorder := httptest.NewRecorder()
	h.Router.ServeHTTP(recorder, req)
	return recorder
}

// DELETE 执行DELETE请求测试
func (h *HTTPTestHelper) DELETE(path string, headers ...map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest("DELETE", path, nil)

	// 添加自定义头部
	for _, headerMap := range headers {
		for key, value := range headerMap {
			req.Header.Set(key, value)
		}
	}

	recorder := httptest.NewRecorder()
	h.Router.ServeHTTP(recorder, req)
	return recorder
}

// WithAuth 添加JWT认证头部
func WithAuth(token string) map[string]string {
	return map[string]string{
		"Authorization": "Bearer " + token,
	}
}

// AssertJSONResponse 断言JSON响应
func AssertJSONResponse(t require.TestingT, recorder *httptest.ResponseRecorder, expectedCode int, expectedBody interface{}) {
	require.Equal(t, expectedCode, recorder.Code)
	require.Equal(t, "application/json; charset=utf-8", recorder.Header().Get("Content-Type"))
	
	if expectedBody != nil {
		expectedJSON, _ := json.Marshal(expectedBody)
		require.JSONEq(t, string(expectedJSON), recorder.Body.String())
	}
}

// ParseJSONResponse 解析JSON响应
func ParseJSONResponse(recorder *httptest.ResponseRecorder, target interface{}) error {
	return json.Unmarshal(recorder.Body.Bytes(), target)
}