package create

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/token"
	"io"
	"math"
	"os"
	"regexp"
	"strings"

	"github.com/duke-git/lancet/v2/strutil"
)

// Spec 是当前开发期生成契约，不替代后续字典后台与枚举发布模型。
type Spec struct {
	Name       string  `json:"name"`
	Title      string  `json:"title"`
	API        string  `json:"api"`
	Permission string  `json:"permission"`
	IDType     string  `json:"idType"`
	Fields     []Field `json:"fields"`
}

type Field struct {
	Name      string      `json:"name"`
	Label     string      `json:"label"`
	Type      string      `json:"type"`
	Required  bool        `json:"required"`
	Search    bool        `json:"search"`
	ReadOnly  bool        `json:"readOnly"`
	MaxLength int         `json:"maxLength"`
	Default   interface{} `json:"default"`
	Options   []Option    `json:"options"`
}

type Option struct {
	Label string      `json:"label"`
	Value interface{} `json:"value"`
}

type templateData struct {
	Spec
	Module string
}

var namePattern = regexp.MustCompile(`^[a-z][a-z0-9]*(?:[-_][a-z0-9]+)*$`)
var fieldPattern = regexp.MustCompile(`^[a-z][a-zA-Z0-9]*$`)
var apiPattern = regexp.MustCompile(`^/[a-zA-Z0-9_-]+(?:/[a-zA-Z0-9_-]+)*$`)
var permissionPattern = regexp.MustCompile(`^[a-z][a-z0-9:_-]*$`)

func loadSpec(name, path string) (Spec, error) {
	spec := Spec{Name: name, Title: name, API: "/api/" + name, Permission: "business:" + name, IDType: "number"}
	if !namePattern.MatchString(name) {
		return spec, fmt.Errorf("名称 %q 必须使用小写英文、数字及中间的连字符或下划线", name)
	}
	if path == "" {
		spec.Fields = []Field{{Name: "name", Label: "名称", Type: "string", Required: true, Search: true, MaxLength: 128}}
	} else {
		data, err := os.ReadFile(path)
		if err != nil {
			return spec, fmt.Errorf("读取字段定义：%w", err)
		}
		decoder := json.NewDecoder(bytes.NewReader(data))
		decoder.DisallowUnknownFields()
		decoder.UseNumber()
		if err := decoder.Decode(&spec); err != nil {
			return spec, fmt.Errorf("解析字段定义：%w", err)
		}
		var extra interface{}
		if err := decoder.Decode(&extra); err != io.EOF {
			return spec, fmt.Errorf("字段定义只能包含一个 JSON 对象")
		}
	}
	if spec.Name != name {
		return spec, fmt.Errorf("字段定义名称 %q 与命令名称 %q 不一致", spec.Name, name)
	}
	if strings.TrimSpace(spec.Title) == "" || !apiPattern.MatchString(spec.API) || !permissionPattern.MatchString(spec.Permission) {
		return spec, fmt.Errorf("必须提供标题、以 / 开头的静态 API 路径和合法权限前缀")
	}
	if spec.IDType != "number" && spec.IDType != "string" {
		return spec, fmt.Errorf("idType 只支持 number 或 string")
	}
	if len(spec.Fields) == 0 {
		return spec, fmt.Errorf("至少需要一个业务字段")
	}
	names, goNames := map[string]bool{"id": true, "constructor": true, "prototype": true}, map[string]bool{"ID": true, "Id": true, "CreatedAt": true, "UpdatedAt": true, "DeletedAt": true, "TableName": true}
	editable := 0
	for i := range spec.Fields {
		field := &spec.Fields[i]
		goName := field.GoName()
		if !fieldPattern.MatchString(field.Name) || names[field.Name] || goNames[goName] || !token.IsIdentifier(goName) {
			return spec, fmt.Errorf("字段名 %q 非法、重复或与系统字段冲突", field.Name)
		}
		if field.Search && (field.Name == "page" || field.Name == "pageSize") {
			return spec, fmt.Errorf("查询字段 %s 与分页参数冲突", field.Name)
		}
		names[field.Name], goNames[goName] = true, true
		if strings.TrimSpace(field.Label) == "" {
			return spec, fmt.Errorf("字段 %s 缺少标签", field.Name)
		}
		switch field.Type {
		case "string", "number", "boolean":
		default:
			return spec, fmt.Errorf("字段 %s 不支持类型 %q，目前支持 string/number/boolean", field.Name, field.Type)
		}
		if field.MaxLength < 0 || (field.MaxLength > 0 && field.Type != "string") {
			return spec, fmt.Errorf("字段 %s 的 maxLength 仅支持非负字符串长度", field.Name)
		}
		values := map[string]bool{}
		for _, option := range field.Options {
			key := jsonValue(option.Value)
			if strings.TrimSpace(option.Label) == "" || !field.accepts(option.Value) || values[key] {
				return spec, fmt.Errorf("字段 %s 的选项标签、类型或值唯一性无效", field.Name)
			}
			values[key] = true
		}
		if field.Default != nil {
			if !field.accepts(field.Default) || (len(values) > 0 && !values[jsonValue(field.Default)]) {
				return spec, fmt.Errorf("字段 %s 的默认值不符合字段类型或选项", field.Name)
			}
			if value, ok := field.Default.(string); ok && field.MaxLength > 0 && len([]rune(value)) > field.MaxLength {
				return spec, fmt.Errorf("字段 %s 默认值超出最大长度", field.Name)
			}
		}
		if !field.ReadOnly {
			editable++
		}
	}
	if editable == 0 {
		return spec, fmt.Errorf("至少需要一个可编辑字段")
	}
	return spec, nil
}

