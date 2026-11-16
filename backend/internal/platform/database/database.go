package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/Edward-Jackie/lambo/backend/pkg/log"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	ormlog "gorm.io/gorm/logger"
)

// NewDb 只负责连接、连接池与就绪检查；迁移表清单由装配层传入，底座不依赖 system 与 business。
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
	logger.Info("PostgreSQL 已连接", zap.Bool("autoMigrate", conf.GetBool("data.auto_migrate")))
	return db, cleanup, nil
}

// Migrate 由装配层给出表清单，按传入顺序初始化表结构。
func Migrate(ctx context.Context, db *gorm.DB, models ...interface{}) error {
	return db.WithContext(ctx).AutoMigrate(models...)
}
