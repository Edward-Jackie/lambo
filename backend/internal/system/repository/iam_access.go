package repository

import (
	"context"
	"github.com/Edward-Jackie/lambo/backend/internal/system/model"
	"gorm.io/gorm"
	"sort"
	"strings"
)

func allMenus(db *gorm.DB) ([]model.SystemMenu, error) {
	items := make([]model.SystemMenu, 0)
	err := db.Order("sort ASC, id ASC").Limit(maxIAMMenus + 1).Find(&items).Error
	if len(items) > maxIAMMenus {
		return nil, iamConflict("菜单超过当前支持的 1000 项上限")
	}
	return items, err
}

func activeMenus(items []model.SystemMenu, grants map[uint64]bool) []model.SystemMenu {
	byID := map[uint64]model.SystemMenu{}
	for _, item := range items {
		byID[item.ID] = item
	}
	valid := func(item model.SystemMenu) bool {
		seen := map[uint64]bool{}
		for {
			if seen[item.ID] || !item.Enabled || !grants[item.ID] {
				return false
			}
			seen[item.ID] = true
			if item.ParentId == 0 {
				return true
			}
			parent, ok := byID[uint64(item.ParentId)]
			if !ok {
				return false
			}
			item = parent
		}
	}
	result := make([]model.SystemMenu, 0)
	for _, item := range items {
		if valid(item) {
			result = append(result, item)
		}
	}
	return result
}

func loadAccess(db *gorm.DB, account *model.Account) error {
	account.Roles = make([]model.SystemRole, 0)
	if err := db.Table("system_role").Select("system_role.*").Joins("JOIN lambo_account_roles ar ON ar.role_id=system_role.id").Where("ar.account_id=? AND system_role.enabled=?", account.ID, true).Order("system_role.id").Scan(&account.Roles).Error; err != nil {
		return err
	}
	grants := map[uint64]bool{}
	ids := make([]uint64, 0)
	if err := db.Table("lambo_role_menus rm").Select("DISTINCT rm.menu_id").Joins("JOIN system_role r ON r.id=rm.role_id AND r.enabled=?", true).Joins("JOIN lambo_account_roles ar ON ar.role_id=r.id").Where("ar.account_id=?", account.ID).Pluck("rm.menu_id", &ids).Error; err != nil {
		return err
	}
	for _, id := range ids {
		grants[id] = true
	}
	menus, err := allMenus(db)
	if err != nil {
		return err
	}
	account.Menus = activeMenus(menus, grants)
	codes := map[string]bool{}
	account.Codes = make([]string, 0)
	for _, m := range account.Menus {
		if code := textValue(m.Permission); code != "" {
			codes[code] = true
		}
	}
	for code := range codes {
		account.Codes = append(account.Codes, code)
	}
	sort.Strings(account.Codes)
	return nil
}

func checkScope(db *gorm.DB, ctx context.Context, menus []model.SystemMenu) error {
	actor := model.Account{ID: model.ActorID(ctx)}
	if err := loadAccess(db, &actor); err != nil {
		return err
	}
	owned := map[string]bool{}
	for _, c := range actor.Codes {
		owned[c] = true
	}
	for _, m := range menus {
		if code := textValue(m.Permission); code != "" && !owned[code] {
			return model.ErrIAMForbidden
		}
	}
	return nil
}

func roleScope(db *gorm.DB, ctx context.Context, roleIDs []uint64) error {
	if len(roleIDs) == 0 {
		return nil
	}
	var menus []model.SystemMenu
	if err := db.Table("system_menu m").Select("DISTINCT m.*").Joins("JOIN lambo_role_menus rm ON rm.menu_id=m.id").Where("rm.role_id IN ?", roleIDs).Scan(&menus).Error; err != nil {
		return err
	}
	return checkScope(db, ctx, menus)
}

// 批量读取关系，避免检查最后管理员时逐个账号查询数据库。
func ensureManager(tx *gorm.DB) error {
	var accounts []model.Account
	var roles []model.SystemRole
	var ar []model.AccountRole
	var rm []model.RoleMenu
	if err := tx.Where("enabled=?", true).Find(&accounts).Error; err != nil {
		return err
	}
	if err := tx.Where("enabled=?", true).Find(&roles).Error; err != nil {
		return err
	}
	if err := tx.Find(&ar).Error; err != nil {
		return err
	}
	if err := tx.Find(&rm).Error; err != nil {
		return err
	}
	menus, err := allMenus(tx)
	if err != nil {
		return err
	}
	enabled := map[uint64]bool{}
	for _, r := range roles {
		enabled[r.ID] = true
	}
	grants := map[uint64]map[uint64]bool{}
	for _, row := range rm {
		if enabled[row.RoleID] {
			if grants[row.RoleID] == nil {
				grants[row.RoleID] = map[uint64]bool{}
			}
			grants[row.RoleID][row.MenuID] = true
		}
	}
	for _, a := range accounts {
		own := map[uint64]bool{}
		for _, row := range ar {
			if row.AccountID == a.ID {
				for mid := range grants[row.RoleID] {
					own[mid] = true
				}
			}
		}
		codes := map[string]bool{}
		pages := map[string]bool{}
		for _, m := range activeMenus(menus, own) {
			codes[textValue(m.Permission)] = true
			if m.Kind == model.MenuPage {
				pages[textValue(m.Component)] = true
			}
		}
		good := true
		for _, c := range model.ManagementCodes() {
			if !codes[c] {
				good = false
			}
		}
		for _, c := range []string{"/business/system-user/index", "/business/system-role/index", "/business/system-menu/index"} {
			if !pages[c] {
				good = false
			}
		}
		if good {
			return nil
		}
	}
	return iamConflict("必须保留至少一个可登录且拥有完整用户、角色、菜单管理权限的账号")
}

func enrichUsers(db *gorm.DB, items []model.SystemUser) error {
	ids := make([]uint64, 0, len(items))
	for _, u := range items {
		ids = append(ids, u.ID)
	}
	if len(ids) == 0 {
		return nil
	}
	var rows []struct {
		AccountID uint64
		RoleID    uint64
		Name      string
	}
	if err := db.Table("lambo_account_roles ar").Select("ar.account_id, ar.role_id, r.name").Joins("JOIN system_role r ON r.id=ar.role_id").Where("ar.account_id IN ?", ids).Order("r.id").Scan(&rows).Error; err != nil {
		return err
	}
	for i := range items {
		items[i].RoleIDs = make([]uint64, 0)
		names := make([]string, 0)
		for _, r := range rows {
			if r.AccountID == items[i].ID {
				items[i].RoleIDs = append(items[i].RoleIDs, r.RoleID)
				names = append(names, r.Name)
			}
		}
		items[i].RoleNames = strings.Join(names, "、")
	}
	return nil
}

func enrichRoles(db *gorm.DB, items []model.SystemRole) error {
	ids := make([]uint64, 0, len(items))
	for _, r := range items {
		ids = append(ids, r.ID)
	}
	if len(ids) == 0 {
		return nil
	}
	var rows []model.RoleMenu
	if err := db.Where("role_id IN ?", ids).Order("menu_id").Find(&rows).Error; err != nil {
		return err
	}
	for i := range items {
		items[i].MenuIDs = make([]uint64, 0)
		for _, row := range rows {
			if row.RoleID == items[i].ID {
				items[i].MenuIDs = append(items[i].MenuIDs, row.MenuID)
			}
		}
	}
	return nil
}
