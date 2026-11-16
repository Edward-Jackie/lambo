// 由 Lambo 生成，可按业务修改；再次生成遇到差异会停止。
package service

import (
	"context"

	"github.com/Edward-Jackie/lambo/backend/internal/business/model"
	"github.com/Edward-Jackie/lambo/backend/internal/business/repository"
)

type CategoryService interface {
	ListCategory(ctx context.Context, query model.CategoryQuery) (*model.CategoryPage, error)
	GetCategory(ctx context.Context, id uint64) (*model.Category, error)
	CreateCategory(ctx context.Context, payload model.CategoryPayload) (*model.Category, error)
	UpdateCategory(ctx context.Context, id uint64, payload model.CategoryPayload) (*model.Category, error)
	DeleteCategory(ctx context.Context, id uint64) error
}

type categoryService struct {
	*Service
	categoryRepository repository.CategoryRepository
}

func NewCategoryService(service *Service, repo repository.CategoryRepository) CategoryService {
	return &categoryService{Service: service, categoryRepository: repo}
}

func (s *categoryService) ListCategory(ctx context.Context, query model.CategoryQuery) (*model.CategoryPage, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}
	return s.categoryRepository.ListCategory(ctx, query)
}

func (s *categoryService) GetCategory(ctx context.Context, id uint64) (*model.Category, error) {
	if err := model.ValidateCategoryID(id); err != nil {
		return nil, err
	}
	return s.categoryRepository.GetCategory(ctx, id)
}

func (s *categoryService) CreateCategory(ctx context.Context, payload model.CategoryPayload) (*model.Category, error) {
	payload.ApplyDefaults()
	if err := payload.Validate(); err != nil {
		return nil, err
	}
	item := &model.Category{

		Name:   *payload.Name,
		Status: *payload.Status,
		Sort:   *payload.Sort,
		Remark: payload.Remark,
	}
	if err := s.categoryRepository.CreateCategory(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *categoryService) UpdateCategory(ctx context.Context, id uint64, payload model.CategoryPayload) (*model.Category, error) {
	if err := model.ValidateCategoryID(id); err != nil {
		return nil, err
	}
	if err := payload.Validate(); err != nil {
		return nil, err
	}
	return s.categoryRepository.UpdateCategory(ctx, id, payload)
}

func (s *categoryService) DeleteCategory(ctx context.Context, id uint64) error {
	if err := model.ValidateCategoryID(id); err != nil {
		return err
	}
	return s.categoryRepository.DeleteCategory(ctx, id)
}
