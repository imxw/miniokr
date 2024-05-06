// Copyright 2024 Roy(徐武) <ixw1991@126.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/imxw/miniokr.

package okr

import (
	"context"

	"github.com/imxw/miniokr/internal/pkg/model"
)

type Service interface {
	ListObjectivesByOwner(ctx context.Context, username string, sortBy string, orderBy string) ([]model.Objective, error)
	ListKeyResultsByOwner(ctx context.Context, username string, sortBy string, orderBy string) ([]model.KeyResult, error)
	CreateObjective(context.Context, model.Objective) (string, error)
	UpdateObjective(context.Context, model.Objective) error
	DeleteObjectiveByID(context.Context, string, []string) error
	CreateKeyResult(context.Context, model.KeyResult) (string, error)
	UpdateKeyResult(context.Context, model.KeyResult) error
	DeleteKeyResultByID(context.Context, string) error
}
