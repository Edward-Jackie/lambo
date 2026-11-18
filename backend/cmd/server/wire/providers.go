package wire

import (
	"context"
	"fmt"
	"time"

	bizhandler "github.com/Edward-Jackie/lambo/backend/internal/business/handler"
	bizmodel "github.com/Edward-Jackie/lambo/backend/internal/business/model"
	platformdatabase "github.com/Edward-Jackie/lambo/backend/internal/platform/database"
	"github.com/Edward-Jackie/lambo/backend/internal/platform/httpserver"
	syshandler "github.com/Edward-Jackie/lambo/backend/internal/system/handler"
	sysmodel "github.com/Edward-Jackie/lambo/backend/internal/system/model"
	sysrepository "github.com/Edward-Jackie/lambo/backend/internal/system/repository"
	"github.com/Edward-Jackie/lambo/backend/pkg/log"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"gorm.io/gorm"
)

// provideEngine 在装配层挂载两区路由，平台包不感知具体模块。
func provideEngine(
	logger *log.Logger,
	authHandler *syshandler.AuthHandler,
	categoryHandler *bizhandler.CategoryHandler,
	userHandler *syshandler.SystemUserHandler,
	roleHandler *syshandler.SystemRoleHandler,
	menuHandler *syshandler.SystemMenuHandler,
) (*gin.Engine, error) {
	if logger == nil || authHandler == nil || categoryHandler == nil || userHandler == nil || roleHandler == nil || menuHandler == nil {
		return nil, fmt.Errorf("HTTP 服务缺少认证或分类依赖")
	}
	r := httpserver.NewEngine()
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
	group.POST("/api/system/users/:id/password", authHandler.RequireCode(sysmodel.PermissionUserPassword), userHandler.ResetPassword)
	return r, nil
}

// provideMenuCatalog 汇集已装配模块的页面和权限定义。
func provideMenuCatalog() sysmodel.MenuCatalog {
	return sysmodel.NewMenuCatalog(bizhandler.CategoryMenuRegistration())
}

// provideDb 建立连接，并按两区表清单执行自动迁移；迁移清单只出现在装配层。
func provideDb(conf *viper.Viper, logger *log.Logger) (*gorm.DB, func(), error) {
	db, cleanup, err := platformdatabase.NewDb(conf, logger)
	if err != nil {
		return nil, nil, err
	}
	if !conf.GetBool("data.auto_migrate") {
		return db, cleanup, nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := sysrepository.CheckUsernameUpgrade(ctx, db); err != nil {
		cleanup()
		return nil, nil, err
	}
	if err := platformdatabase.Migrate(ctx, db,
		&sysmodel.Session{}, &sysmodel.SystemUser{}, &sysmodel.SystemRole{}, &sysmodel.SystemMenu{},
		&sysmodel.AccountRole{}, &sysmodel.RoleMenu{}, &sysmodel.IAMMigration{},
		&bizmodel.Category{},
	); err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("初始化数据表失败：%w", err)
	}
	return db, cleanup, nil
}
