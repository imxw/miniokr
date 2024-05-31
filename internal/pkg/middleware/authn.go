// Copyright 2024 Roy(徐武) <ixw1991@126.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/imxw/miniokr.

package middleware

import (
	"github.com/gin-gonic/gin"

	"github.com/imxw/miniokr/internal/miniokr/services/user"
	"github.com/imxw/miniokr/internal/pkg/core"
	"github.com/imxw/miniokr/internal/pkg/errno"
	"github.com/imxw/miniokr/internal/pkg/known"
	"github.com/imxw/miniokr/pkg/token"
)

type MiddlewareServiceContainer struct {
	UserService user.Service
}

// Authn 是认证中间件，用来从 gin.Context 中提取 token 并验证 token 是否合法，
// 如果合法则将 token 中的 sub 作为<用户名>存放在 gin.Context 的 XUsernameKey 键中.
func Authn(services *MiddlewareServiceContainer) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 解析 JWT Token
		claims, err := token.ParseRequest(c)
		if err != nil {
			core.WriteResponse(c, errno.ErrTokenInvalid, nil)
			c.Abort()

			return
		}

		// 提取并设置用户名和用户ID
		username, usernameOk := claims[known.XUsernameKey].(string)
		userID, userIDOk := claims[known.XUserIDKey].(string)

		// 检查提取是否成功
		if !usernameOk || !userIDOk {
			core.WriteResponse(c, errno.ErrTokenInvalid, nil)
			c.Abort()
			return
		}

		roles, err := services.UserService.GetUserRolesByID(c, userID)
		if err != nil {
			core.WriteResponse(c, errno.InternalServerError, nil)
			c.Abort()
			return
		}

		c.Set(known.XUsernameKey, username)
		c.Set(known.XUserIDKey, userID)
		c.Set(known.UserRolesKey, roles)
		c.Next()
	}
}
