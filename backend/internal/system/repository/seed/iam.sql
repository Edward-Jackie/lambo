-- 系统初始化数据：角色与菜单骨架。
-- 只放稳定数据；按钮由 PermissionCatalog 在启动时派生，账号密码依赖运行时哈希，均不在此文件。
-- 目录和页面均使用数据库序列分配 ID，仅在首次初始化时执行。

INSERT INTO system_role (code, name, enabled, remark, created_at, updated_at) VALUES
('admin', '管理员', true, NULL, now(), now()),
('reader', '只读用户', true, NULL, now(), now());

INSERT INTO system_menu (id, name, kind, parent_id, route_name, path, component, permission, icon, sort, enabled, created_at, updated_at)
SELECT CASE WHEN menu.kind = 'directory' THEN directory.id ELSE nextval(pg_get_serial_sequence('system_menu', 'id')) END,
       menu.name, menu.kind, CASE WHEN menu.under_system THEN directory.id ELSE 0 END,
       menu.route_name, menu.path, menu.component, menu.permission, menu.icon, menu.sort, true, now(), now()
FROM (SELECT nextval(pg_get_serial_sequence('system_menu', 'id')) AS id) AS directory
CROSS JOIN (VALUES
    ('系统管理', 'directory', false, 'SystemManagement', '/business/system', NULL, NULL, 'lucide:settings', 20),
    ('分类管理', 'page', false, 'BusinessCategory', '/business/category', '/business/category/index', 'catalog:category:query', 'lucide:folder', 0),
    ('用户管理', 'page', true, 'BusinessSystemUser', '/business/system-user', '/business/system-user/index', 'system:user:query', 'lucide:users', 1),
    ('角色管理', 'page', true, 'BusinessSystemRole', '/business/system-role', '/business/system-role/index', 'system:role:query', 'lucide:shield', 2),
    ('菜单管理', 'page', true, 'BusinessSystemMenu', '/business/system-menu', '/business/system-menu/index', 'system:menu:query', 'lucide:menu', 3)
) AS menu(name, kind, under_system, route_name, path, component, permission, icon, sort);
