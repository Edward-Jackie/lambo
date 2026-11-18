-- 由 Lambo 生成：分类管理 的菜单与按钮权限；可重复执行，不会重复插入。
-- 菜单真源是数据库；本文件只补齐生成模块缺失的菜单行，不改动已有内容。
-- page 承载 :query，按钮派生 :create/:update/:delete，并授权给 admin 角色。
-- 执行方式：lambo db seed category

INSERT INTO system_menu (name, kind, parent_id, route_name, path, component, permission, icon, sort, enabled, created_at, updated_at)
SELECT '分类管理', 'page', 0, 'BusinessCategory', '/business/category', '/business/category/index', 'catalog:category:query', 'lucide:folder', 0, true, now(), now()
WHERE NOT EXISTS (SELECT 1 FROM system_menu WHERE path = '/business/category');

INSERT INTO system_menu (name, kind, parent_id, permission, sort, enabled, created_at, updated_at)
SELECT '新增', 'button', (SELECT id FROM system_menu WHERE path = '/business/category' ORDER BY id LIMIT 1), 'catalog:category:create', 0, true, now(), now()
WHERE NOT EXISTS (SELECT 1 FROM system_menu WHERE permission = 'catalog:category:create')
  AND EXISTS (SELECT 1 FROM system_menu WHERE path = '/business/category');

INSERT INTO system_menu (name, kind, parent_id, permission, sort, enabled, created_at, updated_at)
SELECT '编辑', 'button', (SELECT id FROM system_menu WHERE path = '/business/category' ORDER BY id LIMIT 1), 'catalog:category:update', 0, true, now(), now()
WHERE NOT EXISTS (SELECT 1 FROM system_menu WHERE permission = 'catalog:category:update')
  AND EXISTS (SELECT 1 FROM system_menu WHERE path = '/business/category');

INSERT INTO system_menu (name, kind, parent_id, permission, sort, enabled, created_at, updated_at)
SELECT '删除', 'button', (SELECT id FROM system_menu WHERE path = '/business/category' ORDER BY id LIMIT 1), 'catalog:category:delete', 0, true, now(), now()
WHERE NOT EXISTS (SELECT 1 FROM system_menu WHERE permission = 'catalog:category:delete')
  AND EXISTS (SELECT 1 FROM system_menu WHERE path = '/business/category');

-- 授权给 admin；其他角色请在菜单管理页按需分配，避免生成物覆盖人工授权。
INSERT INTO lambo_role_menus (role_id, menu_id)
SELECT r.id, m.id
FROM system_role r
JOIN system_menu m ON m.permission IN ('catalog:category:query', 'catalog:category:create', 'catalog:category:update', 'catalog:category:delete')
WHERE r.code = 'admin'
  AND NOT EXISTS (SELECT 1 FROM lambo_role_menus x WHERE x.role_id = r.id AND x.menu_id = m.id);
