package model

import (
	"context"
	"errors"
)

var (
	ErrIAMInvalid   = errors.New("权限配置无效")
	ErrIAMConflict  = errors.New("配置冲突")
	ErrIAMForbidden = errors.New("不能管理或分配自身未拥有的权限")
)

const (
	PermissionUserQuery    = "system:user:query"
	PermissionUserCreate   = "system:user:create"
	PermissionUserUpdate   = "system:user:update"
	PermissionUserDelete   = "system:user:delete"
	PermissionUserPassword = "system:user:password"
	PermissionRoleQuery    = "system:role:query"
	PermissionRoleCreate   = "system:role:create"
	PermissionRoleUpdate   = "system:role:update"
	PermissionRoleDelete   = "system:role:delete"
	PermissionMenuQuery    = "system:menu:query"
	PermissionMenuCreate   = "system:menu:create"
	PermissionMenuUpdate   = "system:menu:update"
	PermissionMenuDelete   = "system:menu:delete"
	MenuDirectory          = "directory"
	MenuPage               = "page"
	MenuButton             = "button"
)

// 权限目录只声明接口需求，授权关系始终来自数据库。
func PermissionCatalog() map[string]string {
	return map[string]string{
		"catalog:category:query": "分类查询", "catalog:category:create": "分类新增", "catalog:category:update": "分类编辑", "catalog:category:delete": "分类删除",
		PermissionUserQuery: "用户查询", PermissionUserCreate: "用户新增", PermissionUserUpdate: "用户编辑与角色分配", PermissionUserDelete: "用户删除", PermissionUserPassword: "用户密码重置",
		PermissionRoleQuery: "角色查询", PermissionRoleCreate: "角色新增", PermissionRoleUpdate: "角色编辑与授权", PermissionRoleDelete: "角色删除",
		PermissionMenuQuery: "菜单查询", PermissionMenuCreate: "菜单新增", PermissionMenuUpdate: "菜单编辑", PermissionMenuDelete: "菜单删除",
	}
}

func PageComponents() map[string]string {
	return map[string]string{"/business/category/index": "分类管理", "/business/system-user/index": "用户管理", "/business/system-role/index": "角色管理", "/business/system-menu/index": "菜单管理"}
}

func ManagementCodes() []string {
	return []string{PermissionUserQuery, PermissionUserCreate, PermissionUserUpdate, PermissionUserDelete, PermissionUserPassword, PermissionRoleQuery, PermissionRoleCreate, PermissionRoleUpdate, PermissionRoleDelete, PermissionMenuQuery, PermissionMenuCreate, PermissionMenuUpdate, PermissionMenuDelete}
}

type AccountRole struct {
	AccountID uint64 `gorm:"primaryKey"`
	RoleID    uint64 `gorm:"primaryKey;index"`
}

func (AccountRole) TableName() string { return "lambo_account_roles" }

type RoleMenu struct {
	RoleID uint64 `gorm:"primaryKey"`
	MenuID uint64 `gorm:"primaryKey;index"`
}

func (RoleMenu) TableName() string { return "lambo_role_menus" }

// 迁移标记保证旧单角色只转换一次，避免启动时恢复已撤销的授权。
type IAMMigration struct {
	Key string `gorm:"primaryKey;size:64"`
}

func (IAMMigration) TableName() string { return "lambo_iam_migrations" }

type actorKey struct{}

func WithActor(ctx context.Context, id uint64) context.Context {
	return context.WithValue(ctx, actorKey{}, id)
}
func ActorID(ctx context.Context) uint64 { id, _ := ctx.Value(actorKey{}).(uint64); return id }

type permissionKey struct{}

func WithRequiredCode(ctx context.Context, code string) context.Context {
	return context.WithValue(ctx, permissionKey{}, code)
}
func RequiredCode(ctx context.Context) string {
	code, _ := ctx.Value(permissionKey{}).(string)
	return code
}
