package server

import (
	"fmt"
	"github.com/Edward-Jackie/lambo/backend/internal/model"

	"github.com/Edward-Jackie/lambo/backend/internal/handler"
	"github.com/Edward-Jackie/lambo/backend/internal/middleware"
	"github.com/Edward-Jackie/lambo/backend/pkg/helper/resp"
	"github.com/Edward-Jackie/lambo/backend/pkg/log"
	"github.com/gin-gonic/gin"
)

func NewServerHTTP(
	logger *log.Logger,
	authHandler *handler.AuthHandler,
	categoryHandler *handler.CategoryHandler,
	userHandler *handler.SystemUserHandler,
	roleHandler *handler.SystemRoleHandler,
	menuHandler *handler.SystemMenuHandler,
) (*gin.Engine, error) {
	if logger == nil || authHandler == nil || categoryHandler == nil || userHandler == nil || roleHandler == nil || menuHandler == nil {
		return nil, fmt.Errorf("HTTP 服务缺少认证或分类依赖")
	}
	r := newBaseHTTPServer()
	authHandler.RegisterRoutes(r)
	if err := categoryHandler.RegisterRoutes(r.Group("", authHandler.RequireSession()), authHandler.Authorize); err != nil {
		return nil, err
	}
	group := r.Group("", authHandler.RequireSession())
	for _, register := range []func() error{func() error { return userHandler.RegisterRoutes(group, authHandler.Authorize) }, func() error { return roleHandler.RegisterRoutes(group, authHandler.Authorize) }, func() error { return menuHandler.RegisterRoutes(group, authHandler.Authorize) }} {
		if err := register(); err != nil {
			return nil, err
		}
	}
	group.POST("/api/system/users/:id/password", authHandler.RequireCode(model.PermissionUserPassword), userHandler.ResetPassword)
	return r, nil
}

func newBaseHTTPServer() *gin.Engine {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(
		middleware.CORSMiddleware(),
	)
	// 仅表示 HTTP 进程可响应，不代替数据库和业务就绪检查。
	r.GET("/healthz", func(ctx *gin.Context) {
		resp.HandleSuccess(ctx, map[string]interface{}{"status": "ok"})
	})
	r.GET("/", func(ctx *gin.Context) {
		resp.HandleSuccess(ctx, map[string]interface{}{
			"say": "Hi Lambo!",
		})
	})
	return r
}
