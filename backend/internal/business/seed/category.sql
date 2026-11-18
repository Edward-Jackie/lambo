-- Lambo 菜单种子 v2
-- 首次执行只为新页面建立菜单及默认授权，已配置的菜单交由 IAM 管理。
-- 执行方式：lambo db seed category

-- 与系统初始化及 IAM 管理写操作共用事务锁。
SELECT pg_advisory_xact_lock(1279345218);

DO $seed$
BEGIN
    IF EXISTS (SELECT 1 FROM lambo_iam_migrations WHERE key = 'menu-seed:' || md5('menu-seed:business:category')) THEN
        RETURN;
    END IF;
    IF (SELECT count(*) FROM system_role WHERE code = 'admin') = 0 THEN
        RAISE EXCEPTION '请先初始化 IAM 管理员角色';
    END IF;
END;
$seed$;

WITH applied AS (
    INSERT INTO lambo_iam_migrations (key)
    VALUES ('menu-seed:' || md5('menu-seed:business:category'))
    ON CONFLICT (key) DO NOTHING
    RETURNING key
), page AS (
    INSERT INTO system_menu (name, kind, parent_id, route_name, path, component, permission, icon, sort, enabled, created_at, updated_at)
    SELECT '分类管理', 'page', 0, 'BusinessCategory', '/business/category', '/business/category/index', 'catalog:category:query', 'lucide:folder', 0, true, now(), now()
    FROM applied
    WHERE NOT EXISTS (SELECT 1 FROM system_menu WHERE path = '/business/category')
      AND NOT EXISTS (SELECT 1 FROM system_menu WHERE route_name = 'BusinessCategory')
      AND NOT EXISTS (SELECT 1 FROM system_menu WHERE component = '/business/category/index')
      AND NOT EXISTS (SELECT 1 FROM system_menu WHERE permission = 'catalog:category:query')
      AND NOT EXISTS (SELECT 1 FROM system_menu WHERE permission IN ('catalog:category:create', 'catalog:category:update', 'catalog:category:delete'))
    RETURNING id
), buttons AS (
    INSERT INTO system_menu (name, kind, parent_id, permission, sort, enabled, created_at, updated_at)
    SELECT button.name, 'button', page.id, button.permission, 0, true, now(), now()
    FROM page CROSS JOIN (VALUES
        ('新增', 'catalog:category:create'),
        ('编辑', 'catalog:category:update'),
        ('删除', 'catalog:category:delete')
    ) AS button(name, permission)
    RETURNING id
), granted AS (
    INSERT INTO lambo_role_menus (role_id, menu_id)
    SELECT r.id, m.id
    FROM system_role r
    CROSS JOIN (SELECT id FROM page UNION ALL SELECT id FROM buttons) m
    WHERE r.code = 'admin'
    RETURNING role_id
)
DELETE FROM lambo_sessions
WHERE account_id IN (
    SELECT account_id FROM lambo_account_roles WHERE role_id IN (SELECT role_id FROM granted)
);
