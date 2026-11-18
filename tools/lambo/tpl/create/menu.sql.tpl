-- Lambo 菜单种子 v2
-- 首次执行只为新页面建立菜单及默认授权，已配置的菜单交由 IAM 管理。
-- 执行方式：lambo db seed {{.Name}}
{{- $marker := printf "menu-seed:%s:%s" .Domain .Name }}
{{- $path := printf "/business/%s" .Name }}
{{- $route := printf "Business%s" .StructName }}
{{- $component := printf "/business/%s/index" .Name }}
{{- $query := printf "%s:query" .Permission }}
{{- $create := printf "%s:create" .Permission }}
{{- $update := printf "%s:update" .Permission }}
{{- $delete := printf "%s:delete" .Permission }}

-- 与系统初始化及 IAM 管理写操作共用事务锁。
SELECT pg_advisory_xact_lock(1279345218);

DO $seed$
BEGIN
    IF EXISTS (SELECT 1 FROM lambo_iam_migrations WHERE key = 'menu-seed:' || md5({{sql $marker}})) THEN
        RETURN;
    END IF;
    IF (SELECT count(*) FROM system_role WHERE code = 'admin') = 0 THEN
        RAISE EXCEPTION '请先初始化 IAM 管理员角色';
    END IF;
END;
$seed$;

WITH applied AS (
    INSERT INTO lambo_iam_migrations (key)
    VALUES ('menu-seed:' || md5({{sql $marker}}))
    ON CONFLICT (key) DO NOTHING
    RETURNING key
), page AS (
    INSERT INTO system_menu (name, kind, parent_id, route_name, path, component, permission, icon, sort, enabled, created_at, updated_at)
    SELECT {{sql .Title}}, 'page', 0, {{sql $route}}, {{sql $path}}, {{sql $component}}, {{sql $query}}, 'lucide:folder', 0, true, now(), now()
    FROM applied
    WHERE NOT EXISTS (SELECT 1 FROM system_menu WHERE path = {{sql $path}})
      AND NOT EXISTS (SELECT 1 FROM system_menu WHERE route_name = {{sql $route}})
      AND NOT EXISTS (SELECT 1 FROM system_menu WHERE component = {{sql $component}})
      AND NOT EXISTS (SELECT 1 FROM system_menu WHERE permission = {{sql $query}})
      AND NOT EXISTS (SELECT 1 FROM system_menu WHERE permission IN ({{sql $create}}, {{sql $update}}, {{sql $delete}}))
    RETURNING id
), buttons AS (
    INSERT INTO system_menu (name, kind, parent_id, permission, sort, enabled, created_at, updated_at)
    SELECT button.name, 'button', page.id, button.permission, 0, true, now(), now()
    FROM page CROSS JOIN (VALUES
        ('新增', {{sql $create}}),
        ('编辑', {{sql $update}}),
        ('删除', {{sql $delete}})
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
