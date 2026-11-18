package repository

import (
	"context"
	"crypto/md5"
	"fmt"

	"github.com/Edward-Jackie/lambo/backend/internal/system/model"
	"gorm.io/gorm"
)

const (
	iamBootstrapKey  = "rbac-v1"
	menuSeedFreshKey = "menu-seed-fresh-v3"
)

// MenuSeedDefinition 是生成器和后端运维命令共用的种子身份。
type MenuSeedDefinition struct {
	Domain     string `json:"domain"`
	Name       string `json:"name"`
	Title      string `json:"title"`
	RouteName  string `json:"routeName"`
	Permission string `json:"permission"`
}

func iamMarked(tx *gorm.DB, key string) (bool, error) {
	var count int64
	err := tx.Model(&model.IAMMigration{}).Where("key = ?", key).Count(&count).Error
	return count > 0, err
}

// ApplyMenuSeed 由持有数据库凭据的运维命令调用，使用 IAM 的事务、目录和菜单校验。
func (r *Repository) ApplyMenuSeed(ctx context.Context, spec MenuSeedDefinition, confirmNew bool) (string, error) {
	if (spec.Domain != "business" && spec.Domain != "system") || !iamCodePattern.MatchString(spec.Name) || !iamCodePattern.MatchString(spec.RouteName) || spec.Title == "" || spec.Permission == "" {
		return "", iamInvalid("菜单种子定义无效")
	}
	hash := fmt.Sprintf("%x", md5.Sum([]byte("menu-seed:"+spec.Domain+":"+spec.Name)))
	key := "menu-seed-v3:" + hash
	path := "/business/" + spec.Name
	component := path + "/index"
	codes := []string{spec.Permission + ":query", spec.Permission + ":create", spec.Permission + ":update", spec.Permission + ":delete"}
	status := "菜单已安装，保留现有菜单及授权"
	err := r.withIAMLock(ctx, func(tx *gorm.DB) error {
		initialized, err := iamMarked(tx, iamBootstrapKey)
		if err != nil {
			return err
		}
		if !initialized {
			return iamInvalid("请先完成 IAM 初始化")
		}
		applied, err := iamMarked(tx, key)
		if err != nil || applied {
			return err
		}
		// 系统骨架自 rbac-v1 起已包含这些模块，即便管理员后来删除，也保留原决策。
		if isBootstrapMenu(spec) {
			return tx.Create(&model.IAMMigration{Key: key}).Error
		}
		menus, err := allMenus(tx)
		if err != nil {
			return err
		}
		for _, menu := range menus {
			if menu.Kind == model.MenuPage && textValue(menu.Component) == component && textValue(menu.Permission) == codes[0] {
				return tx.Create(&model.IAMMigration{Key: key}).Error
			}
		}
		for _, menu := range menus {
			if textValue(menu.Path) == path || textValue(menu.RouteName) == spec.RouteName || textValue(menu.Component) == component || containsCode(codes, textValue(menu.Permission)) {
				return iamConflict(fmt.Sprintf("菜单种子与现有菜单 ID %d 冲突", menu.ID))
			}
		}
		oldMarker, err := iamMarked(tx, "menu-seed:"+hash)
		if err != nil {
			return err
		}
		fresh, err := iamMarked(tx, menuSeedFreshKey)
		if err != nil {
			return err
		}
		if !confirmNew && (oldMarker || !fresh) {
			return iamConflict("无法确认旧库中该模块从未安装，已保留现有授权状态。确认是首次安装后使用 --new")
		}
		registration, registered := r.menuCatalog.Registrations[component]
		if !registered {
			return iamInvalid("请先在 Wire 注册该模块的菜单组件与权限")
		}
		for _, code := range codes {
			if _, registered := registration.Permissions[code]; !registered {
				return iamInvalid("菜单权限与已注册模块不一致")
			}
		}
		if len(menus)+len(codes) > maxIAMMenus {
			return iamConflict(fmt.Sprintf("安装后菜单将超过 %d 项上限", maxIAMMenus))
		}
		var admin model.SystemRole
		if err := tx.Where("code = ? AND enabled", model.RoleAdmin).First(&admin).Error; err != nil {
			return fmt.Errorf("查找启用的管理员角色：%w", err)
		}
		if err := upgradeIAMSequences(tx); err != nil {
			return err
		}
		icon := "lucide:folder"
		page := model.SystemMenu{
			Name: spec.Title, Kind: model.MenuPage, Enabled: true,
			RouteName: &spec.RouteName, Path: &path, Component: &component, Permission: &codes[0], Icon: &icon,
		}
		if err := r.createSeedMenus(tx, &page, codes[1:], admin.ID); err != nil {
			return err
		}
		if err := ensureManager(tx); err != nil {
			return err
		}
		if err := revokeRole(tx, admin.ID); err != nil {
			return err
		}
		if err := tx.Create(&model.IAMMigration{Key: key}).Error; err != nil {
			return err
		}
		status = "已安装菜单并授予管理员权限，请重新登录"
		return nil
	})
	return status, err
}

func isBootstrapMenu(spec MenuSeedDefinition) bool {
	switch spec.Domain {
	case "business":
		return spec.Name == "category"
	case "system":
		return spec.Name == "system-user" || spec.Name == "system-role" || spec.Name == "system-menu"
	}
	return false
}

func containsCode(codes []string, value string) bool {
	for _, code := range codes {
		if value == code {
			return true
		}
	}
	return false
}

func (r *Repository) createSeedMenus(tx *gorm.DB, page *model.SystemMenu, buttonCodes []string, adminID uint64) error {
	insert := func(item *model.SystemMenu) error {
		if err := r.validateMenu(tx, item); err != nil {
			return err
		}
		return persistSystemMenu(tx, item)
	}
	if err := insert(page); err != nil {
		return err
	}
	grants := []model.RoleMenu{{RoleID: adminID, MenuID: page.ID}}
	for index, label := range []string{"新增", "编辑", "删除"} {
		button := model.SystemMenu{Name: label, Kind: model.MenuButton, ParentId: float64(page.ID), Permission: &buttonCodes[index], Enabled: true}
		if err := insert(&button); err != nil {
			return err
		}
		grants = append(grants, model.RoleMenu{RoleID: adminID, MenuID: button.ID})
	}
	return tx.Create(&grants).Error
}
