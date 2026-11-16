package repository

import (
	"github.com/Edward-Jackie/lambo/backend/pkg/log"
	"gorm.io/gorm"
)

type Repository struct {
	db *gorm.DB
	//rdb    *redis.Client
	logger *log.Logger
}

func NewRepository(logger *log.Logger, db *gorm.DB) *Repository {
	return &Repository{
		db: db,
		//rdb:    rdb,
		logger: logger,
	}
}
