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

type SystemUserRepository interface {
	ListSystemUser(ctx context.Context, query model.SystemUserQuery) (*model.SystemUserPage, error)
	GetSystemUser(ctx context.Context, id uint64) (*model.SystemUser, error)
	CreateSystemUser(ctx context.Context, item *model.SystemUser) error
	UpdateSystemUser(ctx context.Context, id uint64, payload model.SystemUserPayload) (*model.SystemUser, error)
	DeleteSystemUser(ctx context.Context, id uint64) error
	ResetSystemUserPassword(ctx context.Context, id uint64, hash string) error
}

type systemUserRepository struct{ *Repository }

func NewSystemUserRepository(repository *Repository) SystemUserRepository {
	return &systemUserRepository{Repository: repository}
}

func (r *systemUserRepository) ListSystemUser(ctx context.Context, query model.SystemUserQuery) (*model.SystemUserPage, error) {
	db := r.db.WithContext(ctx).Model(&model.SystemUser{})
	if query.RoleID != nil {
		db = db.Where("id IN (?)", r.db.Model(&model.AccountRole{}).Select("account_id").Where("role_id=?", *query.RoleID))
	}
	if query.Username != nil {

		// 固定转义符使用户输入中的通配符按字面匹配，条件值始终参数化。
		value := strings.NewReplacer("!", "!!", "%", "!%", "_", "!_").Replace(*query.Username)
		db = db.Where(clause.Expr{SQL: "? LIKE ? ESCAPE '!'", Vars: []interface{}{clause.Column{Name: "username"}, "%" + value + "%"}})

	}
	if query.RealName != nil {

		// 固定转义符使用户输入中的通配符按字面匹配，条件值始终参数化。
		value := strings.NewReplacer("!", "!!", "%", "!%", "_", "!_").Replace(*query.RealName)
		db = db.Where(clause.Expr{SQL: "? LIKE ? ESCAPE '!'", Vars: []interface{}{clause.Column{Name: "real_name"}, "%" + value + "%"}})

	}
	if query.Enabled != nil {
		db = db.Where(clause.Eq{Column: "enabled", Value: *query.Enabled})
	}
	result := &model.SystemUserPage{Items: make([]model.SystemUser, 0)}
	if err := db.Count(&result.Total).Error; err != nil {
		return nil, err
	}
	err := db.Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}, Desc: true}).
		Offset((query.Page - 1) * query.PageSize).Limit(query.PageSize).Find(&result.Items).Error
	if err == nil {
		err = enrichUsers(r.db.WithContext(ctx), result.Items)
	}
	return result, err
}

func (r *systemUserRepository) GetSystemUser(ctx context.Context, id uint64) (*model.SystemUser, error) {
	var item model.SystemUser
	err := r.db.WithContext(ctx).Where(clause.Eq{Column: "id", Value: id}).First(&item).Error
	if err == nil {
		items := []model.SystemUser{item}
		err = enrichUsers(r.db.WithContext(ctx), items)
		item = items[0]
	}
	return &item, systemUserRepositoryError(err)
}

func (r *systemUserRepository) CreateSystemUser(ctx context.Context, item *model.SystemUser) error {
	return r.iamWrite(ctx, func(tx *gorm.DB) error {
		if !iamCodePattern.MatchString(item.Username) {
			return iamInvalid("用户名仅支持字母开头的字母、数字、下划线和短横线")
		}
		if err := tx.Create(item).Error; err != nil {
			return err
		}
		// 数字 ID 超出前端安全范围时回滚，避免保存后产生无法精确读取的记录。
		if err := model.ValidateSystemUserID(item.ID); err != nil {
			return err
		}
		return setAccountRoles(tx, ctx, item.ID, item.RoleIDs)
	})
}

func (r *systemUserRepository) UpdateSystemUser(ctx context.Context, id uint64, payload model.SystemUserPayload) (*model.SystemUser, error) {
	var item model.SystemUser
	err := r.iamWrite(ctx, func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(clause.Eq{Column: "id", Value: id}).First(&item).Error; err != nil {
			return err
		}
		if !iamCodePattern.MatchString(*payload.Username) {
			return iamInvalid("用户名格式无效")
		}
		var oldRoles []uint64
		if err := tx.Model(&model.AccountRole{}).Where("account_id=?", id).Pluck("role_id", &oldRoles).Error; err != nil {
			return err
		}
		if err := roleScope(tx, ctx, oldRoles); err != nil {
			return err
		}
		// 使用字段白名单和 map 保留零值，缺省的可选字段写入 NULL；不修改只读字段。
		values := map[string]interface{}{
			"username":  payload.Username,
			"real_name": payload.RealName,
			"enabled":   payload.Enabled,
			"remark":    payload.Remark,
		}
		if err := tx.Model(&item).Updates(values).Error; err != nil {
			return err
		}
		if err := setAccountRoles(tx, ctx, id, payload.RoleIDs); err != nil {
			return err
		}
		if err := revokeAccount(tx, id); err != nil {
			return err
		}
		return tx.Where(clause.Eq{Column: "id", Value: id}).First(&item).Error
	})
	if err == nil {
		items := []model.SystemUser{item}
		err = enrichUsers(r.db.WithContext(ctx), items)
		item = items[0]
	}
	return &item, systemUserRepositoryError(err)
}

func (r *systemUserRepository) DeleteSystemUser(ctx context.Context, id uint64) error {
	return r.iamWrite(ctx, func(tx *gorm.DB) error {
		var item model.SystemUser
		if err := tx.First(&item, id).Error; err != nil {
			return systemUserRepositoryError(err)
		}
		var ids []uint64
		if err := tx.Model(&model.AccountRole{}).Where("account_id=?", id).Pluck("role_id", &ids).Error; err != nil {
			return err
		}
		if err := roleScope(tx, ctx, ids); err != nil {
			return err
		}
		if err := tx.Where("account_id=?", id).Delete(&model.AccountRole{}).Error; err != nil {
			return err
		}
		if err := revokeAccount(tx, id); err != nil {
			return err
		}
		return tx.Delete(&item).Error
	})
}
func (r *systemUserRepository) ResetSystemUserPassword(ctx context.Context, id uint64, hash string) error {
	return r.iamWrite(ctx, func(tx *gorm.DB) error {
		var item model.SystemUser
		if err := tx.First(&item, id).Error; err != nil {
			return systemUserRepositoryError(err)
		}
		var ids []uint64
		if err := tx.Model(&model.AccountRole{}).Where("account_id=?", id).Pluck("role_id", &ids).Error; err != nil {
			return err
		}
		if err := roleScope(tx, ctx, ids); err != nil {
			return err
		}
		if err := tx.Model(&item).Update("password_hash", hash).Error; err != nil {
			return err
		}
		return revokeAccount(tx, id)
	})
}

func systemUserRepositoryError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.ErrSystemUserNotFound
	}
	return err
}
