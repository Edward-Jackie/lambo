package repository

import (
	"context"
	"strings"

	"github.com/Edward-Jackie/lambo/backend/internal/model"
	"gorm.io/gorm"
)

// InitializeIAM 只迁移一次旧角色与初始授权，不在后续启动时恢复人工撤销的权限。
func (r *AuthRepository) InitializeIAM(ctx context.Context) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", iamWriteLock).Error; err != nil {
			return err
		}
		var count int64
		if err := tx.Model(&model.IAMMigration{}).Where("key=?", "rbac-v1").Count(&count).Error; err != nil {
			return err
		}
		if count > 0 {
			return nil
		}
		var accounts []model.Account
		if err := tx.Find(&accounts).Error; err != nil {
			return err
		}
		if len(accounts) == 0 {
			return iamInvalid("首次初始化权限需要先配置 bootstrap 管理账号")
		}
		admin := model.SystemRole{Code: model.RoleAdmin, Name: "管理员", Enabled: true}
		reader := model.SystemRole{Code: model.RoleReader, Name: "只读用户", Enabled: true}
		if err := tx.Create(&admin).Error; err != nil {
			return err
		}
		if err := tx.Create(&reader).Error; err != nil {
			return err
		}
		str := func(s string) *string { return &s }
		system := model.SystemMenu{Name: "系统管理", Kind: model.MenuDirectory, RouteName: str("SystemManagement"), Path: str("/business/system"), Icon: str("lucide:settings"), Sort: 20, Enabled: true}
		if err := tx.Create(&system).Error; err != nil {
			return err
		}
		all := []uint64{system.ID}
		readerIDs := []uint64{}
		for index, entry := range []struct{ name, prefix, component, path, route, icon string }{
			{"分类管理", "catalog:category", "/business/category/index", "/business/category", "BusinessCategory", "lucide:folder"},
			{"用户管理", "system:user", "/business/system-user/index", "/business/system-user", "BusinessSystemUser", "lucide:users"},
			{"角色管理", "system:role", "/business/system-role/index", "/business/system-role", "BusinessSystemRole", "lucide:shield"},
			{"菜单管理", "system:menu", "/business/system-menu/index", "/business/system-menu", "BusinessSystemMenu", "lucide:menu"},
		} {
			parent := float64(system.ID)
			if index == 0 {
				parent = 0
			}
			page := model.SystemMenu{Name: entry.name, Kind: model.MenuPage, ParentId: parent, RouteName: str(entry.route), Path: str(entry.path), Component: str(entry.component), Permission: str(entry.prefix + ":query"), Icon: str(entry.icon), Sort: float64(index), Enabled: true}
			if err := tx.Create(&page).Error; err != nil {
				return err
			}
			all = append(all, page.ID)
			if index == 0 {
				readerIDs = append(readerIDs, page.ID)
			}
			for code, label := range model.PermissionCatalog() {
				if strings.HasPrefix(code, entry.prefix+":") && !strings.HasSuffix(code, ":query") {
					button := model.SystemMenu{Name: label, Kind: model.MenuButton, ParentId: float64(page.ID), Permission: str(code), Enabled: true}
					if err := tx.Create(&button).Error; err != nil {
						return err
					}
					all = append(all, button.ID)
				}
			}
		}
		for _, id := range all {
			if err := tx.Create(&model.RoleMenu{RoleID: admin.ID, MenuID: id}).Error; err != nil {
				return err
			}
		}
		for _, id := range readerIDs {
			if err := tx.Create(&model.RoleMenu{RoleID: reader.ID, MenuID: id}).Error; err != nil {
				return err
			}
		}
		for _, a := range accounts {
			var rid uint64
			switch a.Role {
			case model.RoleAdmin:
				rid = admin.ID
			case model.RoleReader:
				rid = reader.ID
			}
			if rid != 0 {
				if err := tx.Create(&model.AccountRole{AccountID: a.ID, RoleID: rid}).Error; err != nil {
					return err
				}
			}
		}
		if err := ensureManager(tx); err != nil {
			return err
		}
		if err := tx.Where("1=1").Delete(&model.Session{}).Error; err != nil {
			return err
		}
		return tx.Create(&model.IAMMigration{Key: "rbac-v1"}).Error
	})
}

func (r *AuthRepository) ManagementOptions(ctx context.Context) (map[string]interface{}, error) {
	roles := make([]model.SystemRole, 0)
	if err := r.db.WithContext(ctx).Order("id").Find(&roles).Error; err != nil {
		return nil, err
	}
	if err := enrichRoles(r.db.WithContext(ctx), roles); err != nil {
		return nil, err
	}
	menus, err := allMenus(r.db.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"roles": roles, "menus": menus, "permissions": model.PermissionCatalog(), "components": model.PageComponents()}, nil
}
