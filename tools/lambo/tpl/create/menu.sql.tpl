-- Lambo 菜单种子 v3
-- Lambo 菜单定义 {{json .MenuSeedMetadata}}
-- 后端只读取菜单定义，由 IAM 校验后落库；下方 SQL 用于预览。
-- 执行方式：lambo db seed {{.Name}}
{{- $marker := printf "menu-seed:%s:%s" .Domain .Name }}
{{- $path := printf "/business/%s" .Name }}
{{- $route := printf "Business%s" .StructName }}
{{- $component := printf "/business/%s/index" .Name }}
{{- $query := printf "%s:query" .Permission }}
{{- $create := printf "%s:create" .Permission }}
{{- $update := printf "%s:update" .Permission }}
{{- $delete := printf "%s:delete" .Permission }}

DO $seed$
BEGIN
    RAISE EXCEPTION '请通过 lambo db seed 执行 IAM 校验';
END;
$seed$;

WITH applied AS (
    INSERT INTO lambo_iam_migrations (key)
    VALUES ('menu-seed-v3:' || md5({{sql $marker}}))
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
SELECT count(*) FROM granted;
