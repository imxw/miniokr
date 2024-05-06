// Copyright 2024 Roy(徐武) <ixw1991@126.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/imxw/miniokr.

package auth

import (
	"context"
	"fmt"

	"github.com/zhaoyunxing92/dingtalk/v2"
	"github.com/zhaoyunxing92/dingtalk/v2/request"

	"github.com/imxw/miniokr/internal/pkg/errno"
	"github.com/imxw/miniokr/internal/pkg/log"
	"github.com/imxw/miniokr/pkg/token"
)

type Config struct {
	ClientId     string
	ClientSecret string
}

type DingTalkAuthService struct {
	cfg    Config
	client *dingtalk.DingTalk
}

func NewDingTalkAuthService(cfg Config) (*DingTalkAuthService, error) {
	// 使用 dingtalk 包初始化客户端
	client, err := dingtalk.NewClient(cfg.ClientId, cfg.ClientSecret)
	if err != nil {
		log.Errorw("[NewUserFetchService] Failed to create DingTalk client", "err", err)

		return nil, err
	}
	return &DingTalkAuthService{
		cfg:    cfg,
		client: client,
	}, nil
}

// Fetch 获取钉钉组织内用户姓名
func (d *DingTalkAuthService) Fetch(ctx context.Context, code string) (string, error) {

	userinfo, err := d.client.GetUserInfoByCode(code)
	if err != nil {
		log.Errorw("[d.client.GetUserInfoByCode] Failed to fetch userinfo", "err", err)
		return "", err
	}

	userDetailRequest := request.NewUserDetail(userinfo.UserInfo.UserId).Build()
	userDetail, err := d.client.GetUserDetail(userDetailRequest)
	if err != nil {
		log.Errorw("[d.client.GetUserDetail] Failed to fetch user detail", "userID", userinfo.UserInfo.UserId, "err", err)
		return "", err
	}

	return userDetail.Name, nil
}

func (d *DingTalkAuthService) IssueToken(ctx context.Context, username string) (string, error) {

	t, err := token.Sign(username)
	if err != nil {
		return "", errno.ErrSignToken
	}
	return t, nil
}

// 一个用于检查和获取映射值的辅助函数
func getMappingValue(mapping map[string]string, key string) (string, error) {
	value, ok := mapping[key]
	if !ok {
		return "", fmt.Errorf("required field '%s' is missing from the mapping", key)
	}
	return value, nil
}
