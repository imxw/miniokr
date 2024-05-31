// Copyright 2024 Roy(徐武) <ixw1991@126.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/imxw/miniokr.

package okr

import (
	"errors"
	"fmt"
	"strings"

	"github.com/gin-gonic/gin"

	ctrlV1 "github.com/imxw/miniokr/internal/miniokr/controller/v1"
	"github.com/imxw/miniokr/internal/pkg/bitable"
	"github.com/imxw/miniokr/internal/pkg/core"
	"github.com/imxw/miniokr/internal/pkg/errno"
	"github.com/imxw/miniokr/internal/pkg/known"
	"github.com/imxw/miniokr/internal/pkg/log"
	"github.com/imxw/miniokr/internal/pkg/model"
	v1 "github.com/imxw/miniokr/pkg/api/miniokr/v1"
)

func (ctrl *Controller) ListOkrByUsernameAndMonths(c *gin.Context) {
	log.C(c).Infow("okr ListOkrByUsernameAndMonths function called")

	var req v1.ListOkrRequest
	if err := c.ShouldBind(&req); err != nil {
		core.WriteResponse(c, errno.ErrBind, nil)
		return
	}

	if !isValidSortBy(req.SortBy) || !isValidOrderBy(req.OrderBy) {
		core.WriteResponse(c, errors.New("invalid sortBy or orderBy parameter"), nil)
		return
	}

	var owner string

	// username, ok := c.MustGet(known.XUsernameKey).(string)
	// if !ok {
	// 	core.WriteResponse(c, errno.InternalServerError, nil)
	// 	return
	// }
	userID, ok := c.MustGet(known.XUserIDKey).(string)
	if !ok {
		core.WriteResponse(c, errno.InternalServerError, nil)
		return
	}

	// legalMonths, err := ctrl.fs.GetValidDates(c)
	// if err != nil {
	// 	core.WriteResponse(c, err, nil)
	// 	return
	// }

	//months, err := ctrl.validateMonths(req.Months)
	//if err != nil {
	//	core.WriteResponse(c, err, nil)
	//	return
	//}

	if req.UserID != "" {
		// 请求他人资源需要鉴权
		roles, ok := c.MustGet(known.UserRolesKey).([]string)
		if !ok {
			core.WriteResponse(c, errno.InternalServerError, nil)
			return
		}
		if !ctrlV1.CheckPermission(c, userID, roles, req.UserID, ctrl.us) {
			return
		}
		owner = req.UserID
	} else {
		owner = userID
	}

	objData, krData, err := ctrl.fetchData(c, owner, req.SortBy, req.OrderBy)
	if err != nil {

		if errors.Is(err, bitable.ErrInvalidUser) {
			emptyRes := v1.ListOkrResponse{
				Okrs: make(map[string][]v1.Objective),
			}

			for _, month := range req.Months {
				emptyRes.Okrs[month] = []v1.Objective{}
			}

			core.WriteResponse(c, nil, emptyRes)
			return
		}

		core.WriteResponse(c, err, nil)
		return
	}

	okrResponse := ctrl.constructResponse(req.Months, objData, krData)
	core.WriteResponse(c, nil, okrResponse)
}

func isValidSortBy(sortBy string) bool {
	validSortBys := map[string]bool{"createtime": true, "updatetime": true, "title": true}
	sortBy = strings.ToLower(sortBy)
	return sortBy == "" || validSortBys[sortBy]
}

func isValidOrderBy(orderBy string) bool {
	validOrderBys := map[string]bool{"asc": true, "desc": true}
	orderBy = strings.ToLower(orderBy)
	return orderBy == "" || validOrderBys[orderBy]
}

