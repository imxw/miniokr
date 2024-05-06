// Copyright 2024 Roy(徐武) <ixw1991@126.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/xwlearn/miniokr.

//go:build integration
// +build integration

package token

import (
	"context"
	"os"
	"testing"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	"github.com/stretchr/testify/assert"
)

func TestGetNewTokenIntegration(t *testing.T) {
	appID := os.Getenv("APP_ID")
	appSecret := os.Getenv("APP_SECRET")
	if appID == "" || appSecret == "" {
		t.Skip("APP_ID or APP_SECRET is not set. Skipping integration test.")
	}
	client := lark.NewClient(appID, appSecret)

	// 创建服务
	lts := &LarkTokenService{Client: client}

	// 调用方法
	resp, err := lts.GetNewToken(context.Background(), appID, appSecret)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Data.TenantAccessToken)
	t.Logf("token is: %s", resp.Data.TenantAccessToken)
}

func TestGetNewTokenIntegration_Failure(t *testing.T) {
	// 使用无效的凭据
	appID := "invalid_app_id"
	appSecret := "invalid_app_secret"
	client := lark.NewClient(appID, appSecret)

	lts := LarkTokenService{Client: client}

	_, err := lts.GetNewToken(context.Background(), appID, appSecret)
	assert.Error(t, err) // 确认应该有错误返回

	// 检查错误消息确保包含预期的内容
	if err != nil {
		assert.Contains(t, err.Error(), "服务端返回错误")
	}
}
