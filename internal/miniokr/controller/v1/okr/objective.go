// Copyright 2024 Roy(徐武) <ixw1991@126.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/imxw/miniokr.

package okr

import (
	"errors"

	"github.com/gin-gonic/gin"

	"github.com/imxw/miniokr/internal/pkg/core"
	"github.com/imxw/miniokr/internal/pkg/errno"
	"github.com/imxw/miniokr/internal/pkg/known"
	"github.com/imxw/miniokr/internal/pkg/log"
	"github.com/imxw/miniokr/internal/pkg/model"
	v1 "github.com/imxw/miniokr/pkg/api/miniokr/v1"
)

// 新增 Objective
func (ctrl *Controller) CreateObjective(c *gin.Context) {

	log.C(c).Infow("CreateObjective function Called")

	// 获取参数
	var req v1.CreateOrUpdateObjective
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

	// 转换
	objective := model.Objective{
		Title:  req.Title,
		Owner:  username,
		Date:   req.Date,
		Weight: req.Weight,
	}

	id, err := ctrl.os.CreateObjective(c, objective)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}

	core.WriteResponse(c, nil, v1.CreateObjectiveResponse{ID: "o-" + id})
}

// 更新 Objective
func (ctrl *Controller) UpdateObjective(c *gin.Context) {

	log.C(c).Infow("UpdateObjective function Called")

	// 获取参数
	var req v1.CreateOrUpdateObjective
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

	// 转换
	objective := model.Objective{
		Title:  req.Title,
		Owner:  username,
		Date:   req.Date,
		Weight: req.Weight,
		ID:     trimIDPrefix(uriParam.ID),
	}

	err := ctrl.os.UpdateObjective(c, objective)
	if err != nil {
		core.WriteResponse(c, err, nil)
		return
	}
	core.WriteResponse(c, nil, nil)

}

// 删除 Objective
func (ctrl *Controller) DeleteObjective(c *gin.Context) {

	log.C(c).Infow("DeleteObjective function Called")
	var req v1.DeleteObjectiveReq
	if err := c.ShouldBindUri(&req); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)
		return
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)
		return
	}

	trimmedIDs := make([]string, len(req.KeyResultIDs))
	for i, id := range req.KeyResultIDs {
		trimmedIDs[i] = trimIDPrefix(id)
	}

	if err := ctrl.os.DeleteObjectiveByID(c, trimIDPrefix(req.ID), trimmedIDs); err != nil {
		core.WriteResponse(c, err, nil)
		return
	}
	core.WriteResponse(c, nil, nil)
}
