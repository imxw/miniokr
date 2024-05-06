// Copyright 2024 Roy(徐武) <ixw1991@126.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/imxw/miniokr.

package token

import (
	"context"
	"encoding/json"
	"fmt"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkauth "github.com/larksuite/oapi-sdk-go/v3/service/auth/v3"

	"github.com/imxw/miniokr/internal/pkg/log"
)

type LarkTokenService struct {
	Client *lark.Client
}

var _ Service = (*LarkTokenService)(nil)

func (lts *LarkTokenService) GetNewToken(ctx context.Context, appID string, appSecret string) (*Response, error) {
	reqBody := larkauth.NewInternalTenantAccessTokenReqBodyBuilder().
		AppId(appID).
		AppSecret(appSecret).
		Build()

	req := larkauth.NewInternalTenantAccessTokenReqBuilder().Body(reqBody).Build()

	resp, err := lts.Client.Auth.TenantAccessToken.Internal(ctx, req)
	if err != nil {
		log.C(ctx).Errorw("获取Token失败", err)
		return nil, fmt.Errorf("获取Token失败: %v", err)
	}
	if !resp.Success() {
		log.C(ctx).Errorw("获取token失败", "errMsg", resp.Msg)
		return nil, fmt.Errorf("服务端返回错误: code=%d, msg=%s, requestId=%s", resp.Code, resp.Msg, resp.RequestId())
	}

	var tokenResp Response
	if err := json.Unmarshal(resp.RawBody, &tokenResp); err != nil {
		return nil, fmt.Errorf("解析Token JSON出错: %v", err)
	}
	return &tokenResp, nil
}
