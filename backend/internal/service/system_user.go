// 由 Lambo 生成，可按业务修改；再次生成遇到差异会停止。
package service

import (
	"context"
	"fmt"
	"golang.org/x/crypto/bcrypt"

	"github.com/Edward-Jackie/lambo/backend/internal/model"
	"github.com/Edward-Jackie/lambo/backend/internal/repository"
)

type SystemUserService interface {
	ListSystemUser(ctx context.Context, query model.SystemUserQuery) (*model.SystemUserPage, error)
	GetSystemUser(ctx context.Context, id uint64) (*model.SystemUser, error)
	CreateSystemUser(ctx context.Context, payload model.SystemUserPayload) (*model.SystemUser, error)
	UpdateSystemUser(ctx context.Context, id uint64, payload model.SystemUserPayload) (*model.SystemUser, error)
	DeleteSystemUser(ctx context.Context, id uint64) error
	ResetSystemUserPassword(ctx context.Context, id uint64, password string) error
}

type systemUserService struct {
	*Service
	systemUserRepository repository.SystemUserRepository
}

func NewSystemUserService(service *Service, repo repository.SystemUserRepository) SystemUserService {
	return &systemUserService{Service: service, systemUserRepository: repo}
}

func (s *systemUserService) ListSystemUser(ctx context.Context, query model.SystemUserQuery) (*model.SystemUserPage, error) {
	if err := query.Validate(); err != nil {
		return nil, err
	}
	return s.systemUserRepository.ListSystemUser(ctx, query)
}

func (s *systemUserService) GetSystemUser(ctx context.Context, id uint64) (*model.SystemUser, error) {
	if err := model.ValidateSystemUserID(id); err != nil {
		return nil, err
	}
	return s.systemUserRepository.GetSystemUser(ctx, id)
}

func (s *systemUserService) CreateSystemUser(ctx context.Context, payload model.SystemUserPayload) (*model.SystemUser, error) {
	payload.ApplyDefaults()
	if err := payload.Validate(); err != nil {
		return nil, err
	}
	if payload.Password == nil {
		return nil, fmt.Errorf("%w: 请提供初始密码", model.ErrIAMInvalid)
	}
	hash, err := hashAccountPassword(*payload.Password)
	if err != nil {
		return nil, err
	}
	item := &model.SystemUser{
		PasswordHash: hash, RoleIDs: payload.RoleIDs,

		Username: *payload.Username,
		RealName: *payload.RealName,
		Enabled:  *payload.Enabled,
		Remark:   payload.Remark,
	}
	if err := s.systemUserRepository.CreateSystemUser(ctx, item); err != nil {
		return nil, err
	}
	return item, nil
}

func (s *systemUserService) UpdateSystemUser(ctx context.Context, id uint64, payload model.SystemUserPayload) (*model.SystemUser, error) {
	if err := model.ValidateSystemUserID(id); err != nil {
		return nil, err
	}
	if err := payload.Validate(); err != nil {
		return nil, err
	}
	if payload.Password != nil {
		return nil, fmt.Errorf("%w: 修改密码请使用重置密码操作", model.ErrIAMInvalid)
	}
	return s.systemUserRepository.UpdateSystemUser(ctx, id, payload)
}

func (s *systemUserService) DeleteSystemUser(ctx context.Context, id uint64) error {
	if err := model.ValidateSystemUserID(id); err != nil {
		return err
	}
	return s.systemUserRepository.DeleteSystemUser(ctx, id)
}

func hashAccountPassword(password string) (string, error) {
	if len(password) < 12 || len(password) > 72 {
		return "", fmt.Errorf("%w: 密码必须为 12—72 字节", model.ErrIAMInvalid)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}
func (s *systemUserService) ResetSystemUserPassword(ctx context.Context, id uint64, password string) error {
	if err := model.ValidateSystemUserID(id); err != nil {
		return err
	}
	hash, err := hashAccountPassword(password)
	if err != nil {
		return err
	}
	return s.systemUserRepository.ResetSystemUserPassword(ctx, id, hash)
}
