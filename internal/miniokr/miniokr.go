// Copyright 2024 Roy(徐武) <ixw1991@126.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file. The original repo for
// this file is https://github.com/imxw/miniokr.

package miniokr

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	lark "github.com/larksuite/oapi-sdk-go/v3"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	ac "github.com/imxw/miniokr/internal/miniokr/controller/v1/auth"
	fc "github.com/imxw/miniokr/internal/miniokr/controller/v1/field"
	oc "github.com/imxw/miniokr/internal/miniokr/controller/v1/okr"
	syncv1 "github.com/imxw/miniokr/internal/miniokr/controller/v1/sync"
	uc "github.com/imxw/miniokr/internal/miniokr/controller/v1/user"
	"github.com/imxw/miniokr/internal/miniokr/services/auth"
	fs "github.com/imxw/miniokr/internal/miniokr/services/field"
	okrs "github.com/imxw/miniokr/internal/miniokr/services/okr"
	users "github.com/imxw/miniokr/internal/miniokr/services/user"
	"github.com/imxw/miniokr/internal/miniokr/store"
	repo "github.com/imxw/miniokr/internal/miniokr/store"
	"github.com/imxw/miniokr/internal/pkg/bitable"
	"github.com/imxw/miniokr/internal/pkg/bitable/field"
	larkToken "github.com/imxw/miniokr/internal/pkg/bitable/token"
	"github.com/imxw/miniokr/internal/pkg/log"
	"github.com/imxw/miniokr/internal/pkg/middleware"
	mw "github.com/imxw/miniokr/internal/pkg/middleware"
	"github.com/imxw/miniokr/pkg/token"
	"github.com/imxw/miniokr/pkg/version/verflag"
)

var cfgFile string

// NewMiniOkrCommand 创建一个 *cobra.Command 对象. 之后，可以使用 Command 对象的 Execute 方法来启动应用程序.
func NewMiniOkrCommand() *cobra.Command {
	cmd := &cobra.Command{
		// 指定命令的名字，该名字会出现在帮助信息中
		Use: "miniokr",
		// 命令的简短描述
		Short: "A good Go practical project",
		// 命令的详细描述
		Long: `A good Go practical project, used to create user with basic information.

Find more miniokr information at:
	https://github.com/imxw/miniokr#readme`,

		// 命令出错时，不打印帮助信息。不需要打印帮助信息，设置为 true 可以保持命令出错时一眼就能看到错误信息
		SilenceUsage: true,
		// 指定调用 cmd.Execute() 时，执行的 Run 函数，函数执行失败会返回错误信息
		RunE: func(cmd *cobra.Command, args []string) error {
			// 如果 `--version=true`，则打印版本并退出
			verflag.PrintAndExitIfRequested()

			// 初始化日志
			log.Init(logOptions())
			defer log.Sync() // Sync 将缓存中的日志刷新到磁盘文件中

			return run()
		},
		// 这里设置命令运行时，不需要指定命令行参数
		Args: func(cmd *cobra.Command, args []string) error {
			for _, arg := range args {
				if len(arg) > 0 {
					return fmt.Errorf("%q does not take any arguments, got %q", cmd.CommandPath(), args)
				}
			}

			return nil
		},
	}

	// 以下设置，使得 initConfig 函数在每个命令运行时都会被调用以读取配置
	cobra.OnInitialize(initConfig)

	// 在这里您将定义标志和配置设置。

	// Cobra 支持持久性标志(PersistentFlag)，该标志可用于它所分配的命令以及该命令下的每个子命令
	cmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "The path to the miniokr configuration file. Empty string for no configuration file.")

	// Cobra 也支持本地标志，本地标志只能在其所绑定的命令上使用
	cmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")

	// 添加 --version 标志
	verflag.AddFlags(cmd.PersistentFlags())

	return cmd
}

