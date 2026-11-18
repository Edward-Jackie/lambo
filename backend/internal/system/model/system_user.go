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
	SystemUserDefaultPageSize = 20
	SystemUserMaxPageSize     = 100
	SystemUserMaxPage         = 1000000
)

var (
	ErrSystemUserInvalid  = errors.New("参数无效")
	ErrSystemUserNotFound = errors.New("记录不存在")
)

type SystemUser struct {
	// 兼容旧表非空列；权限迁移完成后不再使用此字段授权。
	LegacyRole   string   `json:"-" gorm:"column:role;size:32;not null;default:''"`
	PasswordHash string   `json:"-" gorm:"size:100;not null"`
	RoleIDs      []uint64 `json:"roleIds" gorm:"-"`
	RoleNames    string   `json:"roleNames" gorm:"-"`
	ID           uint64   `json:"id" gorm:"primaryKey"`
	// 用户名
	Username string `json:"username" gorm:"column:username;not null;size:64;uniqueIndex:ux_lambo_accounts_username"`
	// 姓名
	RealName string `json:"realName" gorm:"column:real_name;not null;size:64;index"`
	// 状态
	Enabled bool `json:"enabled" gorm:"column:enabled;not null;index"`
	// 备注
	Remark    *string   `json:"remark,omitempty" gorm:"column:remark;size:500"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (m *SystemUser) TableName() string { return "lambo_accounts" }

// SystemUserPayload 仅接收可编辑字段；指针区分缺省与合法零值。
type SystemUserPayload struct {
	RoleIDs  []uint64 `json:"roleIds"`
	Password *string  `json:"password,omitempty"`
	Username *string  `json:"username"`
	RealName *string  `json:"realName"`
	Enabled  *bool    `json:"enabled"`
	Remark   *string  `json:"remark"`
}

type SystemUserQuery struct {
	RoleID   *uint64 `form:"roleId"`
	Page     int     `form:"page"`
	PageSize int     `form:"pageSize"`
	Username *string `form:"username"`
	RealName *string `form:"realName"`
	Enabled  *bool   `form:"enabled"`
}

type SystemUserPage struct {
	Items []SystemUser `json:"items"`
	Total int64        `json:"total"`
}

// ApplyDefaults 只在新增时应用默认值，编辑缺省的可选字段按完整替换清空。
func (p *SystemUserPayload) ApplyDefaults() {
	if p.Enabled == nil {
		value := true
		p.Enabled = &value
	}
}

func (p *SystemUserPayload) Validate() error {
	invalid := func(message string) error { return fmt.Errorf("%w: %s", ErrSystemUserInvalid, message) }
	if p == nil {
		return invalid("缺少请求体")
	}
	if p.Username == nil {
		return invalid("请提供用户名")
	} else {

		value := strings.TrimSpace(*p.Username)
		p.Username = &value
		if value == "" {
			return invalid("请输入用户名")
		}

		if len(utf16.Encode([]rune(*p.Username))) > 64 {
			return invalid("用户名超出最大长度")
		}

	}
	if p.RealName == nil {
		return invalid("请提供姓名")
	} else {

		value := strings.TrimSpace(*p.RealName)
		p.RealName = &value
		if value == "" {
			return invalid("请输入姓名")
		}

		if len(utf16.Encode([]rune(*p.RealName))) > 64 {
			return invalid("姓名超出最大长度")
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

func (p *SystemUserQuery) Validate() error {
	invalid := func(message string) error { return fmt.Errorf("%w: %s", ErrSystemUserInvalid, message) }
	if p == nil {
		return invalid("缺少查询参数")
	}
	if p.Page < 1 || p.Page > SystemUserMaxPage {
		return invalid("page 超出范围")
	}
	if p.PageSize < 1 || p.PageSize > SystemUserMaxPageSize {
		return invalid("pageSize 超出范围")
	}
	if p.Username != nil {

		value := strings.TrimSpace(*p.Username)
		p.Username = &value
		if value == "" {
			return invalid("请输入用户名")
		}

		if len(utf16.Encode([]rune(*p.Username))) > 64 {
			return invalid("用户名超出最大长度")
		}

	}
	if p.RealName != nil {

		value := strings.TrimSpace(*p.RealName)
		p.RealName = &value
		if value == "" {
			return invalid("请输入姓名")
		}

		if len(utf16.Encode([]rune(*p.RealName))) > 64 {
			return invalid("姓名超出最大长度")
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

func ValidateSystemUserID(id uint64) error {
	if id == 0 || id > 9007199254740991 {
		return fmt.Errorf("%w: ID 无效", ErrSystemUserInvalid)
	}
	return nil
}
