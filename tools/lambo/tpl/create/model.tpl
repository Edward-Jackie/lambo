// 由 Lambo 生成，可按业务修改；再次生成遇到差异会停止。
package model

import (
 "errors"
 "fmt"
 {{if .UsesNumberValidation}}"math"{{end}}
 {{if .UsesStringValidation}}"strings"{{end}}
 "time"
 {{if .UsesStringLength}}"unicode/utf16"{{end}}
)

const (
 {{.StructName}}DefaultPageSize = 20
 {{.StructName}}MaxPageSize = 100
 {{.StructName}}MaxPage = 1000000
)

var (
 Err{{.StructName}}Invalid = errors.New("参数无效")
 Err{{.StructName}}NotFound = errors.New("记录不存在")
)

type {{.StructName}} struct {
 ID {{.IDGoType}} `json:"id" gorm:"primaryKey{{if eq .IDType "string"}};size:128{{end}}"`
{{range .Fields}} // {{.Comment}}
 {{.GoName}} {{if not .Required}}*{{end}}{{.GoType}} `json:"{{.Name}}{{if not .Required}},omitempty{{end}}" gorm:"column:{{.ColumnName}}{{if .Required}};not null{{end}}{{if .MaxLength}};size:{{.MaxLength}}{{end}}{{if .Search}};index{{end}}"`
{{end}} CreatedAt time.Time `json:"createdAt"`
 UpdatedAt time.Time `json:"updatedAt"`
}

func (m *{{.StructName}}) TableName() string { return "{{.SnakeName}}" }

// {{.StructName}}Payload 仅接收可编辑字段；指针区分缺省与合法零值。
type {{.StructName}}Payload struct {
{{range .EditableFields}} {{.GoName}} *{{.GoType}} `json:"{{.Name}}"`
{{end}}}

type {{.StructName}}Query struct {
 Page int `form:"page"`
 PageSize int `form:"pageSize"`
{{range .SearchFields}} {{.GoName}} *{{.GoType}} `form:"{{.Name}}"`
{{end}}}

type {{.StructName}}Page struct {
 Items []{{.StructName}} `json:"items"`
 Total int64 `json:"total"`
}

// ApplyDefaults 只在新增时应用默认值，编辑缺省的可选字段按完整替换清空。
func (p *{{.StructName}}Payload) ApplyDefaults() {
{{range .EditableFields}}{{if .HasDefault}} if p.{{.GoName}} == nil {
  value := {{if eq .Type "number"}}float64({{json .Default}}){{else}}{{json .Default}}{{end}}
  p.{{.GoName}} = &value
 }
{{end}}{{end}}}

func (p *{{.StructName}}Payload) Validate() error {
 invalid := func(message string) error { return fmt.Errorf("%w: %s", Err{{.StructName}}Invalid, message) }
 if p == nil { return invalid("缺少请求体") }
{{range .EditableFields}} if p.{{.GoName}} == nil {
  {{if .Required}}return invalid({{json (printf "请提供%s" .Label)}}){{end}}
 } else {
  {{template "validate-field" .}}
 }
{{end}} return nil
}

func (p *{{.StructName}}Query) Validate() error {
 invalid := func(message string) error { return fmt.Errorf("%w: %s", Err{{.StructName}}Invalid, message) }
 if p == nil { return invalid("缺少查询参数") }
 if p.Page < 1 || p.Page > {{.StructName}}MaxPage { return invalid("page 超出范围") }
 if p.PageSize < 1 || p.PageSize > {{.StructName}}MaxPageSize { return invalid("pageSize 超出范围") }
{{range .SearchFields}} if p.{{.GoName}} != nil {
  {{template "validate-field" .}}
 }
{{end}} return nil
}

func Validate{{.StructName}}ID(id {{.IDGoType}}) error {
 {{if eq .IDType "string"}}if id == "" || len(id) > 128 || strings.TrimSpace(id) != id || strings.ContainsAny(id, "/\\\x00") {
 {{else}}if id == 0 || id > 9007199254740991 {
 {{end}} return fmt.Errorf("%w: ID 无效", Err{{.StructName}}Invalid)
 }
 return nil
}

{{define "validate-field"}}
 {{if eq .Type "string"}}{{if .Required}}
 value := strings.TrimSpace(*p.{{.GoName}})
 p.{{.GoName}} = &value
 if value == "" { return invalid({{json (printf "请输入%s" .Label)}}) }
 {{end}}{{if .MaxLength}}
 if len(utf16.Encode([]rune(*p.{{.GoName}}))) > {{.MaxLength}} { return invalid({{json (printf "%s超出最大长度" .Label)}}) }
 {{end}}{{else if eq .Type "number"}}
 if math.IsNaN(*p.{{.GoName}}) || math.IsInf(*p.{{.GoName}}, 0) || (math.Trunc(*p.{{.GoName}}) == *p.{{.GoName}} && math.Abs(*p.{{.GoName}}) > 9007199254740991) {
  return invalid({{json (printf "%s必须是有效数字" .Label)}})
 }
 {{end}}{{if .Options}}
 switch *p.{{.GoName}} {
 case {{range $index, $option := .Options}}{{if $index}},{{end}}{{json $option.Value}}{{end}}:
 default: return invalid({{json (printf "请选择有效的%s" .Label)}})
 }
 {{end}}
{{end}}
