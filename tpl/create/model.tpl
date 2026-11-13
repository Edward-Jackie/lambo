// 由 Lambo 生成的起始代码，可按业务修改；再次生成遇到差异会停止。
package model

import "time"

type {{ .StructName }} struct {
 ID {{if eq .IDType "string"}}string{{else}}uint64{{end}} `json:"id" gorm:"primaryKey"`
{{range .Fields}} // {{.Comment}}
 {{.GoName}} {{.GoType}} `json:"{{.Name}}"`
{{end}} CreatedAt time.Time `json:"createdAt"`
 UpdatedAt time.Time `json:"updatedAt"`
}

func (m *{{ .StructName }}) TableName() string { return "{{.SnakeName}}" }
