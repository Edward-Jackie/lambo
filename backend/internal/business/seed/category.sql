-- Lambo 菜单种子 v3
-- Lambo 菜单定义 {"domain":"business","name":"category","permission":"catalog:category","routeName":"BusinessCategory","title":"分类管理"}
-- 后端只读取菜单定义，由 IAM 校验后落库；下方 SQL 用于预览。
-- 执行方式：lambo db seed category

DO $seed$
BEGIN
    RAISE EXCEPTION '请通过 lambo db seed 执行 IAM 校验';
END;
$seed$;

WITH applied AS (
    INSERT INTO lambo_iam_migrations (key)
    VALUES ('menu-seed-v3:' || md5('menu-seed:business:category'))
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
SELECT count(*) FROM granted;
