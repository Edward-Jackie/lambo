// 由 Lambo 生成，可按业务修改；再次生成遇到差异会停止。
package model

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
	"unicode/utf16"
)

const (
	SystemMenuDefaultPageSize = 20
	SystemMenuMaxPageSize     = 100
	SystemMenuMaxPage         = 1000000
)

var (
	ErrSystemMenuInvalid  = errors.New("参数无效")
	ErrSystemMenuNotFound = errors.New("记录不存在")
)

type SystemMenu struct {
	ID uint64 `json:"id" gorm:"primaryKey"`
	// 菜单名称
	Name string `json:"name" gorm:"column:name;not null;size:64;index"`
	// 类型
	Kind string `json:"kind" gorm:"column:kind;not null;index"`
	// 上级菜单
	ParentId float64 `json:"parentId" gorm:"column:parent_id;not null"`
	// 路由名称
	RouteName *string `json:"routeName,omitempty" gorm:"column:route_name;size:64"`
	// 路由路径
	Path *string `json:"path,omitempty" gorm:"column:path;size:128"`
	// 页面组件
	Component *string `json:"component,omitempty" gorm:"column:component;size:128"`
	// 权限码
	Permission *string `json:"permission,omitempty" gorm:"column:permission;size:96"`
	// 图标
	Icon *string `json:"icon,omitempty" gorm:"column:icon;size:64"`
	// 排序
	Sort float64 `json:"sort" gorm:"column:sort;not null"`
	// 状态
	Enabled   bool      `json:"enabled" gorm:"column:enabled;not null;index"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (m *SystemMenu) TableName() string { return "system_menu" }

// SystemMenuPayload 仅接收可编辑字段；指针区分缺省与合法零值。
type SystemMenuPayload struct {
	Name       *string  `json:"name"`
	Kind       *string  `json:"kind"`
	ParentId   *float64 `json:"parentId"`
	RouteName  *string  `json:"routeName"`
	Path       *string  `json:"path"`
	Component  *string  `json:"component"`
	Permission *string  `json:"permission"`
	Icon       *string  `json:"icon"`
	Sort       *float64 `json:"sort"`
	Enabled    *bool    `json:"enabled"`
}

type SystemMenuQuery struct {
	Page     int     `form:"page"`
	PageSize int     `form:"pageSize"`
	Name     *string `form:"name"`
	Kind     *string `form:"kind"`
	Enabled  *bool   `form:"enabled"`
}

type SystemMenuPage struct {
	Items []SystemMenu `json:"items"`
	Total int64        `json:"total"`
}

// ApplyDefaults 只在新增时应用默认值，编辑缺省的可选字段按完整替换清空。
func (p *SystemMenuPayload) ApplyDefaults() {
	if p.ParentId == nil {
		value := float64(0)
		p.ParentId = &value
	}
	if p.Sort == nil {
		value := float64(0)
		p.Sort = &value
	}
	if p.Enabled == nil {
		value := true
		p.Enabled = &value
	}
}

func (p *SystemMenuPayload) Validate() error {
	invalid := func(message string) error { return fmt.Errorf("%w: %s", ErrSystemMenuInvalid, message) }
	if p == nil {
		return invalid("缺少请求体")
	}
	if p.Name == nil {
		return invalid("请提供菜单名称")
	} else {

		value := strings.TrimSpace(*p.Name)
		p.Name = &value
		if value == "" {
			return invalid("请输入菜单名称")
		}

		if len(utf16.Encode([]rune(*p.Name))) > 64 {
			return invalid("菜单名称超出最大长度")
		}

	}
	if p.Kind == nil {
		return invalid("请提供类型")
	} else {

		value := strings.TrimSpace(*p.Kind)
		p.Kind = &value
		if value == "" {
			return invalid("请输入类型")
		}

		switch *p.Kind {
		case "directory", "page", "button":
		default:
			return invalid("请选择有效的类型")
		}

	}
	if p.ParentId == nil {
		return invalid("请提供上级菜单")
	} else {

		if math.IsNaN(*p.ParentId) || math.IsInf(*p.ParentId, 0) || (math.Trunc(*p.ParentId) == *p.ParentId && math.Abs(*p.ParentId) > 9007199254740991) {
			return invalid("上级菜单必须是有效数字")
		}

	}
	if p.RouteName == nil {

	} else {

		if len(utf16.Encode([]rune(*p.RouteName))) > 64 {
			return invalid("路由名称超出最大长度")
		}

	}
	if p.Path == nil {

	} else {

		if len(utf16.Encode([]rune(*p.Path))) > 128 {
			return invalid("路由路径超出最大长度")
		}

	}
	if p.Component == nil {

	} else {

		if len(utf16.Encode([]rune(*p.Component))) > 128 {
			return invalid("页面组件超出最大长度")
		}

	}
	if p.Permission == nil {

	} else {

		if len(utf16.Encode([]rune(*p.Permission))) > 96 {
			return invalid("权限码超出最大长度")
		}

	}
	if p.Icon == nil {

	} else {

		if len(utf16.Encode([]rune(*p.Icon))) > 64 {
			return invalid("图标超出最大长度")
		}

	}
	if p.Sort == nil {
		return invalid("请提供排序")
	} else {

		if math.IsNaN(*p.Sort) || math.IsInf(*p.Sort, 0) || (math.Trunc(*p.Sort) == *p.Sort && math.Abs(*p.Sort) > 9007199254740991) {
			return invalid("排序必须是有效数字")
		}

	}
	if p.Enabled == nil {
		return invalid("请提供状态")
	} else {

		switch *p.Enabled {
		case true, false:
		default:
			return invalid("请选择有效的状态")
		}

	}
	return nil
}

func (p *SystemMenuQuery) Validate() error {
	invalid := func(message string) error { return fmt.Errorf("%w: %s", ErrSystemMenuInvalid, message) }
	if p == nil {
		return invalid("缺少查询参数")
	}
	if p.Page < 1 || p.Page > SystemMenuMaxPage {
		return invalid("page 超出范围")
	}
	if p.PageSize < 1 || p.PageSize > SystemMenuMaxPageSize {
		return invalid("pageSize 超出范围")
	}
	if p.Name != nil {

		value := strings.TrimSpace(*p.Name)
		p.Name = &value
		if value == "" {
			return invalid("请输入菜单名称")
		}

		if len(utf16.Encode([]rune(*p.Name))) > 64 {
			return invalid("菜单名称超出最大长度")
		}

	}
	if p.Kind != nil {

		value := strings.TrimSpace(*p.Kind)
		p.Kind = &value
		if value == "" {
			return invalid("请输入类型")
		}

		switch *p.Kind {
		case "directory", "page", "button":
		default:
			return invalid("请选择有效的类型")
		}

	}
	if p.Enabled != nil {

		switch *p.Enabled {
		case true, false:
		default:
			return invalid("请选择有效的状态")
		}

	}
	return nil
}

func ValidateSystemMenuID(id uint64) error {
	if id == 0 || id > 9007199254740991 {
		return fmt.Errorf("%w: ID 无效", ErrSystemMenuInvalid)
	}
	return nil
}
