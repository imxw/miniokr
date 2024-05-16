// Copyright 2024 Roy(徐武) <ixw1991@126.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/imxw/miniokr.

package model

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
