// Copyright 2024 Roy(徐武) <ixw1991@126.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/xwlearn/miniokr.

package okr

import (
	"github.com/xwlearn/miniokr/internal/miniokr/services/field"
	"github.com/xwlearn/miniokr/internal/miniokr/services/okr"
)

type Controller struct {
	fs field.Service
	os okr.Service
}

func New(fs field.Service, os okr.Service) *Controller {
	return &Controller{fs: fs, os: os}
}
