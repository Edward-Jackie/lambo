// 由 Lambo 生成的起始代码，可按业务修改；再次生成遇到差异会停止。
import type { VbenFormSchema } from '#/adapter/form';
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { CategoryApi } from '#/api/business/category';

import { z } from '#/adapter/form';

export const title = "分类管理";
export const permissions = {
  create: "catalog:category:create",
  update: "catalog:category:update",
  delete: "catalog:category:delete",
  query: "catalog:category:query",
};

export const fields = [
  { name: "name", label: "分类名称" },
  { name: "status", label: "状态" },
  { name: "sort", label: "排序" },
  { name: "remark", label: "备注" },
] as const;

const options: Record<string, { label: string; value: boolean | number | string }[]> = {
  status: [{"label":"启用","value":1},{"label":"停用","value":0}],
};

export function fieldText(name: string, value: unknown) {
  if (value === null || value === undefined || value === '') return '—';
  const option = options[name]?.find((item) => item.value === value);
  return option?.label ?? String(value);
}

export function getFormSchema(): VbenFormSchema<CategoryApi.Payload>[] {
  return [
    {
      fieldName: "name",
      label: "分类名称",
      component: "Input",
      defaultValue: "",
      componentProps: {
        maxlength: 64,
        showWordLimit: true,
        class: 'w-full',
      },
      rules: z.string().trim().min(1, "请输入分类名称").max(64),
    },
    {
      fieldName: "status",
      label: "状态",
      component: "Select",
      defaultValue: 1,
      componentProps: {
        options: options.status,
        class: 'w-full',
      },
      rules: z.number().finite().refine((value) => options.status?.some((item) => item.value === value), "请选择有效的状态"),
    },
    {
      fieldName: "sort",
      label: "排序",
      component: "InputNumber",
      defaultValue: 0,
      componentProps: {
        class: 'w-full',
      },
      rules: z.number().finite(),
    },
    {
      fieldName: "remark",
      label: "备注",
      component: "Input",
      defaultValue: "",
      componentProps: {
        maxlength: 500,
        showWordLimit: true,
        class: 'w-full',
      },
      rules: z.preprocess((value) => value === null || value === '' ? undefined : value, z.string().max(500).optional()),
    },
  ];
}

export function getSearchSchema(): VbenFormSchema[] {
  return [
    {
      fieldName: "name",
      label: "分类名称",
      component: "Input",
      componentProps: { clearable: true,  },
    },
    {
      fieldName: "status",
      label: "状态",
      component: "Select",
      componentProps: { clearable: true, options: options.status, },
    },
  ];
}

export const columns: VxeTableGridOptions<CategoryApi.Row>['columns'] = [
  { field: 'id', title: 'ID', minWidth: 100 },
  { field: "name", title: "分类名称", minWidth: 140, formatter: ({ cellValue }) => fieldText("name", cellValue) },
  { field: "status", title: "状态", minWidth: 140, formatter: ({ cellValue }) => fieldText("status", cellValue) },
  { field: "sort", title: "排序", minWidth: 140, formatter: ({ cellValue }) => fieldText("sort", cellValue) },
  { field: "remark", title: "备注", minWidth: 140, formatter: ({ cellValue }) => fieldText("remark", cellValue) },
  { title: '操作', width: 200, fixed: 'right', slots: { default: 'actions' } },
];
