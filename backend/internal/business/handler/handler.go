package handler

import (
	"github.com/Edward-Jackie/lambo/backend/pkg/log"
)

// Handler 是业务区 handler 的公共依赖；系统区有同名结构，两区互不引用。
type Handler struct {
	logger *log.Logger
}

func NewHandler(logger *log.Logger) *Handler {
	return &Handler{
		logger: logger,
	}
}
