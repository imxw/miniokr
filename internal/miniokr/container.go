// Copyright 2024 Roy(徐武) <ixw1991@126.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/xwlearn/miniokr.

package miniokr

import (
	"github.com/xwlearn/miniokr/internal/miniokr/controller/v1/auth"
	"github.com/xwlearn/miniokr/internal/miniokr/controller/v1/field"
	"github.com/xwlearn/miniokr/internal/miniokr/controller/v1/okr"
)

type ServiceContainer struct {
	AuthController  *auth.Controller
	FieldController *field.Controller
	OkrController   *okr.Controller
}