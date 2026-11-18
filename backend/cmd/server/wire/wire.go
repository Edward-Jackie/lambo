//go:build wireinject
// +build wireinject

package wire

import (
	bizhandler "github.com/Edward-Jackie/lambo/backend/internal/business/handler"
	bizrepository "github.com/Edward-Jackie/lambo/backend/internal/business/repository"
	bizservice "github.com/Edward-Jackie/lambo/backend/internal/business/service"
	syshandler "github.com/Edward-Jackie/lambo/backend/internal/system/handler"
	sysrepository "github.com/Edward-Jackie/lambo/backend/internal/system/repository"
	sysservice "github.com/Edward-Jackie/lambo/backend/internal/system/service"
	"github.com/Edward-Jackie/lambo/backend/pkg/log"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/spf13/viper"
)

var ServerSet = wire.NewSet(provideEngine)

var RepositorySet = wire.NewSet(
	provideDb,
	provideMenuCatalog,
	sysrepository.NewRepository,
	sysrepository.NewAuthRepository,
	sysrepository.NewSystemUserRepository,
	sysrepository.NewSystemRoleRepository,
	sysrepository.NewSystemMenuRepository,
	bizrepository.NewRepository,
	bizrepository.NewCategoryRepository,
)

var ServiceSet = wire.NewSet(
	sysservice.NewService,
	sysservice.NewAuthService,
	sysservice.NewSystemUserService,
	sysservice.NewSystemRoleService,
	sysservice.NewSystemMenuService,
	bizservice.NewService,
	bizservice.NewCategoryService,
)

var HandlerSet = wire.NewSet(
	syshandler.NewHandler,
	syshandler.NewAuthHandler,
	syshandler.NewSystemUserHandler,
	syshandler.NewSystemRoleHandler,
	syshandler.NewSystemMenuHandler,
	bizhandler.NewHandler,
	bizhandler.NewCategoryHandler,
)

func NewWire(*viper.Viper, *log.Logger) (*gin.Engine, func(), error) {
	panic(wire.Build(
		ServerSet,
		RepositorySet,
		ServiceSet,
		HandlerSet,
	))
}