// run 函数是实际的业务代码入口函数.
func run() error {

	// 初始化 store 层
	db, err := initStore()
	if err != nil {
		return err
	}

	// 初始化钉钉客户端
	dingClient, err := initDingTalkClient()
	if err != nil {
		log.Fatalw("Failed to initialize DingTalk client", "error", err)
		return err
	}

	// 初始化同步服务
	syncService, err := initSyncService(db, dingClient)
	if err != nil {
		log.Fatalw("Failed to initialize sync service", "error", err)
		return err
	}

	// 立即运行一次同步任务
	log.Infow("Running initial sync task...")
	go func() {
		ctx := context.Background()
		if err := syncService.SyncDepartmentsAndUsers(ctx); err != nil {
			log.Fatalw("Initial sync task failed", "error", err)
		} else {
			log.Infow("Initial sync task succeeded")
			// 在同步后初始化角色和用户角色
			if err := store.S.InitRoles(); err != nil {
				log.Fatalw("Failed to initialize roles", "error", err)
			}
			if err := store.S.InitUserRoles(); err != nil {
				log.Fatalw("Failed to initialize user roles", "error", err)
			}
		}
	}()

	// 启动定时任务
	syncController := syncv1.NewSyncController(syncService)
	go syncController.StartCronJob()

	// 初始化钉钉服务
	as, err := auth.NewDingTalkAuthService(auth.Config{
		ClientId:     viper.GetString("dingtalk.client-id"),
		ClientSecret: viper.GetString("dingtalk.client-secret"),
	})
	if err != nil {
		log.Fatalw("Failed to initialize DingTalk AuthService", "error", err)
		return err
	}

	ctx := context.Background()
	fsAppID := viper.GetString("feishu.app-id")
	fsAppSecret := viper.GetString("feishu.app-secret")
	fsAppToken := viper.GetString("feishu.app-token")
	client := lark.NewClient(fsAppID, fsAppSecret)

	feishuToken := &larkToken.LarkTokenService{Client: client}

	store := larkToken.NewTokenStorage()
	clock := larkToken.NewRealClock()
	fm := larkToken.NewManager(feishuToken, store, clock, fsAppID, fsAppSecret)
	err = fm.Initialize(ctx)
	if err != nil {
		log.Fatalw("Failed to Initialize", "error", err)
	}
	fieldManager := field.NewManager(client, fsAppToken, fm)
	oTableID := viper.GetString("feishu.o-table-id")
	krTableID := viper.GetString("feishu.kr-table-id")

	// 初始化缓存
	fieldManager.LoadOrRefreshFieldMapping(ctx, oTableID)
	fieldManager.LoadOrRefreshFieldMapping(ctx, krTableID)

	// 初始化Field服务
	fieldService, err := fs.NewFeishufieldService(oTableID, krTableID, fieldManager)
	if err != nil {
		log.Fatalw("Failed to Initialize Feishufieldservice", "error", err)
		return err
	}

	rm := bitable.NewRecordManager(client, fsAppToken, fm)
	// 初始化Okr服务
	okrService, err := okrs.NewFeishuOkrService(oTableID, krTableID, fieldManager, rm)

	// 初始化用户服务
	userService := users.NewUserService(repo.S.Users())

	container := &ServiceContainer{
		AuthController:  ac.New(as),
		FieldController: fc.New(fieldService),
		OkrController:   oc.New(fieldService, okrService, userService),
		UserController:  uc.New(userService),
	}

	msc := &middleware.MiddlewareServiceContainer{
		UserService: userService,
	}

	// 初始化飞书服务

	// 设置 token 包的签发密钥，用于 token 包 token 的签发和解析
	token.Init(viper.GetString("jwt.secret"), viper.GetDuration("jwt.expiration"))

	// 设置 Gin 模式
	gin.SetMode(viper.GetString("runmode"))

	// 创建 Gin 引擎
	g := gin.New()

	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		err := v.RegisterValidation("monthYearFormat", monthYearFormat)
		if err != nil {
			panic(err)
		}
	}
	// gin.Recovery() 中间件，用来捕获任何 panic，并恢复
	mws := []gin.HandlerFunc{gin.Recovery(), mw.NoCache, mw.Cors, mw.Secure, mw.RequestID()}

	g.Use(mws...)

	if err := installRouters(g, container, msc); err != nil {
		return err
	}

	// 创建并运行 HTTP 服务器
	httpsrv := startInsecureServer(g)

	// 创建并运行 HTTPS 服务器
	// httpssrv := startSecureServer(g)

	// 创建并运行 GRPC 服务器
	// grpcsrv := startGRPCServer()

	// 等待中断信号优雅地关闭服务器（10 秒超时)。
	quit := make(chan os.Signal, 1)
	// kill 默认会发送 syscall.SIGTERM 信号
	// kill -2 发送 syscall.SIGINT 信号，我们常用的 CTRL + C 就是触发系统 SIGINT 信号
	// kill -9 发送 syscall.SIGKILL 信号，但是不能被捕获，所以不需要添加它
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM) // 此处不会阻塞
	<-quit                                               // 阻塞在此，当接收到上述两种信号时才会往下执行
	log.Infow("Shutting down server ...")

	// 创建 ctx 用于通知服务器 goroutine, 它有 10 秒时间完成当前正在处理的请求
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 10 秒内优雅关闭服务（将未处理完的请求处理完再关闭服务），超过 10 秒就超时退出
	if err := httpsrv.Shutdown(ctx); err != nil {
		log.Errorw("Insecure Server forced to shutdown", "err", err)
		return err
	}
	// if err := httpssrv.Shutdown(ctx); err != nil {
	// 	log.Errorw("Secure Server forced to shutdown", "err", err)
	// 	return err
	// }

	// grpcsrv.GracefulStop()

	log.Infow("Server exiting")

	return nil
}

