package wire

import (
	"context"

	platformdatabase "github.com/Edward-Jackie/lambo/backend/internal/platform/database"
	sysrepository "github.com/Edward-Jackie/lambo/backend/internal/system/repository"
	"github.com/Edward-Jackie/lambo/backend/pkg/log"
	"github.com/spf13/viper"
)

// ApplyMenuSeed 使用与 HTTP 服务相同的模块装配，直接访问目标库。
func ApplyMenuSeed(ctx context.Context, conf *viper.Viper, logger *log.Logger, spec sysrepository.MenuSeedDefinition, confirmNew bool) (string, error) {
	db, cleanup, err := platformdatabase.NewDb(conf, logger)
	if err != nil {
		return "", err
	}
	defer cleanup()
	return sysrepository.NewRepository(logger, db, provideMenuCatalog()).ApplyMenuSeed(ctx, spec, confirmNew)
}