func (ctrl *Controller) prepareUserAndMonths(c *gin.Context) (string, []string, error) {
	username, ok := c.MustGet(known.XUsernameKey).(string)
	if !ok {
		return "", nil, errors.New("无法获取用户名")
	}

	legalUsers, err := ctrl.fs.GetValidUsers(c)
	if err != nil {
		return "", nil, err
	}

	if !ctrlV1.Contains(legalUsers, username) {
		return "", nil, errors.New("用户不合法")
	}

	legalMonths, err := ctrl.fs.GetValidDates(c)
	if err != nil {
		return "", nil, err
	}

	return username, legalMonths, nil
}

func (ctrl *Controller) validateMonths(reqMonths []string) ([]string, error) {
	var months []string
	for _, v := range reqMonths {
		date, err := standardizeMonthFormat(v)
		if err != nil {
			return nil, errors.New("日期格式不对")
		}
		months = append(months, date)
	}

	recentMonths := formatRecentMonths()
	if len(months) == 0 {
		return recentMonths, nil
	} else {
		months = intersectStrings(recentMonths, months)
	}

	if len(months) == 0 {
		return nil, fmt.Errorf("只允许获取数据的月份: %v", recentMonths)
	}

	return months, nil
}

func (ctrl *Controller) fetchData(c *gin.Context, userid string, sortBy string, orderBy string) ([]model.Objective, []model.KeyResult, error) {

	user, err := ctrl.us.GetUserByID(c, userid)
	if err != nil {
		return nil, nil, err
	}

	objData, err := ctrl.os.ListObjectivesByOwner(c, user.Name, sortBy, orderBy)
	if err != nil {
		return nil, nil, err
	}
	krData, err := ctrl.os.ListKeyResultsByOwner(c, user.Name, sortBy, orderBy)

	if err != nil {
		return nil, nil, err
	}
	return objData, krData, nil
}

func (ctrl *Controller) constructResponse(months []string, objData []model.Objective, krData []model.KeyResult) v1.ListOkrResponse {
	krMap := make(map[string][]v1.KeyResult)
	for _, kr := range krData {
		v1Kr := convertToV1KeyResult(kr)
		krMap[kr.ObjectiveID] = append(krMap[kr.ObjectiveID], v1Kr)
	}

	groupedObjectives := make(map[string][]v1.Objective)
	// 初始化月份分组以确保每个月都有条目，即使是空的
	for _, month := range months {
		groupedObjectives[month] = []v1.Objective{}
	}

	// 创建一个查找集合，用于检查日期是否在提供的月份中
	monthSet := make(map[string]bool)
	for _, m := range months {
		monthSet[m] = true
	}

	// dataFound := false
	for _, obj := range objData {
		if _, found := monthSet[obj.Date]; found { // 只处理指定月份中的目标
			// dataFound = true
			krList, exists := krMap[obj.ID]
			if !exists {
				krList = []v1.KeyResult{} // 确保有一个空切片而不是nil
			}
			okr := v1.Objective{
				ID:               obj.ID,
				Title:            obj.Title,
				Owner:            obj.Owner,
				Date:             obj.Date,
				Weight:           obj.Weight,
				KeyResults:       krList,
				CreatedTime:      obj.CreatedTime,
				LastModifiedTime: obj.LastModifiedTime,
			}
			groupedObjectives[obj.Date] = append(groupedObjectives[obj.Date], okr)
		}
	}

	// if !dataFound {
	// 	return v1.ListOkrResponse{}
	// }

	return v1.ListOkrResponse{Okrs: groupedObjectives}
}

func convertToV1KeyResult(kr model.KeyResult) v1.KeyResult {
	return v1.KeyResult{
		ID:               kr.ID,
		Title:            kr.Title,
		Weight:           kr.Weight,
		Owner:            kr.Owner,
		Date:             kr.Date,
		Completed:        kr.Completed,
		SelfRating:       kr.SelfRating,
		Reason:           kr.Reason,
		ObjectiveID:      kr.ObjectiveID,
		Leader:           kr.Leader,
		LeaderRating:     kr.LeaderRating,
		Department:       kr.Department,
		Criteria:         kr.Criteria,
		CreatedTime:      kr.CreatedTime,
		LastModifiedTime: kr.LastModifiedTime,
	}
}
