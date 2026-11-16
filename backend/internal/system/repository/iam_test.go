package repository

import (
	"github.com/Edward-Jackie/lambo/backend/internal/system/model"
	"testing"
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
