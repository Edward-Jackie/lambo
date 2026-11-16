// 由 Lambo 生成，可按业务修改；再次生成遇到差异会停止。
package model

import (
	"errors"
	"fmt"

	"strings"
	"time"
	"unicode/utf16"
)

const (
	SystemRoleDefaultPageSize = 20
	SystemRoleMaxPageSize     = 100
	SystemRoleMaxPage         = 1000000
)

var (
	ErrSystemRoleInvalid  = errors.New("参数无效")
	ErrSystemRoleNotFound = errors.New("记录不存在")
)

type SystemRole struct {
	MenuIDs []uint64 `json:"menuIds" gorm:"-"`
	ID      uint64   `json:"id" gorm:"primaryKey"`
	// 角色编码
	Code string `json:"code" gorm:"column:code;not null;size:64;uniqueIndex"`
	// 角色名称
	Name string `json:"name" gorm:"column:name;not null;size:64;index"`
	// 状态
	Enabled bool `json:"enabled" gorm:"column:enabled;not null;index"`
	// 备注
	Remark    *string   `json:"remark,omitempty" gorm:"column:remark;size:500"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (m *SystemRole) TableName() string { return "system_role" }

// SystemRolePayload 仅接收可编辑字段；指针区分缺省与合法零值。
type SystemRolePayload struct {
	MenuIDs []uint64 `json:"menuIds"`
	Code    *string  `json:"code"`
	Name    *string  `json:"name"`
	Enabled *bool    `json:"enabled"`
	Remark  *string  `json:"remark"`
}

type SystemRoleQuery struct {
	Page     int     `form:"page"`
	PageSize int     `form:"pageSize"`
	Code     *string `form:"code"`
	Name     *string `form:"name"`
	Enabled  *bool   `form:"enabled"`
}

type SystemRolePage struct {
	Items []SystemRole `json:"items"`
	Total int64        `json:"total"`
}

// ApplyDefaults 只在新增时应用默认值，编辑缺省的可选字段按完整替换清空。
func (p *SystemRolePayload) ApplyDefaults() {
	if p.Enabled == nil {
		value := true
		p.Enabled = &value
	}
}

func (p *SystemRolePayload) Validate() error {
	invalid := func(message string) error { return fmt.Errorf("%w: %s", ErrSystemRoleInvalid, message) }
	if p == nil {
		return invalid("缺少请求体")
	}
	if p.Code == nil {
		return invalid("请提供角色编码")
	} else {

		value := strings.TrimSpace(*p.Code)
		p.Code = &value
		if value == "" {
			return invalid("请输入角色编码")
		}

		if len(utf16.Encode([]rune(*p.Code))) > 64 {
			return invalid("角色编码超出最大长度")
		}

	}
	if p.Name == nil {
		return invalid("请提供角色名称")
	} else {

		value := strings.TrimSpace(*p.Name)
		p.Name = &value
		if value == "" {
			return invalid("请输入角色名称")
		}

		if len(utf16.Encode([]rune(*p.Name))) > 64 {
			return invalid("角色名称超出最大长度")
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
	if p.Remark == nil {

	} else {

		if len(utf16.Encode([]rune(*p.Remark))) > 500 {
			return invalid("备注超出最大长度")
		}

	}
	return nil
}

func (p *SystemRoleQuery) Validate() error {
	invalid := func(message string) error { return fmt.Errorf("%w: %s", ErrSystemRoleInvalid, message) }
	if p == nil {
		return invalid("缺少查询参数")
	}
	if p.Page < 1 || p.Page > SystemRoleMaxPage {
		return invalid("page 超出范围")
	}
	if p.PageSize < 1 || p.PageSize > SystemRoleMaxPageSize {
		return invalid("pageSize 超出范围")
	}
	if p.Code != nil {

		value := strings.TrimSpace(*p.Code)
		p.Code = &value
		if value == "" {
			return invalid("请输入角色编码")
		}

		if len(utf16.Encode([]rune(*p.Code))) > 64 {
			return invalid("角色编码超出最大长度")
		}

	}
	if p.Name != nil {

		value := strings.TrimSpace(*p.Name)
		p.Name = &value
		if value == "" {
			return invalid("请输入角色名称")
		}

		if len(utf16.Encode([]rune(*p.Name))) > 64 {
			return invalid("角色名称超出最大长度")
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

func ValidateSystemRoleID(id uint64) error {
	if id == 0 || id > 9007199254740991 {
		return fmt.Errorf("%w: ID 无效", ErrSystemRoleInvalid)
	}
	return nil
}
