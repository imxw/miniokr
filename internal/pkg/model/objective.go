// Copyright 2024 Roy(徐武) <ixw1991@126.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/xwlearn/miniokr.

package model

type Objective struct {
	ID               string   `json:"id"`
	Title            string   `json:"title"`
	Owner            string   `json:"owner"`
	Date             string   `json:"date"`
	Weight           int      `json:"weight"`
	KrsIds           []string `json:"krsIds"`
	CreatedTime      int64    `json:"createdTime"`
	LastModifiedTime int64    `json:"lastModifiedTime"`
}
