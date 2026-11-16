// 由 Lambo 生成，可按业务修改；再次生成遇到差异会停止。
package repository

import (
	"context"
	"errors"
	"github.com/Edward-Jackie/lambo/backend/internal/business/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"strings"
)

type CategoryRepository interface {
	ListCategory(ctx context.Context, query model.CategoryQuery) (*model.CategoryPage, error)
	GetCategory(ctx context.Context, id uint64) (*model.Category, error)
	CreateCategory(ctx context.Context, item *model.Category) error
	UpdateCategory(ctx context.Context, id uint64, payload model.CategoryPayload) (*model.Category, error)
	DeleteCategory(ctx context.Context, id uint64) error
}

type categoryRepository struct{ *Repository }

func NewCategoryRepository(repository *Repository) CategoryRepository {
	return &categoryRepository{Repository: repository}
}

func (r *categoryRepository) ListCategory(ctx context.Context, query model.CategoryQuery) (*model.CategoryPage, error) {
	db := r.db.WithContext(ctx).Model(&model.Category{})
	if query.Name != nil {

		// 固定转义符使用户输入中的通配符按字面匹配，条件值始终参数化。
		value := strings.NewReplacer("!", "!!", "%", "!%", "_", "!_").Replace(*query.Name)
		db = db.Where(clause.Expr{SQL: "? LIKE ? ESCAPE '!'", Vars: []interface{}{clause.Column{Name: "name"}, "%" + value + "%"}})

	}
	if query.Status != nil {
		db = db.Where(clause.Eq{Column: "status", Value: *query.Status})
	}
	result := &model.CategoryPage{Items: make([]model.Category, 0)}
	if err := db.Count(&result.Total).Error; err != nil {
		return nil, err
	}
	err := db.Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}, Desc: true}).
		Offset((query.Page - 1) * query.PageSize).Limit(query.PageSize).Find(&result.Items).Error
	return result, err
}

func (r *categoryRepository) GetCategory(ctx context.Context, id uint64) (*model.Category, error) {
	var item model.Category
	err := r.db.WithContext(ctx).Where(clause.Eq{Column: "id", Value: id}).First(&item).Error
	return &item, categoryRepositoryError(err)
}

func (r *categoryRepository) CreateCategory(ctx context.Context, item *model.Category) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(item).Error; err != nil {
			return err
		}
		// 数字 ID 超出前端安全范围时回滚，避免保存后产生无法精确读取的记录。
		return model.ValidateCategoryID(item.ID)
	})
}

func (r *categoryRepository) UpdateCategory(ctx context.Context, id uint64, payload model.CategoryPayload) (*model.Category, error) {
	var item model.Category
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(clause.Eq{Column: "id", Value: id}).First(&item).Error; err != nil {
			return err
		}
		// 使用字段白名单和 map 保留零值，缺省的可选字段写入 NULL；不修改只读字段。
		values := map[string]interface{}{
			"name":   payload.Name,
			"status": payload.Status,
			"sort":   payload.Sort,
			"remark": payload.Remark,
		}
		if err := tx.Model(&item).Updates(values).Error; err != nil {
			return err
		}
		return tx.Where(clause.Eq{Column: "id", Value: id}).First(&item).Error
	})
	return &item, categoryRepositoryError(err)
}

func (r *categoryRepository) DeleteCategory(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).Where(clause.Eq{Column: "id", Value: id}).Delete(&model.Category{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return model.ErrCategoryNotFound
	}
	return nil
}

func categoryRepositoryError(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.ErrCategoryNotFound
	}
	return err
}
