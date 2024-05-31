// Copyright 2024 Roy(徐武) <ixw1991@126.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/imxw/miniokr.

package okr

import (
	"errors"

	"github.com/gin-gonic/gin"

	ctrlV1 "github.com/imxw/miniokr/internal/miniokr/controller/v1"
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
		Title: req.Title,
		// Owner:  username,
		Date:   req.Date,
		Weight: req.Weight,
	}

	if req.UserId != "" {

		// 校验权限
		userID, ok := c.MustGet(known.XUserIDKey).(string)
		if !ok {
			core.WriteResponse(c, errno.InternalServerError, nil)
			return
		}
		roles, ok := c.MustGet(known.UserRolesKey).([]string)
		if !ok {
			core.WriteResponse(c, errno.InternalServerError, nil)
			return
		}
		if !ctrlV1.CheckPermission(c, userID, roles, req.UserId, ctrl.us) {
			return
		}
		// 查询目标用户名
		user, err := ctrl.us.GetUserByID(c, req.UserId)
		if err != nil {
			core.WriteResponse(c, errno.ErrUserNotFound, nil)
			return
		}
		objective.Owner = user.Name
	} else {
		objective.Owner = username
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
		Title: req.Title,
		// Owner:  username,
		Date:   req.Date,
		Weight: req.Weight,
		ID:     trimIDPrefix(uriParam.ID),
	}

	if req.UserId != "" {
		// 校验权限
		userID, ok := c.MustGet(known.XUserIDKey).(string)
		if !ok {
			core.WriteResponse(c, errno.InternalServerError, nil)
			return
		}
		roles, ok := c.MustGet(known.UserRolesKey).([]string)
		if !ok {
			core.WriteResponse(c, errno.InternalServerError, nil)
			return
		}
		if !ctrlV1.CheckPermission(c, userID, roles, req.UserId, ctrl.us) {
			return
		}
		// 查询目标用户名
		user, err := ctrl.us.GetUserByID(c, req.UserId)
		if err != nil {
			core.WriteResponse(c, errno.ErrUserNotFound, nil)
			return
		}
		objective.Owner = user.Name
	} else {
		objective.Owner = username
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

	if req.UserId != "" {
		// 校验权限
		userID, ok := c.MustGet(known.XUserIDKey).(string)
		if !ok {
			core.WriteResponse(c, errno.InternalServerError, nil)
			return
		}
		roles, ok := c.MustGet(known.UserRolesKey).([]string)
		if !ok {
			core.WriteResponse(c, errno.InternalServerError, nil)
			return
		}
		if !ctrlV1.CheckPermission(c, userID, roles, req.UserId, ctrl.us) {
			return
		}

	}

	// TODO: 更加精细地检查权限，如O或KR的owner是不是自己或自己的下属

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
