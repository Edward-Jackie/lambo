// 由 Lambo 生成，可按业务修改；再次生成遇到差异会停止。
package service

import (
	"context"

	"github.com/Edward-Jackie/lambo/backend/internal/model"
	"github.com/Edward-Jackie/lambo/backend/internal/repository"
)

type SystemMenuService interface {
	ListSystemMenu(ctx context.Context, query model.SystemMenuQuery) (*model.SystemMenuPage, error)
	GetSystemMenu(ctx context.Context, id uint64) (*model.SystemMenu, error)
	CreateSystemMenu(ctx context.Context, payload model.SystemMenuPayload) (*model.SystemMenu, error)
	UpdateSystemMenu(ctx context.Context, id uint64, payload model.SystemMenuPayload) (*model.SystemMenu, error)
	DeleteSystemMenu(ctx context.Context, id uint64) error
}

type systemMenuService struct {
	*Service
	systemMenuRepository repository.SystemMenuRepository
}

func NewSystemMenuService(service *Service, repo repository.SystemMenuRepository) SystemMenuService {
	return &systemMenuService{Service: service, systemMenuRepository: repo}
}

func (s *systemMenuService) ListSystemMenu(ctx context.Context, query model.SystemMenuQuery) (*model.SystemMenuPage, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}
	return s.systemMenuRepository.ListSystemMenu(ctx, query)
}

func (s *systemMenuService) GetSystemMenu(ctx context.Context, id uint64) (*model.SystemMenu, error) {
	if err := model.ValidateSystemMenuID(id); err != nil {
		return nil, err
	}
	return s.systemMenuRepository.GetSystemMenu(ctx, id)
}

func (s *systemMenuService) CreateSystemMenu(ctx context.Context, payload model.SystemMenuPayload) (*model.SystemMenu, error) {
	payload.ApplyDefaults()
	if err := payload.Validate(); err != nil {
		return nil, err
	}
	item := &model.SystemMenu{

		Name:       *payload.Name,
		Kind:       *payload.Kind,
		ParentId:   *payload.ParentId,
		RouteName:  payload.RouteName,
		Path:       payload.Path,
		Component:  payload.Component,
		Permission: payload.Permission,
		Icon:       payload.Icon,
		Sort:       *payload.Sort,
		Enabled:    *payload.Enabled,
	}
	if err := s.systemMenuRepository.CreateSystemMenu(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *systemMenuService) UpdateSystemMenu(ctx context.Context, id uint64, payload model.SystemMenuPayload) (*model.SystemMenu, error) {
	if err := model.ValidateSystemMenuID(id); err != nil {
		return nil, err
	}
	if err := payload.Validate(); err != nil {
		return nil, err
	}
	return s.systemMenuRepository.UpdateSystemMenu(ctx, id, payload)
}

func (s *systemMenuService) DeleteSystemMenu(ctx context.Context, id uint64) error {
	if err := model.ValidateSystemMenuID(id); err != nil {
		return err
	}
	return s.systemMenuRepository.DeleteSystemMenu(ctx, id)
}
