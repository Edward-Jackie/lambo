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
	CategoryDefaultPageSize = 20
	CategoryMaxPageSize     = 100
	CategoryMaxPage         = 1000000
)

var (
	ErrCategoryInvalid  = errors.New("参数无效")
	ErrCategoryNotFound = errors.New("记录不存在")
)

type Category struct {
	ID uint64 `json:"id" gorm:"primaryKey"`
	// 分类名称
	Name string `json:"name" gorm:"column:name;not null;size:64;index"`
	// 状态
	Status float64 `json:"status" gorm:"column:status;not null;index"`
	// 排序
	Sort float64 `json:"sort" gorm:"column:sort;not null"`
	// 备注
	Remark    *string   `json:"remark,omitempty" gorm:"column:remark;size:500"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func (m *Category) TableName() string { return "category" }

// CategoryPayload 仅接收可编辑字段；指针区分缺省与合法零值。
type CategoryPayload struct {
	Name   *string  `json:"name"`
	Status *float64 `json:"status"`
	Sort   *float64 `json:"sort"`
	Remark *string  `json:"remark"`
}

type CategoryQuery struct {
	Page     int      `form:"page"`
	PageSize int      `form:"pageSize"`
	Name     *string  `form:"name"`
	Status   *float64 `form:"status"`
}

type CategoryPage struct {
	Items []Category `json:"items"`
	Total int64      `json:"total"`
}

// ApplyDefaults 只在新增时应用默认值，编辑缺省的可选字段按完整替换清空。
func (p *CategoryPayload) ApplyDefaults() {
	if p.Status == nil {
		value := float64(1)
		p.Status = &value
	}
	if p.Sort == nil {
		value := float64(0)
		p.Sort = &value
	}
}

func (p *CategoryPayload) Validate() error {
	invalid := func(message string) error { return fmt.Errorf("%w: %s", ErrCategoryInvalid, message) }
	if p == nil {
		return invalid("缺少请求体")
	}
	if p.Name == nil {
		return invalid("请提供分类名称")
	} else {

		value := strings.TrimSpace(*p.Name)
		p.Name = &value
		if value == "" {
			return invalid("请输入分类名称")
		}

		if len(utf16.Encode([]rune(*p.Name))) > 64 {
			return invalid("分类名称超出最大长度")
		}

	}
	if p.Status == nil {
		return invalid("请提供状态")
	} else {

		if math.IsNaN(*p.Status) || math.IsInf(*p.Status, 0) || (math.Trunc(*p.Status) == *p.Status && math.Abs(*p.Status) > 9007199254740991) {
			return invalid("状态必须是有效数字")
		}

		switch *p.Status {
		case 1, 0:
		default:
			return invalid("请选择有效的状态")
		}

	}
	if p.Sort == nil {
		return invalid("请提供排序")
	} else {

		if math.IsNaN(*p.Sort) || math.IsInf(*p.Sort, 0) || (math.Trunc(*p.Sort) == *p.Sort && math.Abs(*p.Sort) > 9007199254740991) {
			return invalid("排序必须是有效数字")
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

func (p *CategoryQuery) Validate() error {
	invalid := func(message string) error { return fmt.Errorf("%w: %s", ErrCategoryInvalid, message) }
	if p == nil {
		return invalid("缺少查询参数")
	}
	if p.Page < 1 || p.Page > CategoryMaxPage {
		return invalid("page 超出范围")
	}
	if p.PageSize < 1 || p.PageSize > CategoryMaxPageSize {
		return invalid("pageSize 超出范围")
	}
	if p.Name != nil {

		value := strings.TrimSpace(*p.Name)
		p.Name = &value
		if value == "" {
			return invalid("请输入分类名称")
		}

		if len(utf16.Encode([]rune(*p.Name))) > 64 {
			return invalid("分类名称超出最大长度")
		}

	}
	if p.Status != nil {

		if math.IsNaN(*p.Status) || math.IsInf(*p.Status, 0) || (math.Trunc(*p.Status) == *p.Status && math.Abs(*p.Status) > 9007199254740991) {
			return invalid("状态必须是有效数字")
		}

		switch *p.Status {
		case 1, 0:
		default:
			return invalid("请选择有效的状态")
		}

	}
	return nil
}

func ValidateCategoryID(id uint64) error {
	if id == 0 || id > 9007199254740991 {
		return fmt.Errorf("%w: ID 无效", ErrCategoryInvalid)
	}
	return nil
}
