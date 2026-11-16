package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Edward-Jackie/lambo/backend/internal/model"
	"github.com/Edward-Jackie/lambo/backend/pkg/log"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	ormlog "gorm.io/gorm/logger"
)

func NewDb(conf *viper.Viper, logger *log.Logger) (*gorm.DB, func(), error) {
	dsn := strings.TrimSpace(conf.GetString("data.postgres.dsn"))
	if dsn == "" {
		return nil, nil, fmt.Errorf("请配置 data.postgres.dsn 或 LAMBO_POSTGRES_DSN")
	}
	dbLogger := ormlog.New(zap.NewStdLog(logger.Logger), ormlog.Config{
		LogLevel: ormlog.Warn, SlowThreshold: 200 * time.Millisecond,
		IgnoreRecordNotFoundError: true, ParameterizedQueries: true,
	})
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: dbLogger})
	if err != nil {
		return nil, nil, fmt.Errorf("连接 PostgreSQL 失败：%w", err)
	}
	pool, err := db.DB()
	if err != nil {
		return nil, nil, err
	}
	cleanup := func() { _ = pool.Close() }
	pool.SetMaxOpenConns(10)
	pool.SetMaxIdleConns(5)
	pool.SetConnMaxLifetime(30 * time.Minute)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := pool.PingContext(ctx); err != nil {
		cleanup()
		return nil, nil, fmt.Errorf("PostgreSQL 就绪检查失败：%w", err)
	}
	if conf.GetBool("data.auto_migrate") {
		if err := db.WithContext(ctx).AutoMigrate(&model.Category{}, &model.Account{}, &model.Session{}); err != nil {
			cleanup()
			return nil, nil, fmt.Errorf("初始化业务表失败：%w", err)
		}
	}
	logger.Info("PostgreSQL 已连接", zap.Bool("autoMigrate", conf.GetBool("data.auto_migrate")))
	return db, cleanup, nil
}
