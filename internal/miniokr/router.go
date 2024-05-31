// Copyright 2024 Roy(徐武) <ixw1991@126.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/imxw/miniokr.

package miniokr

import (
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"

	"github.com/imxw/miniokr/internal/pkg/core"
	"github.com/imxw/miniokr/internal/pkg/log"
	"github.com/imxw/miniokr/internal/pkg/middleware"
)

// installRouters 安装 miniokr 接口路由.
func installRouters(g *gin.Engine, sc *ServiceContainer, msc *middleware.MiddlewareServiceContainer) error {

	g.Static("/static", "./frontend/build/static")
	// 将首页路由到React的入口HTML
	g.GET("/", func(c *gin.Context) {
		c.File("./frontend/build/index.html")
	})
	g.NoRoute(func(c *gin.Context) {
		c.File("./frontend/build/index.html")
	})

	// 注册 404 Handler.
	// g.NoRoute(func(c *gin.Context) {
	// 	core.WriteResponse(c, errno.ErrPageNotFound, nil)
	// })

	// 注册 /healthz handler.
	g.GET("/healthz", func(c *gin.Context) {
		log.C(c).Infow("Healthz function called")

		core.WriteResponse(c, nil, map[string]string{"status": "ok"})
	})

	// 注册 pprof 路由
	pprof.Register(g)

	// 创建v1路由分组
	v1 := g.Group("/api/v1")
	v1.POST("/auth/dingtalk", sc.AuthController.Auth)
	v1.Use(middleware.Authn(msc))
	v1.GET("/fields", sc.FieldController.List)
	v1.GET("/okrs", sc.OkrController.ListOkrByUsernameAndMonths)
	v1.POST("/okrs", sc.OkrController.ListOkrByUsernameAndMonths)
	v1.POST("/objectives", sc.OkrController.CreateObjective)
	v1.PUT("/objectives/:id", sc.OkrController.UpdateObjective)
	v1.DELETE("/objectives/:id", sc.OkrController.DeleteObjective)
	v1.POST("/keyresults", sc.OkrController.CreateKeyResult)
	v1.PUT("/keyresults/:id", sc.OkrController.UpdateKeyResult)
	v1.DELETE("/keyresults/:id", sc.OkrController.DeleteKeyResult)
	// v1.GET("/users", sc.UserController.GetUser)
	v1.GET("/users/:id/departments/tree", sc.UserController.GetUserDepartmentsTree)
	v1.GET("/user/departments/tree", sc.UserController.GetDepartmentsTree)
	v1.GET("/me", sc.UserController.GetCurrentUser)

	return nil
}
