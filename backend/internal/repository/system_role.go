// 由 Lambo 生成，可按业务修改；再次生成遇到差异会停止。
package repository

import (
	"context"
	"errors"
	"github.com/Edward-Jackie/lambo/backend/internal/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
)

type SystemRoleRepository interface {
	ListSystemRole(ctx context.Context, query model.SystemRoleQuery) (*model.SystemRolePage, error)
	GetSystemRole(ctx context.Context, id uint64) (*model.SystemRole, error)
	CreateSystemRole(ctx context.Context, item *model.SystemRole) error
	UpdateSystemRole(ctx context.Context, id uint64, payload model.SystemRolePayload) (*model.SystemRole, error)
	DeleteSystemRole(ctx context.Context, id uint64) error
}

type systemRoleRepository struct{ *Repository }

func NewSystemRoleRepository(repository *Repository) SystemRoleRepository {
	return &systemRoleRepository{Repository: repository}
}

func (r *systemRoleRepository) ListSystemRole(ctx context.Context, query model.SystemRoleQuery) (*model.SystemRolePage, error) {
	db := r.db.WithContext(ctx).Model(&model.SystemRole{})
	if query.Code != nil {

		// 固定转义符使用户输入中的通配符按字面匹配，条件值始终参数化。
		value := strings.NewReplacer("!", "!!", "%", "!%", "_", "!_").Replace(*query.Code)
		db = db.Where(clause.Expr{SQL: "? LIKE ? ESCAPE '!'", Vars: []interface{}{clause.Column{Name: "code"}, "%" + value + "%"}})

	}
	if query.Name != nil {

		// 固定转义符使用户输入中的通配符按字面匹配，条件值始终参数化。
		value := strings.NewReplacer("!", "!!", "%", "!%", "_", "!_").Replace(*query.Name)
		db = db.Where(clause.Expr{SQL: "? LIKE ? ESCAPE '!'", Vars: []interface{}{clause.Column{Name: "name"}, "%" + value + "%"}})

	}
	if query.Enabled != nil {
		db = db.Where(clause.Eq{Column: "enabled", Value: *query.Enabled})
	}
	result := &model.SystemRolePage{Items: make([]model.SystemRole, 0)}
	if err := db.Count(&result.Total).Error; err != nil {
		return nil, err
	}
	err := db.Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}, Desc: true}).
		Offset((query.Page - 1) * query.PageSize).Limit(query.PageSize).Find(&result.Items).Error
	if err == nil {
		err = enrichRoles(r.db.WithContext(ctx), result.Items)
	}
	return result, err
}

func (r *systemRoleRepository) GetSystemRole(ctx context.Context, id uint64) (*model.SystemRole, error) {
	var item model.SystemRole
	err := r.db.WithContext(ctx).Where(clause.Eq{Column: "id", Value: id}).First(&item).Error
	if err == nil {
		items := []model.SystemRole{item}
		err = enrichRoles(r.db.WithContext(ctx), items)
		item = items[0]
	}
	return &item, systemRoleRepositoryError(err)
}

func (r *systemRoleRepository) CreateSystemRole(ctx context.Context, item *model.SystemRole) error {
	return r.iamWrite(ctx, func(tx *gorm.DB) error {
		if !iamCodePattern.MatchString(item.Code) {
			return iamInvalid("角色编码格式无效")
		}
		if err := tx.Create(item).Error; err != nil {
			return err
		}
		// 数字 ID 超出前端安全范围时回滚，避免保存后产生无法精确读取的记录。
		if err := model.ValidateSystemRoleID(item.ID); err != nil {
			return err
		}
		return setRoleMenus(tx, ctx, item.ID, item.MenuIDs)
	})
}

func (r *systemRoleRepository) UpdateSystemRole(ctx context.Context, id uint64, payload model.SystemRolePayload) (*model.SystemRole, error) {
	var item model.SystemRole
	err := r.iamWrite(ctx, func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(clause.Eq{Column: "id", Value: id}).First(&item).Error; err != nil {
			return err
		}
		if !iamCodePattern.MatchString(*payload.Code) {
			return iamInvalid("角色编码格式无效")
		}
		if err := roleScope(tx, ctx, []uint64{id}); err != nil {
			return err
		}
		// 使用字段白名单和 map 保留零值，缺省的可选字段写入 NULL；不修改只读字段。
		values := map[string]interface{}{
			"code":    payload.Code,
			"name":    payload.Name,
			"enabled": payload.Enabled,
			"remark":  payload.Remark,
		}
		if err := tx.Model(&item).Updates(values).Error; err != nil {
			return err
		}
		if err := setRoleMenus(tx, ctx, id, payload.MenuIDs); err != nil {
			return err
		}
		if err := revokeRole(tx, id); err != nil {
			return err
		}
		return tx.Where(clause.Eq{Column: "id", Value: id}).First(&item).Error
	})
	if err == nil {
		items := []model.SystemRole{item}
		err = enrichRoles(r.db.WithContext(ctx), items)
		item = items[0]
	}
	return &item, systemRoleRepositoryError(err)
}

func (r *systemRoleRepository) DeleteSystemRole(ctx context.Context, id uint64) error {
	return r.iamWrite(ctx, func(tx *gorm.DB) error {
		var item model.SystemRole
		if err := tx.First(&item, id).Error; err != nil {
			return systemRoleRepositoryError(err)
		}
		if err := roleScope(tx, ctx, []uint64{id}); err != nil {
			return err
		}
		var count int64
		if err := tx.Model(&model.AccountRole{}).Where("role_id=?", id).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return iamConflict("角色仍分配给用户，请先解除关联")
		}
		if err := tx.Where("role_id=?", id).Delete(&model.RoleMenu{}).Error; err != nil {
			return err
		}
		return tx.Delete(&item).Error
	})
}

func systemRoleRepositoryError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.ErrSystemRoleNotFound
	}
	return err
}
