// 由 Lambo 生成，可按业务修改；再次生成遇到差异会停止。
package repository

import (
	"context"
	"errors"
	"github.com/Edward-Jackie/lambo/backend/internal/system/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
)

type SystemMenuRepository interface {
	ListSystemMenu(ctx context.Context, query model.SystemMenuQuery) (*model.SystemMenuPage, error)
	GetSystemMenu(ctx context.Context, id uint64) (*model.SystemMenu, error)
	CreateSystemMenu(ctx context.Context, item *model.SystemMenu) error
	UpdateSystemMenu(ctx context.Context, id uint64, payload model.SystemMenuPayload) (*model.SystemMenu, error)
	DeleteSystemMenu(ctx context.Context, id uint64) error
}

type systemMenuRepository struct{ *Repository }

func NewSystemMenuRepository(repository *Repository) SystemMenuRepository {
	return &systemMenuRepository{Repository: repository}
}

func (r *systemMenuRepository) ListSystemMenu(ctx context.Context, query model.SystemMenuQuery) (*model.SystemMenuPage, error) {
	db := r.db.WithContext(ctx).Model(&model.SystemMenu{})
	if query.Name != nil {

		// 固定转义符使用户输入中的通配符按字面匹配，条件值始终参数化。
		value := strings.NewReplacer("!", "!!", "%", "!%", "_", "!_").Replace(*query.Name)
		db = db.Where(clause.Expr{SQL: "? LIKE ? ESCAPE '!'", Vars: []interface{}{clause.Column{Name: "name"}, "%" + value + "%"}})

	}
	if query.Kind != nil {
		db = db.Where(clause.Eq{Column: "kind", Value: *query.Kind})
	}
	if query.Enabled != nil {
		db = db.Where(clause.Eq{Column: "enabled", Value: *query.Enabled})
	}
	result := &model.SystemMenuPage{Items: make([]model.SystemMenu, 0)}
	if err := db.Count(&result.Total).Error; err != nil {
		return nil, err
	}
	err := db.Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}, Desc: true}).
		Offset((query.Page - 1) * query.PageSize).Limit(query.PageSize).Find(&result.Items).Error
	return result, err
}

func (r *systemMenuRepository) GetSystemMenu(ctx context.Context, id uint64) (*model.SystemMenu, error) {
	var item model.SystemMenu
	err := r.db.WithContext(ctx).Where(clause.Eq{Column: "id", Value: id}).First(&item).Error
	return &item, systemMenuRepositoryError(err)
}

func (r *systemMenuRepository) CreateSystemMenu(ctx context.Context, item *model.SystemMenu) error {
	return r.iamWrite(ctx, func(tx *gorm.DB) error {
		if err := r.validateMenu(tx, item); err != nil {
			return err
		}
		if err := checkScope(tx, ctx, []model.SystemMenu{*item}); err != nil {
			return err
		}
		return persistSystemMenu(tx, item)
	})
}

func persistSystemMenu(tx *gorm.DB, item *model.SystemMenu) error {
	if err := tx.Create(item).Error; err != nil {
		return err
	}
	// 数字 ID 超出前端安全范围时回滚，避免保存后产生无法精确读取的记录。
	return model.ValidateSystemMenuID(item.ID)
}

func (r *systemMenuRepository) UpdateSystemMenu(ctx context.Context, id uint64, payload model.SystemMenuPayload) (*model.SystemMenu, error) {
	var item model.SystemMenu
	err := r.iamWrite(ctx, func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(clause.Eq{Column: "id", Value: id}).First(&item).Error; err != nil {
			return err
		}
		if err := checkScope(tx, ctx, []model.SystemMenu{item}); err != nil {
			return err
		}
		// 使用字段白名单和 map 保留零值，缺省的可选字段写入 NULL；不修改只读字段。
		values := map[string]interface{}{
			"name":       payload.Name,
			"kind":       payload.Kind,
			"parent_id":  payload.ParentId,
			"route_name": payload.RouteName,
			"path":       payload.Path,
			"component":  payload.Component,
			"permission": payload.Permission,
			"icon":       payload.Icon,
			"sort":       payload.Sort,
			"enabled":    payload.Enabled,
		}
		if err := tx.Model(&item).Updates(values).Error; err != nil {
			return err
		}
		if err := tx.Where(clause.Eq{Column: "id", Value: id}).First(&item).Error; err != nil {
			return err
		}
		if err := r.validateMenu(tx, &item); err != nil {
			return err
		}
		if err := checkScope(tx, ctx, []model.SystemMenu{item}); err != nil {
			return err
		}
		return revokeMenu(tx, id)
	})
	return &item, systemMenuRepositoryError(err)
}

func (r *systemMenuRepository) DeleteSystemMenu(ctx context.Context, id uint64) error {
	return r.iamWrite(ctx, func(tx *gorm.DB) error {
		var item model.SystemMenu
		if err := tx.First(&item, id).Error; err != nil {
			return systemMenuRepositoryError(err)
		}
		if err := checkScope(tx, ctx, []model.SystemMenu{item}); err != nil {
			return err
		}
		var count int64
		if err := tx.Model(&model.SystemMenu{}).Where("parent_id=?", id).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return iamConflict("菜单仍有子节点，请先删除子节点")
		}
		if err := tx.Model(&model.RoleMenu{}).Where("menu_id=?", id).Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return iamConflict("菜单已被角色授权，请先解除关联")
		}
		return tx.Delete(&item).Error
	})
}

func systemMenuRepositoryError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.ErrSystemMenuNotFound
	}
	return err
}
