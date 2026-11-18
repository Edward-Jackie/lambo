-- Lambo 基础数据初始化（PostgreSQL）。
-- 在 GORM 建表后执行：psql -X -v ON_ERROR_STOP=1 -d 数据库名 -f docs/init.sql
-- 此脚本可以先于或晚于服务的首次 IAM 初始化执行。
-- 分类页面也属于初始菜单。按钮、角色授权和账号绑定由首次 IAM 初始化按权限目录生成。
-- 示例账号及密码从服务的 bootstrap 配置读取。

BEGIN;
SELECT pg_advisory_xact_lock(1279345218);

DO $init$
BEGIN
    IF EXISTS (
        SELECT 1 FROM lambo_iam_migrations
        WHERE key IN ('rbac-base-sql-v1', 'rbac-v1')
    ) THEN
        RETURN;
    END IF;

    IF EXISTS (SELECT 1 FROM system_role) OR EXISTS (SELECT 1 FROM system_menu) THEN
        RAISE EXCEPTION '检测到已有角色或菜单但缺少初始化标记，请核对目标库';
    END IF;

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

    INSERT INTO lambo_iam_migrations (key) VALUES ('rbac-base-sql-v1');
END;
$init$;

COMMIT;
