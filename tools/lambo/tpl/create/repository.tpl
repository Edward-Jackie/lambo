// 由 Lambo 生成，可按业务修改；再次生成遇到差异会停止。
package repository

import (
 "context"
 "errors"
 {{if .UsesSubstringSearch}}"strings"{{end}}
 "{{ .Module }}/internal/model"
 "gorm.io/gorm"
 "gorm.io/gorm/clause"
)

type {{ .StructName }}Repository interface {
 List{{.StructName}}(ctx context.Context, query model.{{.StructName}}Query) (*model.{{.StructName}}Page, error)
 Get{{.StructName}}(ctx context.Context, id {{.IDGoType}}) (*model.{{.StructName}}, error)
 Create{{.StructName}}(ctx context.Context, item *model.{{.StructName}}) error
 Update{{.StructName}}(ctx context.Context, id {{.IDGoType}}, payload model.{{.StructName}}Payload) (*model.{{.StructName}}, error)
 Delete{{.StructName}}(ctx context.Context, id {{.IDGoType}}) error
}

type {{ .LowerName }}Repository struct { *Repository }

func New{{ .StructName }}Repository(repository *Repository) {{ .StructName }}Repository {
 return &{{ .LowerName }}Repository{Repository: repository}
}

func (r *{{.LowerName}}Repository) List{{.StructName}}(ctx context.Context, query model.{{.StructName}}Query) (*model.{{.StructName}}Page, error) {
 db := r.db.WithContext(ctx).Model(&model.{{.StructName}}{})
{{range .SearchFields}} if query.{{.GoName}} != nil {
 {{if and (eq .Type "string") (not .Options)}}
  // 固定转义符使用户输入中的通配符按字面匹配，条件值始终参数化。
  value := strings.NewReplacer("!", "!!", "%", "!%", "_", "!_").Replace(*query.{{.GoName}})
  db = db.Where(clause.Expr{SQL: "? LIKE ? ESCAPE '!'", Vars: []interface{}{clause.Column{Name: {{json .ColumnName}}}, "%" + value + "%"}})
 {{else}}db = db.Where(clause.Eq{Column: {{json .ColumnName}}, Value: *query.{{.GoName}}}){{end}}
 }
{{end}} result := &model.{{.StructName}}Page{Items: make([]model.{{.StructName}}, 0)}
 if err := db.Count(&result.Total).Error; err != nil { return nil, err }
 err := db.Order(clause.OrderByColumn{Column: clause.Column{Name: "id"}, Desc: true}).
  Offset((query.Page - 1) * query.PageSize).Limit(query.PageSize).Find(&result.Items).Error
 return result, err
}

func (r *{{.LowerName}}Repository) Get{{.StructName}}(ctx context.Context, id {{.IDGoType}}) (*model.{{.StructName}}, error) {
 var item model.{{.StructName}}
 err := r.db.WithContext(ctx).Where(clause.Eq{Column: "id", Value: id}).First(&item).Error
 return &item, {{.LowerName}}RepositoryError(err)
}

func (r *{{.LowerName}}Repository) Create{{.StructName}}(ctx context.Context, item *model.{{.StructName}}) error {
 return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
  if err := tx.Create(item).Error; err != nil { return err }
  // 数字 ID 超出前端安全范围时回滚，避免保存后产生无法精确读取的记录。
  return model.Validate{{.StructName}}ID(item.ID)
 })
}

func (r *{{.LowerName}}Repository) Update{{.StructName}}(ctx context.Context, id {{.IDGoType}}, payload model.{{.StructName}}Payload) (*model.{{.StructName}}, error) {
 var item model.{{.StructName}}
 err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
  if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where(clause.Eq{Column: "id", Value: id}).First(&item).Error; err != nil { return err }
  // 使用字段白名单和 map 保留零值，缺省的可选字段写入 NULL；不修改只读字段。
  values := map[string]interface{}{
{{range .EditableFields}}   {{json .ColumnName}}: payload.{{.GoName}},
{{end}}  }
  if err := tx.Model(&item).Updates(values).Error; err != nil { return err }
  return tx.Where(clause.Eq{Column: "id", Value: id}).First(&item).Error
 })
 return &item, {{.LowerName}}RepositoryError(err)
}

func (r *{{.LowerName}}Repository) Delete{{.StructName}}(ctx context.Context, id {{.IDGoType}}) error {
 result := r.db.WithContext(ctx).Where(clause.Eq{Column: "id", Value: id}).Delete(&model.{{.StructName}}{})
 if result.Error != nil { return result.Error }
 if result.RowsAffected == 0 { return model.Err{{.StructName}}NotFound }
 return nil
}

func {{.LowerName}}RepositoryError(err error) error {
 if errors.Is(err, gorm.ErrRecordNotFound) { return model.Err{{.StructName}}NotFound }
 return err
}
