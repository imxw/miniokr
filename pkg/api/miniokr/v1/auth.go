// Copyright 2024 Roy(徐武) <ixw1991@126.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/xwlearn/miniokr.

package v1

// AuthRequest 指定了 `POST /api/v1/auth/dingtalk` 接口的请求参数
type AuthRequest struct {
	AuthCode string `json:"authCode" valid:"required"`
}

// AuthResponse 指定了 `POST /api/v1/auth/dingtalk` 接口的返回参数
type AuthResponse struct {
	Token string `json:"token"`
}
