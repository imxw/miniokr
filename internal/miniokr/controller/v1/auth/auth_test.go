// Copyright 2024 Roy(徐武) <ixw1991@126.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/imxw/miniokr.

package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	v1 "github.com/imxw/miniokr/pkg/api/miniokr/v1"
)

type MockAuthenticator struct {
	mock.Mock
}

func (m *MockAuthenticator) Fetch(ctx context.Context, code string) (string, error) {
	args := m.Called(ctx, code)
	return args.String(0), args.Error(1)
}

func (m *MockAuthenticator) IssueToken(ctx context.Context, username string) (string, error) {
	args := m.Called(ctx, username)
	return args.String(0), args.Error(1)
}

func TestAuth(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建模拟的 Authenticator
	mockAuth := new(MockAuthenticator)
	mockAuth.On("Fetch", mock.AnythingOfType("*gin.Context"), "valid-code").Return("testuser", nil)
	mockAuth.On("IssueToken", mock.AnythingOfType("*gin.Context"), "testuser").Return("valid-token", nil)

	// 创建 Controller 实例
	controller := New(mockAuth)

	// 设置 Gin 的路由
	r := gin.Default()
	r.POST("/auth", controller.Auth)

	// 创建一个请求体
	reqBody := v1.AuthRequest{AuthCode: "valid-code"}
	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/auth", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	// 创建响应记录器
	w := httptest.NewRecorder()

	// 调用 handler
	r.ServeHTTP(w, req)

	// 测试状态码和响应体
	assert.Equal(t, http.StatusOK, w.Code)
	var resp v1.AuthResponse
	json.Unmarshal(w.Body.Bytes(), &resp)
	assert.Equal(t, "valid-token", resp.Token)

	// 验证预期的方法被调用
	mockAuth.AssertExpectations(t)
}

// TestAuthWithInvalidAuthCode 测试无效的 AuthCode
func TestAuthWithInvalidAuthCode(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockAuth := new(MockAuthenticator)
	mockAuth.On("Fetch", mock.AnythingOfType("*gin.Context"), "invalid-code").Return("", errors.New("invalid auth code"))

	controller := New(mockAuth)
	router := gin.Default()
	router.POST("/auth", controller.Auth)

	reqBody := v1.AuthRequest{AuthCode: "invalid-code"}
	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/auth", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code) // 或者根据您的错误处理逻辑选择正确的状态码
	mockAuth.AssertExpectations(t)
}

// TestAuthWhenIssueTokenFails 测试 IssueToken 返回错误
func TestAuthWhenIssueTokenFails(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockAuth := new(MockAuthenticator)
	mockAuth.On("Fetch", mock.AnythingOfType("*gin.Context"), "valid-code").Return("testuser", nil)
	mockAuth.On("IssueToken", mock.AnythingOfType("*gin.Context"), "testuser").Return("", errors.New("token issue failed"))

	controller := New(mockAuth)
	router := gin.Default()
	router.POST("/auth", controller.Auth)

	reqBody := v1.AuthRequest{AuthCode: "valid-code"}
	bodyBytes, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", "/auth", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code) // 或者根据您的错误处理逻辑选择正确的状态码
	mockAuth.AssertExpectations(t)
}

func TestAuthWithInvalidRequestBody(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// 创建 mock 的 Authenticator
	mockAuth := new(MockAuthenticator)
	controller := New(mockAuth)

	router := gin.Default()
	router.POST("/auth", controller.Auth)

	// 创建一个请求体，故意构造为无效的 JSON 格式，以触发绑定错误
	invalidJSON := `{"authCode":123}` // authCode 应该是一个字符串，这里故意写成数字
	req, _ := http.NewRequest("POST", "/auth", bytes.NewBufferString(invalidJSON))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// 检查返回的状态码是否为错误的状态码
	assert.Equal(t, http.StatusBadRequest, w.Code) // 确认是否返回了错误的状态码，根据您的实际情况可能是 400 或其他

	// 解析响应体以验证错误详情
	var errResponse struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &errResponse); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// 校验错误代码和消息是否与 ErrBind 相匹配
	assert.Equal(t, "InvalidParameter.BindError", errResponse.Code)
	assert.Equal(t, "Error occurred while binding the request body to the struct.", errResponse.Message)
}
