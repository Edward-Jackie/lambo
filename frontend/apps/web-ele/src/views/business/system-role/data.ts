// 由 Lambo 生成的起始代码，可按业务修改；再次生成遇到差异会停止。
import type { VbenFormSchema } from '#/adapter/form';
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { SystemRoleApi } from '#/api/business/system-role';

import { z } from '#/adapter/form';

export const title = '角色管理';
export const permissions = {
  create: 'system:role:create',
  update: 'system:role:update',
  delete: 'system:role:delete',
  query: 'system:role:query',
};

export const fields = [
  { name: 'code', label: '角色编码' },
  { name: 'name', label: '角色名称' },
  { name: 'enabled', label: '状态' },
  { name: 'remark', label: '备注' },
] as const;

const options: Record<
  string,
  { label: string; value: boolean | number | string }[]
> = {
  enabled: [
    { label: '启用', value: true },
    { label: '停用', value: false },
  ],
};

export function fieldText(name: string, value: unknown) {
  if (value === null || value === undefined || value === '') return '—';
  const option = options[name]?.find((item) => item.value === value);
  return option?.label ?? String(value);
}

export function getFormSchema(): VbenFormSchema<SystemRoleApi.Payload>[] {
  return [
    {
      fieldName: 'code',
      label: '角色编码',
      component: 'Input',
      defaultValue: '',
      componentProps: {
        maxlength: 64,
        showWordLimit: true,
        class: 'w-full',
      },
      rules: z.string().trim().min(1, '请输入角色编码').max(64),
    },
    {
      fieldName: 'name',
      label: '角色名称',
      component: 'Input',
      defaultValue: '',
      componentProps: {
        maxlength: 64,
        showWordLimit: true,
        class: 'w-full',
      },
      rules: z.string().trim().min(1, '请输入角色名称').max(64),
    },
    {
      fieldName: 'enabled',
      label: '状态',
      component: 'Select',
      defaultValue: true,
      componentProps: {
        options: options.enabled,
        class: 'w-full',
      },
      rules: z
        .boolean()
        .refine(
          (value) => options.enabled?.some((item) => item.value === value),
          '请选择有效的状态',
        ),
    },
    {
      fieldName: 'remark',
      label: '备注',
      component: 'Input',
      defaultValue: '',
      componentProps: {
        maxlength: 500,
        showWordLimit: true,
        class: 'w-full',
      },
      rules: z.preprocess(
        (value) => (value === null || value === '' ? undefined : value),
        z.string().max(500).optional(),
      ),
    },
  ];
}

export function getSearchSchema(): VbenFormSchema[] {
  return [
    {
      fieldName: 'code',
      label: '角色编码',
      component: 'Input',
      componentProps: { clearable: true },
    },
    {
      fieldName: 'name',
      label: '角色名称',
      component: 'Input',
      componentProps: { clearable: true },
    },
    {
      fieldName: 'enabled',
      label: '状态',
      component: 'Select',
      componentProps: { clearable: true, options: options.enabled },
    },
  ];
}

export const columns: VxeTableGridOptions<SystemRoleApi.Row>['columns'] = [
  { field: 'id', title: 'ID', minWidth: 100 },
  {
    field: 'code',
    title: '角色编码',
    minWidth: 140,
    formatter: ({ cellValue }) => fieldText('code', cellValue),
  },
  {
    field: 'name',
    title: '角色名称',
    minWidth: 140,
    formatter: ({ cellValue }) => fieldText('name', cellValue),
  },
  {
    field: 'enabled',
    title: '状态',
    minWidth: 140,
    formatter: ({ cellValue }) => fieldText('enabled', cellValue),
  },
  {
    field: 'remark',
    title: '备注',
    minWidth: 140,
    formatter: ({ cellValue }) => fieldText('remark', cellValue),
  },
  { title: '操作', width: 200, fixed: 'right', slots: { default: 'actions' } },
];
