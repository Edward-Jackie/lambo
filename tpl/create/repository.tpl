// 由 Lambo 生成的起始代码，可按业务修改；再次生成遇到差异会停止。
package repository

import (
 "context"
 "fmt"
 "{{ .Module }}/internal/model"
)

type {{ .StructName }}Repository interface {
 Get{{ .StructName }}(ctx context.Context, id {{if eq .IDType "string"}}string{{else}}uint64{{end}}) (*model.{{ .StructName }}, error)
}

type {{ .LowerName }}Repository struct { *Repository }

func New{{ .StructName }}Repository(repository *Repository) {{ .StructName }}Repository {
 return &{{ .LowerName }}Repository{Repository: repository}
}

func (r *{{ .LowerName }}Repository) Get{{ .StructName }}(ctx context.Context, id {{if eq .IDType "string"}}string{{else}}uint64{{end}}) (*model.{{ .StructName }}, error) {
 // 接入真实数据库、数据范围和错误映射后再注册此组件，避免占位成功。
 return nil, fmt.Errorf("{{ .StructName }} 数据访问尚未实现")
}