// startInsecureServer 创建并运行 HTTP 服务器.
func startInsecureServer(g *gin.Engine) *http.Server {
	// 创建 HTTP Server 实例
	httpsrv := &http.Server{Addr: viper.GetString("addr"), Handler: g}

	// 运行 HTTP 服务器。在 goroutine 中启动服务器，它不会阻止下面的正常关闭处理流程
	// 打印一条日志，用来提示 HTTP 服务已经起来，方便排障
	log.Infow("Start to listening the incoming requests on http address", "addr", viper.GetString("addr"))
	go func() {
		if err := httpsrv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalw(err.Error())
		}
	}()

	return httpsrv
}

// startSecureServer 创建并运行 HTTPS 服务器.
func startSecureServer(g *gin.Engine) *http.Server {
	// 创建 HTTPS Server 实例
	httpssrv := &http.Server{Addr: viper.GetString("tls.addr"), Handler: g}

	// 运行 HTTPS 服务器。在 goroutine 中启动服务器，它不会阻止下面的正常关闭处理流程
	// 打印一条日志，用来提示 HTTPS 服务已经起来，方便排障
	log.Infow("Start to listening the incoming requests on https address", "addr", viper.GetString("tls.addr"))
	cert, key := viper.GetString("tls.cert"), viper.GetString("tls.key")
	if cert != "" && key != "" {
		go func() {
			if err := httpssrv.ListenAndServeTLS(cert, key); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatalw(err.Error())
			}
		}()
	}

	return httpssrv
}

// startGRPCServer 创建并运行 GRPC 服务器.
// func startGRPCServer() *grpc.Server {
// 	lis, err := net.Listen("tcp", viper.GetString("grpc.addr"))
// 	if err != nil {
// 		log.Fatalw("Failed to listen", "err", err)
// 	}

// 	// 创建 GRPC Server 实例
// 	grpcsrv := grpc.NewServer()
// 	pb.RegisterMiniBlogServer(grpcsrv, user.New(store.S, nil))

// 	// 运行 GRPC 服务器。在 goroutine 中启动服务器，它不会阻止下面的正常关闭处理流程
// 	// 打印一条日志，用来提示 GRPC 服务已经起来，方便排障
// 	log.Infow("Start to listening the incoming requests on grpc address", "addr", viper.GetString("grpc.addr"))
// 	go func() {
// 		if err := grpcsrv.Serve(lis); err != nil {
// 			log.Fatalw(err.Error())
// 		}
// 	}()

// 	return grpcsrv
// }
