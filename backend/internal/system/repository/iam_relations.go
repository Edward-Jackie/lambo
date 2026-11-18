package repository

import (
	"context"
	"github.com/Edward-Jackie/lambo/backend/internal/system/model"
	"gorm.io/gorm"
	"math"
	"strings"
)

func setAccountRoles(tx *gorm.DB, ctx context.Context, id uint64, roleIDs []uint64) error {
	if roleIDs == nil || len(roleIDs) > 100 {
		return iamInvalid("请提供角色列表，最多 100 个")
	}
	ids, err := uniqueIDs(roleIDs)
	if err != nil {
		return err
	}
	var count int64
	if len(ids) > 0 {
		if err := tx.Model(&model.SystemRole{}).Where("id IN ?", ids).Count(&count).Error; err != nil {
			return err
		}
		if count != int64(len(ids)) {
			return iamInvalid("角色不存在")
		}
	}
	if err := roleScope(tx, ctx, ids); err != nil {
		return err
	}
	if err := tx.Where("account_id=?", id).Delete(&model.AccountRole{}).Error; err != nil {
		return err
	}
	for _, rid := range ids {
		if err := tx.Create(&model.AccountRole{AccountID: id, RoleID: rid}).Error; err != nil {
			return err
		}
	}
	return nil
}

func setRoleMenus(tx *gorm.DB, ctx context.Context, id uint64, menuIDs []uint64) error {
	if menuIDs == nil || len(menuIDs) > maxIAMMenus {
		return iamInvalid("请提供菜单权限列表")
	}
	ids, err := uniqueIDs(menuIDs)
	if err != nil {
		return err
	}
	items, err := allMenus(tx)
	if err != nil {
		return err
	}
	byID := map[uint64]model.SystemMenu{}
	for _, m := range items {
		byID[m.ID] = m
	}
	selected := map[uint64]bool{}
	for _, mid := range ids {
		seen := map[uint64]bool{}
		for mid != 0 {
			if seen[mid] {
				return iamInvalid("菜单关系存在循环")
			}
			seen[mid] = true
			m, ok := byID[mid]
			if !ok {
				return iamInvalid("菜单不存在")
			}
			selected[mid] = true
			mid = uint64(m.ParentId)
		}
	}
	chosen := make([]model.SystemMenu, 0)
	for mid := range selected {
		chosen = append(chosen, byID[mid])
	}
	if err := checkScope(tx, ctx, chosen); err != nil {
		return err
	}
	if err := tx.Where("role_id=?", id).Delete(&model.RoleMenu{}).Error; err != nil {
		return err
	}
	for mid := range selected {
		if err := tx.Create(&model.RoleMenu{RoleID: id, MenuID: mid}).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) validateMenu(tx *gorm.DB, item *model.SystemMenu) error {
	if math.Trunc(item.ParentId) != item.ParentId || item.ParentId < 0 || math.Trunc(item.Sort) != item.Sort {
		return iamInvalid("上级菜单和排序必须是有效整数")
	}
	menus, err := allMenus(tx)
	if err != nil {
		return err
	}
	if item.ID == 0 && len(menus) >= maxIAMMenus {
		return iamConflict("菜单最多 1000 项")
	}
	byID := map[uint64]model.SystemMenu{}
	for _, m := range menus {
		byID[m.ID] = m
	}
	parentID := uint64(item.ParentId)
	if parentID != 0 {
		parent, ok := byID[parentID]
		if !ok {
			return iamInvalid("上级菜单不存在")
		}
		if item.Kind == model.MenuButton && parent.Kind != model.MenuPage {
			return iamInvalid("按钮必须属于页面")
		}
		if item.Kind != model.MenuButton && parent.Kind != model.MenuDirectory {
			return iamInvalid("目录和页面的上级必须是目录")
		}
	}
	if item.Kind == model.MenuButton && parentID == 0 {
		return iamInvalid("按钮必须属于页面")
	}
	seen := map[uint64]bool{}
	for id := parentID; id != 0; {
		if id == item.ID || seen[id] {
			return iamInvalid("上级菜单不能指向自身或后代")
		}
		seen[id] = true
		id = uint64(byID[id].ParentId)
	}
	for _, m := range menus {
		if m.ID == item.ID {
			continue
		}
		if item.ID != 0 && uint64(m.ParentId) == item.ID {
			if (m.Kind == model.MenuButton && item.Kind != model.MenuPage) || (m.Kind != model.MenuButton && item.Kind != model.MenuDirectory) {
				return iamConflict("当前菜单仍有不兼容的子节点")
			}
		}
	}
	code := textValue(item.Permission)
	if code != "" {
		if _, ok := r.menuCatalog.Permissions[code]; !ok {
			return iamInvalid("请选择已注册的接口权限码")
		}
	}
	if item.Kind == model.MenuDirectory {
		if code != "" || textValue(item.Component) != "" {
			return iamInvalid("目录不配置接口权限或页面组件")
		}
	}
	if item.Kind == model.MenuPage {
		if _, ok := r.menuCatalog.Components[textValue(item.Component)]; !ok {
			return iamInvalid("请选择已注册的页面组件")
		}
		if code == "" || !strings.HasSuffix(code, ":query") {
			return iamInvalid("页面需要查询权限码")
		}
	}
	if item.Kind == model.MenuButton {
		if code == "" || strings.HasSuffix(code, ":query") {
			return iamInvalid("按钮需要操作权限码")
		}
		if textValue(item.Component) != "" || textValue(item.Path) != "" || textValue(item.RouteName) != "" {
			return iamInvalid("按钮不配置路由或页面组件")
		}
	}
	if item.Kind != model.MenuButton {
		if !iamCodePattern.MatchString(textValue(item.RouteName)) || !iamPathPattern.MatchString(textValue(item.Path)) || !strings.HasPrefix(textValue(item.Path), "/business/") {
			return iamInvalid("请提供有效且位于 /business/ 下的路由名称和路径")
		}
	}
	for _, m := range menus {
		if m.ID == item.ID {
			continue
		}
		if (textValue(item.Path) != "" && textValue(item.Path) == textValue(m.Path)) || (textValue(item.RouteName) != "" && textValue(item.RouteName) == textValue(m.RouteName)) {
			return iamConflict("路由名称和路径不能重复")
		}
	}
	return nil
}
