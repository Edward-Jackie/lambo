// 由 Lambo 生成的起始代码，可按业务修改；再次生成遇到差异会停止。
import type { VbenFormSchema } from '#/adapter/form';
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { {{.StructName}}Api } from '#/api/business/{{.Name}}';

import { z } from '#/adapter/form';

export const title = {{json .Title}};
export const permissions = {
  create: {{json (printf "%s:create" .Permission)}},
  update: {{json (printf "%s:update" .Permission)}},
  delete: {{json (printf "%s:delete" .Permission)}},
  query: {{json (printf "%s:query" .Permission)}},
};

export const fields = [
{{range .Fields}}  { name: {{json .Name}}, label: {{json .Label}} },
{{end}}] as const;

const options: Record<string, { label: string; value: boolean | number | string }[]> = {
{{range .Fields}}{{if .Choices}}  {{.Name}}: {{json .Choices}},
{{end}}{{end}}};

export function fieldText(name: string, value: unknown) {
  if (value === null || value === undefined || value === '') return '—';
  const option = options[name]?.find((item) => item.value === value);
  return option?.label ?? String(value);
}

export function getFormSchema(): VbenFormSchema<{{.StructName}}Api.Payload>[] {
  return [
{{range .EditableFields}}    {
      fieldName: {{json .Name}},
      label: {{json .Label}},
      component: {{json .Component}},
      defaultValue: {{.DefaultJSON}},
      componentProps: {
{{if .Choices}}        options: options.{{.Name}},
{{end}}{{if .MaxLength}}        maxlength: {{.MaxLength}},
        showWordLimit: true,
{{end}}        class: 'w-full',
      },
      rules: {{if not .Required}}z.preprocess((value) => value === null || value === '' ? undefined : value, {{end}}z.{{if eq .Type "string"}}string(){{if .Required}}.trim().min(1, {{json (printf "请输入%s" .Label)}}){{end}}{{if .MaxLength}}.max({{.MaxLength}}){{end}}{{else if eq .Type "number"}}number().finite(){{else}}boolean(){{end}}{{if .Choices}}.refine((value) => options.{{.Name}}?.some((item) => item.value === value), {{json (printf "请选择有效的%s" .Label)}}){{end}}{{if not .Required}}.optional()){{end}},
    },
{{end}}  ];
}

export function getSearchSchema(): VbenFormSchema[] {
  return [
{{range .SearchFields}}    {
      fieldName: {{json .Name}},
      label: {{json .Label}},
      component: {{json .Component}},
      componentProps: { clearable: true, {{if .Choices}}options: options.{{.Name}},{{end}} },
    },
{{end}}  ];
}

export const columns: VxeTableGridOptions<{{.StructName}}Api.Row>['columns'] = [
  { field: 'id', title: 'ID', minWidth: 100 },
{{range .Fields}}  { field: {{json .Name}}, title: {{json .Label}}, minWidth: 140, formatter: ({ cellValue }) => fieldText({{json .Name}}, cellValue) },
{{end}}  { title: '操作', width: 200, fixed: 'right', slots: { default: 'actions' } },
];
