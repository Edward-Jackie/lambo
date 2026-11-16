-- 系统初始化数据：角色与菜单骨架。
-- 只放稳定数据；按钮由 PermissionCatalog 在启动时派生，账号密码依赖运行时哈希，均不在此文件。
-- 固定 ID 用于建立父子关系，仅在首次初始化（rbac-v1 标记之前）执行。

INSERT INTO system_role (id, code, name, enabled, remark, created_at, updated_at) VALUES
(1, 'admin', '管理员', true, NULL, now(), now()),
(2, 'reader', '只读用户', true, NULL, now(), now());

INSERT INTO system_menu (id, name, kind, parent_id, route_name, path, component, permission, icon, sort, enabled, created_at, updated_at) VALUES
(1, '系统管理', 'directory', 0, 'SystemManagement', '/business/system', NULL, NULL, 'lucide:settings', 20, true, now(), now()),
(2, '分类管理', 'page', 0, 'BusinessCategory', '/business/category', '/business/category/index', 'catalog:category:query', 'lucide:folder', 0, true, now(), now()),
(3, '用户管理', 'page', 1, 'BusinessSystemUser', '/business/system-user', '/business/system-user/index', 'system:user:query', 'lucide:users', 1, true, now(), now()),
(4, '角色管理', 'page', 1, 'BusinessSystemRole', '/business/system-role', '/business/system-role/index', 'system:role:query', 'lucide:shield', 2, true, now(), now()),
(5, '菜单管理', 'page', 1, 'BusinessSystemMenu', '/business/system-menu', '/business/system-menu/index', 'system:menu:query', 'lucide:menu', 3, true, now(), now());
