// Copyright 2024 Roy(徐武) <ixw1991@126.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/xwlearn/miniokr.

package field

import (
	"github.com/gin-gonic/gin"

	"github.com/xwlearn/miniokr/internal/miniokr/services/field"
	"github.com/xwlearn/miniokr/internal/pkg/core"
	"github.com/xwlearn/miniokr/internal/pkg/log"
)

type Controller struct {
	fs field.Service
}

func New(fs field.Service) *Controller {
	return &Controller{fs: fs}
}

func (ctrl *Controller) List(c *gin.Context) {
	log.C(c).Infow("List function called")

	resp, err := ctrl.fs.GetFieldDefinitions(c)
	if err != nil {
		// TODO: 优化错误处理
		core.WriteResponse(c, err, nil)
		return
	}
	core.WriteResponse(c, nil, resp)
}
