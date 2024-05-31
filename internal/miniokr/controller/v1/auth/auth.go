// Copyright 2024 Roy(徐武) <ixw1991@126.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/imxw/miniokr.

package auth

import (
	"github.com/gin-gonic/gin"

	"github.com/imxw/miniokr/internal/miniokr/services/auth"
	"github.com/imxw/miniokr/internal/pkg/core"
	"github.com/imxw/miniokr/internal/pkg/errno"
	"github.com/imxw/miniokr/internal/pkg/log"
	v1 "github.com/imxw/miniokr/pkg/api/miniokr/v1"
)

type Controller struct {
	a auth.Authenticator
}

func New(a auth.Authenticator) *Controller {
	return &Controller{a: a}
}

func (ctrl *Controller) Auth(c *gin.Context) {
	log.C(c).Infow("Auth function called")

	var r v1.AuthRequest
	if err := c.ShouldBindJSON(&r); err != nil {
		log.C(c).Errorw("请求参数绑定失败", "authCode", r.AuthCode)
		core.WriteResponse(c, errno.ErrBind, nil)
		return
	}
	userid, username, err := ctrl.a.Fetch(c, r.AuthCode)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}

	token, err := ctrl.a.IssueToken(c, userid, username)
	log.C(c).Infow("获取用户token", "token", token, "username", username)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}
	core.WriteResponse(c, nil, v1.AuthResponse{Token: token})
}
