//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/Edward-Jackie/lambo/backend/internal/handler"
	"github.com/Edward-Jackie/lambo/backend/internal/repository"
	"github.com/Edward-Jackie/lambo/backend/internal/server"
	"github.com/Edward-Jackie/lambo/backend/internal/service"
	"github.com/Edward-Jackie/lambo/backend/pkg/log"
	"github.com/gin-gonic/gin"
	"github.com/google/wire"
	"github.com/spf13/viper"
)

var ServerSet = wire.NewSet(server.NewServerHTTP)

var RepositorySet = wire.NewSet(
	repository.NewDb,
	repository.NewRepository,
	repository.NewCategoryRepository,
	repository.NewAuthRepository,
)

var ServiceSet = wire.NewSet(
	service.NewService,
	service.NewCategoryService,
	service.NewAuthService,
)

var HandlerSet = wire.NewSet(
	handler.NewHandler,
	handler.NewCategoryHandler,
	handler.NewAuthHandler,
)

func NewWire(*viper.Viper, *log.Logger) (*gin.Engine, func(), error) {
	panic(wire.Build(
		ServerSet,
		RepositorySet,
		ServiceSet,
		HandlerSet,
	))
}
