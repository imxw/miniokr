// Copyright 2024 Roy(徐武) <ixw1991@126.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/imxw/miniokr.

package v1

// ListOkrResquest 指定了 `GET|POST /api/v1/okrs` 接口的请求参数.
type ListOkrRequest struct {
	Months  []string `json:"months" form:"months" binding:"omitempty,dive,monthYearFormat"`
	SortBy  string   `json:"sortBy" form:"sortBy" binding:"omitempty"`
	UserID  string   `json:"userId" form:"userId" binding:"omitempty"`
	OrderBy string   `json:"orderBy" form:"orderBy" binding:"omitempty"`
}

// ListOkrResponse 指定了 `GET|POST /api/v1/okrs` 接口的返回参数.
type ListOkrResponse struct {
	Okrs map[string][]Objective `json:"okrs"`
}

type Objective struct {
	ID               string      `json:"id"`
	Title            string      `json:"title"`
	Owner            string      `json:"owner"`
	Date             string      `json:"date"`
	Weight           int         `json:"weight"`
	KeyResults       []KeyResult `json:"keyResults"`
	CreatedTime      int64       `json:"createdTime"`
	LastModifiedTime int64       `json:"lastModifiedTime"`
}

type KeyResult struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	Weight           int    `json:"weight"`
	Owner            string `json:"owner"`
	Date             string `json:"date"`
	Completed        string `json:"completed"`
	SelfRating       *int   `json:"selfRating"`
	Reason           string `json:"reason"`
	ObjectiveID      string `json:"objectiveID"`
	Criteria         string `json:"criteria"`
	Leader           string `json:"leader"`
	LeaderRating     *int   `json:"leaderRating"`
	Department       string `json:"department"`
	CreatedTime      int64  `json:"createdTime"`
	LastModifiedTime int64  `json:"lastModifiedTime"`
}

type CreateOrUpdateObjective struct {
	Title  string `json:"title" binding:"required"`
	Date   string `json:"date" binding:"required,monthYearFormat"`
	Weight int    `json:"weight" binding:"omitempty,min=0,max=100"`
	UserId string `json:"userId,omitempty"`
}

type CreateObjectiveResponse struct {
	ID string `json:"id"`
}
type CreateKeyResultResponse struct {
	ID string `json:"id"`
}

type CreateOrUpdateKeyResult struct {
	Title        string `json:"title" binding:"required"`
	Weight       int    `json:"weight" binding:"required,min=1,max=100"`
	Date         string `json:"date" binding:"required,monthYearFormat"`
	Completed    string `json:"completed" binding:"required,oneof=未开始 已完成 未完成"`
	SelfRating   *int   `json:"selfRating" binding:"omitempty,min=0,max=120"`
	Reason       string `json:"reason" binding:"omitempty"`
	LeaderRating *int   `json:"leaderRating,omitempty" binding:"omitempty"`
	UserId       string `json:"userId" binding:"omitempty"`
	ObjectiveID  string `json:"objectiveID" binding:"omitempty"`
	Criteria     string `json:"criteria" binding:"omitempty"`
}

type DeleteObjectiveReq struct {
	DeleteRecordReq
	KeyResultIDs []string `json:"keyResultIds"` // 从 JSON 请求体绑定关键结果的 IDs
	UserId       string   `json:"userId,omitempty"`
}

type DeleteRecordReq struct {
	ID string `uri:"id" binding:"required"`
}

type UpdateRecordReq struct {
	ID string `uri:"id" binding:"required"`
}
