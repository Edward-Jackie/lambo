package repository

import (
	"github.com/Edward-Jackie/lambo/backend/pkg/log"
	"gorm.io/gorm"
)

// Repository 是业务区 repository 的公共依赖；系统区有同名结构，两区互不引用。
type Repository struct {
	db     *gorm.DB
	logger *log.Logger
}

func NewRepository(logger *log.Logger, db *gorm.DB) *Repository {
	return &Repository{
		db:     db,
		logger: logger,
	}
}
