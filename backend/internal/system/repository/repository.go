package repository

import (
	"github.com/Edward-Jackie/lambo/backend/internal/system/model"
	"github.com/Edward-Jackie/lambo/backend/pkg/log"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
	//rdb    *redis.Client
	logger      *log.Logger
	menuCatalog model.MenuCatalog
}

func NewRepository(logger *log.Logger, db *gorm.DB, catalog model.MenuCatalog) *Repository {
	return &Repository{
		db: db,
		//rdb:    rdb,
		logger:      logger,
		menuCatalog: catalog,
	}
}
