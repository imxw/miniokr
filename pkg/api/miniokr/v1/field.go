// Copyright 2024 Roy(徐武) <ixw1991@126.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/imxw/miniokr.

package v1

type ObjectiveField struct {
	Title        string `json:"title" field:"fldp1iQXFv"`
	Owner        string `json:"owner" field:"fld8vFDsXz"`
	Date         string `json:"date" field:"fldc36J6LW"`
	Weight       string `json:"weight" field:"fldgooDzO7"`
	KeyResultIDs string `json:"keyResultIds" field:"fld1dLUo8S"`
}

type KeyResultField struct {
	Title       string `json:"title" field:"fldrxgL9LV"`
	Owner       string `json:"owner" field:"fldjEZmY3S"`
	Date        string `json:"date" field:"fldnMxJlKj"`
	Weight      string `json:"weight" field:"fld5HXzwmN"`
	Completed   string `json:"completed" field:"fldG2nSDTZ"`
	SelfRating  string `json:"selfRating" field:"fldvjvtxRr"`
	Criteria    string `json:"criteria" field:"fldPGxpg2b"`
	ObjectiveID string `json:"objectiveId" field:"fldW8TFesB"`
	Reason      string `json:"Reason" field:"fld6OsYad8"`
}

// FieldMappingsResponse 指定了 `POST /api/v1/fields` 接口的返回参数.
type FieldMappingsResponse struct {
	Objective ObjectiveField `json:"objective"`
	KeyResult KeyResultField `json:"keyResult"`
}
