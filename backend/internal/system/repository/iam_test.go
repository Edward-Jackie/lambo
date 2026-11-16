package repository

import (
	"strings"
	"testing"

	"github.com/Edward-Jackie/lambo/backend/internal/system/model"
)

func TestActiveMenusRejectDisabledAncestorsAndCycles(t *testing.T) {
	items := []model.SystemMenu{{ID: 1, Kind: model.MenuDirectory, Enabled: true}, {ID: 2, ParentId: 1, Kind: model.MenuPage, Enabled: true}, {ID: 3, ParentId: 2, Kind: model.MenuButton, Enabled: true}}
	granted := map[uint64]bool{1: true, 2: true, 3: true}
	if got := activeMenus(items, granted); len(got) != 3 {
		t.Fatal("完整授权链应允许目录、页面和按钮")
	}
	items[0].Enabled = false
	if got := activeMenus(items, granted); len(got) != 0 {
		t.Fatal("上级停用后不能残留页面或按钮权限")
	}
	items[0].Enabled = true
	delete(granted, 1)
	if got := activeMenus(items, granted); len(got) != 0 {
		t.Fatal("缺少上级授权时不能绕过权限树")
	}
	granted[1] = true
	items[0].ParentId = 3
	if got := activeMenus(items, granted); len(got) != 0 {
		t.Fatal("循环不能授予权限或导致无限递归")
	}
}
func TestIAMRelationIDsRejectUnsafeAndDuplicateValues(t *testing.T) {
	for _, ids := range [][]uint64{{0}, {1, 1}, {9007199254740992}} {
		if _, err := uniqueIDs(ids); err == nil {
			t.Fatal("应拒绝非法关联 ID", ids)
		}
	}
	if ids, err := uniqueIDs([]uint64{1, 2}); err != nil || len(ids) != 2 {
		t.Fatal("合法多角色关联应保留")
	}
}

func TestSeedSQLSplitsStatementsAndMatchesPermissionCatalog(t *testing.T) {
	statements := splitSQL(iamSeedSQL)
	if len(statements) != 2 {
		t.Fatalf("初始化脚本应为角色与菜单两条语句，实际 %d 条", len(statements))
	}
	for _, statement := range statements {
		if strings.HasPrefix(statement, "--") || !strings.HasPrefix(strings.ToUpper(statement), "INSERT INTO") {
			t.Fatalf("初始化语句应去掉注释且为 INSERT：%s", statement)
		}
	}
	// 脚本里的页面权限码必须存在于权限目录，避免菜单引用了不存在的权限。
	catalog := model.PermissionCatalog()
	for _, page := range []string{"catalog:category:query", "system:user:query", "system:role:query", "system:menu:query"} {
		if !strings.Contains(iamSeedSQL, page) {
			t.Fatalf("初始化菜单缺少页面权限 %s", page)
		}
		if _, ok := catalog[page]; !ok {
			t.Fatalf("初始化菜单引用了权限目录中不存在的 %s", page)
		}
	}
}
