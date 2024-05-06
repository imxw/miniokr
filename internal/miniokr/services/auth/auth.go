// Copyright 2024 Roy(徐武) <ixw1991@126.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/imxw/miniokr.

package auth

import (
	"context"
)

type Authenticator interface {
	TokenIssuer
	UserFetcher
}

type TokenIssuer interface {
	IssueToken(ctx context.Context, username string) (string, error)
}

type UserFetcher interface {
	Fetch(ctx context.Context, code string) (string, error)
}