func (s Spec) StructName() string { return strutil.UpperFirst(strutil.CamelCase(s.Name)) }
func (s Spec) LowerName() string  { return strutil.LowerFirst(s.StructName()) }
func (s Spec) SnakeName() string  { return strutil.SnakeCase(s.StructName()) }
func (s Spec) EditableFields() []Field {
	fields := []Field{}
	for _, field := range s.Fields {
		if !field.ReadOnly {
			fields = append(fields, field)
		}
	}
	return fields
}
func (s Spec) SearchFields() []Field {
	fields := []Field{}
	for _, field := range s.Fields {
		if field.Search {
			fields = append(fields, field)
		}
	}
	return fields
}
func (f Field) Comment() string { return strings.Join(strings.Fields(f.Label), " ") }
func (f Field) GoName() string  { return strutil.UpperFirst(f.Name) }
func (f Field) GoType() string {
	switch f.Type {
	case "number":
		return "float64"
	case "boolean":
		return "bool"
	default:
		return "string"
	}
}
func (f Field) Component() string {
	if len(f.Options) > 0 || f.Type == "boolean" {
		return "Select"
	}
	if f.Type == "number" {
		return "InputNumber"
	}
	return "Input"
}
func (f Field) Choices() []Option {
	if len(f.Options) > 0 {
		return f.Options
	}
	if f.Type == "boolean" {
		return []Option{{Label: "是", Value: true}, {Label: "否", Value: false}}
	}
	return nil
}
func (f Field) DefaultJSON() string {
	if f.Default != nil {
		return jsonValue(f.Default)
	}
	if len(f.Options) > 0 {
		return "undefined"
	}
	switch f.Type {
	case "string":
		return `""`
	case "boolean":
		return "false"
	default:
		return "undefined"
	}
}
func (f Field) accepts(value interface{}) bool {
	switch f.Type {
	case "string":
		_, ok := value.(string)
		return ok
	case "number":
		number, ok := value.(json.Number)
		if !ok {
			return false
		}
		value, err := number.Float64()
		return err == nil && !math.IsNaN(value) && !math.IsInf(value, 0) && (math.Trunc(value) != value || math.Abs(value) <= 9007199254740991)
	case "boolean":
		_, ok := value.(bool)
		return ok
	default:
		return false
	}
}
func jsonValue(value interface{}) string {
	if number, ok := value.(json.Number); ok {
		value, _ = number.Float64()
	}
	data, _ := json.Marshal(value)
	return string(data)
}
