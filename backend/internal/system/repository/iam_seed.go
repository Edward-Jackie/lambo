package repository

import (
	"context"
	_ "embed"
	"fmt"
	"strings"

	"github.com/Edward-Jackie/lambo/backend/internal/system/model"
	"gorm.io/gorm"
)

//go:embed seed/iam.sql
var iamSeedSQL string

// InitializeIAM 只初始化一次基础角色与菜单，不在后续启动时恢复人工撤销的权限。
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
		if err := execSeedSQL(tx, iamSeedSQL); err != nil {
			return err
		}
		var roles []model.SystemRole
		if err := tx.Find(&roles).Error; err != nil {
			return err
		}
		roleIDs := map[string]uint64{}
		for _, role := range roles {
			roleIDs[role.Code] = role.ID
		}
		adminID, readerID := roleIDs[model.RoleAdmin], roleIDs[model.RoleReader]
		if adminID == 0 || readerID == 0 {
			return iamInvalid("初始化数据缺少管理员或只读角色")
		}
		// 按钮由页面权限前缀从权限目录派生，菜单骨架与权限码只维护一份。
		var pages []model.SystemMenu
		if err := tx.Where(&model.SystemMenu{Kind: model.MenuPage}).Order("id").Find(&pages).Error; err != nil {
			return err
		}
		for _, page := range pages {
			prefix := strings.TrimSuffix(textValue(page.Permission), ":query")
			for code, label := range model.PermissionCatalog() {
				if !strings.HasPrefix(code, prefix+":") || strings.HasSuffix(code, ":query") {
					continue
				}
				permission := code
				button := model.SystemMenu{Name: label, Kind: model.MenuButton, ParentId: float64(page.ID), Permission: &permission, Enabled: true}
				if err := tx.Create(&button).Error; err != nil {
					return err
				}
			}
		}
		var menus []model.SystemMenu
		if err := tx.Find(&menus).Error; err != nil {
			return err
		}
		for _, menu := range menus {
			if err := tx.Create(&model.RoleMenu{RoleID: adminID, MenuID: menu.ID}).Error; err != nil {
				return err
			}
		}
		for _, page := range pages {
			// 不挂在系统管理下的页面是业务页面，只读角色默认可查询。
			if uint64(page.ParentId) != 0 {
				continue
			}
			if err := tx.Create(&model.RoleMenu{RoleID: readerID, MenuID: page.ID}).Error; err != nil {
				return err
			}
		}
		for _, a := range accounts {
			var rid uint64
			switch a.Role {
			case model.RoleAdmin:
				rid = adminID
			case model.RoleReader:
				rid = readerID
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

// execSeedSQL 逐条执行初始化脚本，跳过空行与注释。
func execSeedSQL(tx *gorm.DB, script string) error {
	for _, statement := range splitSQL(script) {
		if err := tx.Exec(statement).Error; err != nil {
			return fmt.Errorf("执行初始化脚本：%w", err)
		}
	}
	return nil
}

// splitSQL 按分号拆分语句，丢弃注释与空行。
func splitSQL(script string) []string {
	statements := make([]string, 0, 4)
	for _, statement := range strings.Split(script, ";") {
		lines := make([]string, 0, 8)
		for _, line := range strings.Split(statement, "\n") {
			if trimmed := strings.TrimSpace(line); trimmed != "" && !strings.HasPrefix(trimmed, "--") {
				lines = append(lines, line)
			}
		}
		if len(lines) > 0 {
			statements = append(statements, strings.Join(lines, "\n"))
		}
	}
	return statements
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
