package service

import "github.com/Edward-Jackie/lambo/backend/pkg/log"

// Service 是业务区 service 的公共依赖；系统区有同名结构，两区互不引用。
type Service struct {
	logger *log.Logger
}

func NewService(logger *log.Logger) *Service {
	return &Service{
		logger: logger,
	}
}
