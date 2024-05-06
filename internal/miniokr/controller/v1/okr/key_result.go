// Copyright 2024 Roy(徐武) <ixw1991@126.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/xwlearn/miniokr.

package okr

import (
	"errors"

	"github.com/gin-gonic/gin"

	"github.com/xwlearn/miniokr/internal/pkg/core"
	"github.com/xwlearn/miniokr/internal/pkg/errno"
	"github.com/xwlearn/miniokr/internal/pkg/known"
	"github.com/xwlearn/miniokr/internal/pkg/log"
	"github.com/xwlearn/miniokr/internal/pkg/model"
	v1 "github.com/xwlearn/miniokr/pkg/api/miniokr/v1"
)

// 新增 KeyResult
func (ctrl *Controller) CreateKeyResult(c *gin.Context) {
	log.C(c).Infow("CreateKeyResult function Called")

	// 获取参数
	var req v1.CreateOrUpdateKeyResult
	if err := c.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)
		return
	}

	// 校验用户
	username, ok := c.MustGet(known.XUsernameKey).(string)
	if !ok {
		core.WriteResponse(c, errors.New("无法获取用户名"), nil)
		return
	}
	date, err := standardizeMonthFormat(req.Date)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}

	// 转换
	kr := model.KeyResult{
		Title:       req.Title,
		Owner:       username,
		Date:        date,
		Weight:      req.Weight,
		Completed:   req.Completed,
		SelfRating:  req.SelfRating,
		Reason:      req.Reason,
		ObjectiveID: trimIDPrefix(req.ObjectiveID),
		Criteria:    req.Criteria,
	}

	id, err := ctrl.os.CreateKeyResult(c, kr)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}
	core.WriteResponse(c, nil, v1.CreateKeyResultResponse{ID: "kr-" + id})
}

// 更新 KeyResult
func (ctrl *Controller) UpdateKeyResult(c *gin.Context) {
	log.C(c).Infow("UpdateKeyResult function Called")

	// 获取参数
	var req v1.CreateOrUpdateKeyResult
	if err := c.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)
		return
	}

	var uriParam v1.UpdateRecordReq
	if err := c.ShouldBindUri(&uriParam); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)
		return
	}

	// 校验用户
	username, ok := c.MustGet(known.XUsernameKey).(string)
	if !ok {
		core.WriteResponse(c, errors.New("无法获取用户名"), nil)
		return
	}

	date, err := standardizeMonthFormat(req.Date)
	if err != nil {
		core.WriteResponse(c, nil, nil)
		return
	}

	// 转换
	kr := model.KeyResult{
		Title:       req.Title,
		Owner:       username,
		Date:        date,
		Weight:      req.Weight,
		Completed:   req.Completed,
		SelfRating:  req.SelfRating,
		Reason:      req.Reason,
		ObjectiveID: trimIDPrefix(req.ObjectiveID),
		Criteria:    req.Criteria,
		ID:          trimIDPrefix(uriParam.ID),
	}

	if err := ctrl.os.UpdateKeyResult(c, kr); err != nil {

		core.WriteResponse(c, err, nil)
		return
	}
	core.WriteResponse(c, nil, nil)
}

// 删除 KeyResult
func (ctrl *Controller) DeleteKeyResult(c *gin.Context) {
	log.C(c).Infow("DeleteKeyResult function Called")
	var req v1.DeleteRecordReq
	if err := c.ShouldBindUri(&req); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)
		return
	}

	if err := ctrl.os.DeleteKeyResultByID(c, trimIDPrefix(req.ID)); err != nil {
		core.WriteResponse(c, err, nil)
		return
	}
	core.WriteResponse(c, nil, nil)
}
