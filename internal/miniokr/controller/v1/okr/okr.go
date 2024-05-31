// Copyright 2024 Roy(徐武) <ixw1991@126.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/imxw/miniokr.

package okr

import (
	"github.com/imxw/miniokr/internal/miniokr/services/field"
	"github.com/imxw/miniokr/internal/miniokr/services/okr"
	"github.com/imxw/miniokr/internal/miniokr/services/user"
)

type Controller struct {
	fs field.Service
	os okr.Service
	us user.Service
}

func New(fs field.Service, os okr.Service, us user.Service) *Controller {
	return &Controller{fs: fs, os: os, us: us}
}
