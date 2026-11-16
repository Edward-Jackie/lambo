// 由 Lambo 生成，可按业务修改；再次生成遇到差异会停止。
package service

import (
	"context"

	"github.com/Edward-Jackie/lambo/backend/internal/model"
	"github.com/Edward-Jackie/lambo/backend/internal/repository"
)

type SystemRoleService interface {
	ListSystemRole(ctx context.Context, query model.SystemRoleQuery) (*model.SystemRolePage, error)
	GetSystemRole(ctx context.Context, id uint64) (*model.SystemRole, error)
	CreateSystemRole(ctx context.Context, payload model.SystemRolePayload) (*model.SystemRole, error)
	UpdateSystemRole(ctx context.Context, id uint64, payload model.SystemRolePayload) (*model.SystemRole, error)
	DeleteSystemRole(ctx context.Context, id uint64) error
}

type systemRoleService struct {
	*Service
	systemRoleRepository repository.SystemRoleRepository
}

func NewSystemRoleService(service *Service, repo repository.SystemRoleRepository) SystemRoleService {
	return &systemRoleService{Service: service, systemRoleRepository: repo}
}

func (s *systemRoleService) ListSystemRole(ctx context.Context, query model.SystemRoleQuery) (*model.SystemRolePage, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}
	return s.systemRoleRepository.ListSystemRole(ctx, query)
}

func (s *systemRoleService) GetSystemRole(ctx context.Context, id uint64) (*model.SystemRole, error) {
	if err := model.ValidateSystemRoleID(id); err != nil {
		return nil, err
	}
	return s.systemRoleRepository.GetSystemRole(ctx, id)
}

func (s *systemRoleService) CreateSystemRole(ctx context.Context, payload model.SystemRolePayload) (*model.SystemRole, error) {
	payload.ApplyDefaults()
	if err := payload.Validate(); err != nil {
		return nil, err
	}
	item := &model.SystemRole{
		MenuIDs: payload.MenuIDs,

		Code:    *payload.Code,
		Name:    *payload.Name,
		Enabled: *payload.Enabled,
		Remark:  payload.Remark,
	}
	if err := s.systemRoleRepository.CreateSystemRole(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *systemRoleService) UpdateSystemRole(ctx context.Context, id uint64, payload model.SystemRolePayload) (*model.SystemRole, error) {
	if err := model.ValidateSystemRoleID(id); err != nil {
		return nil, err
	}
	if err := payload.Validate(); err != nil {
		return nil, err
	}
	return s.systemRoleRepository.UpdateSystemRole(ctx, id, payload)
}

func (s *systemRoleService) DeleteSystemRole(ctx context.Context, id uint64) error {
	if err := model.ValidateSystemRoleID(id); err != nil {
		return err
	}
	return s.systemRoleRepository.DeleteSystemRole(ctx, id)
}
