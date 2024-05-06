// Copyright 2024 Roy(徐武) <ixw1991@126.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/xwlearn/miniokr.

package field

import (
	"context"

	v1 "github.com/xwlearn/miniokr/pkg/api/miniokr/v1"
)

type Service interface {
	GetFieldDefinitions(ctx context.Context) (v1.FieldMappingsResponse, error)
	GetValidDates(ctx context.Context) ([]string, error)
	GetValidUsers(ctx context.Context) ([]string, error)
}
