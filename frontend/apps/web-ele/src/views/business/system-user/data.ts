// 由 Lambo 生成的起始代码，可按业务修改；再次生成遇到差异会停止。
import type { VbenFormSchema } from '#/adapter/form';
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { SystemUserApi } from '#/api/business/system-user';

import { z } from '#/adapter/form';

export const title = '用户管理';
export const permissions = {
  create: 'system:user:create',
  update: 'system:user:update',
  delete: 'system:user:delete',
  query: 'system:user:query',
};

export const fields = [
  { name: 'username', label: '用户名' },
  { name: 'realName', label: '姓名' },
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

export function getFormSchema(): VbenFormSchema<SystemUserApi.Payload>[] {
  return [
    {
      fieldName: 'username',
      label: '用户名',
      component: 'Input',
      defaultValue: '',
      componentProps: {
        maxlength: 64,
        showWordLimit: true,
        class: 'w-full',
      },
      rules: z.string().trim().min(1, '请输入用户名').max(64),
    },
    {
      fieldName: 'realName',
      label: '姓名',
      component: 'Input',
      defaultValue: '',
      componentProps: {
        maxlength: 64,
        showWordLimit: true,
        class: 'w-full',
      },
      rules: z.string().trim().min(1, '请输入姓名').max(64),
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
      fieldName: 'username',
      label: '用户名',
      component: 'Input',
      componentProps: { clearable: true },
    },
    {
      fieldName: 'realName',
      label: '姓名',
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

export const columns: VxeTableGridOptions<SystemUserApi.Row>['columns'] = [
  { field: 'id', title: 'ID', width: 70 },
  { field: 'username', title: '用户名', minWidth: 130 },
  { field: 'realName', title: '姓名', minWidth: 150 },
  {
    field: 'enabled',
    title: '状态',
    width: 80,
    formatter: ({ cellValue }) => fieldText('enabled', cellValue),
  },
  { field: 'roleNames', title: '所属角色', minWidth: 180 },
  { field: 'remark', title: '备注', minWidth: 140 },
  { title: '操作', width: 240, fixed: 'right', slots: { default: 'actions' } },
];
