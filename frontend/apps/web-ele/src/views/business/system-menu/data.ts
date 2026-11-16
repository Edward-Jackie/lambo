// 由 Lambo 生成的起始代码，可按业务修改；再次生成遇到差异会停止。
import type { VbenFormSchema } from '#/adapter/form';
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { SystemMenuApi } from '#/api/business/system-menu';

import { z } from '#/adapter/form';

export const title = '菜单管理';
export const permissions = {
  create: 'system:menu:create',
  update: 'system:menu:update',
  delete: 'system:menu:delete',
  query: 'system:menu:query',
};

export const fields = [
  { name: 'name', label: '菜单名称' },
  { name: 'kind', label: '类型' },
  { name: 'parentId', label: '上级菜单' },
  { name: 'routeName', label: '路由名称' },
  { name: 'path', label: '路由路径' },
  { name: 'component', label: '页面组件' },
  { name: 'permission', label: '权限码' },
  { name: 'icon', label: '图标' },
  { name: 'sort', label: '排序' },
  { name: 'enabled', label: '状态' },
] as const;

const options: Record<
  string,
  { label: string; value: boolean | number | string }[]
> = {
  kind: [
    { label: '目录', value: 'directory' },
    { label: '页面', value: 'page' },
    { label: '按钮', value: 'button' },
  ],
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

export function getFormSchema(): VbenFormSchema<SystemMenuApi.Payload>[] {
  return [
    {
      fieldName: 'name',
      label: '菜单名称',
      component: 'Input',
      defaultValue: '',
      componentProps: {
        maxlength: 64,
        showWordLimit: true,
        class: 'w-full',
      },
      rules: z.string().trim().min(1, '请输入菜单名称').max(64),
    },
    {
      fieldName: 'kind',
      label: '类型',
      component: 'Select',
      defaultValue: undefined,
      componentProps: {
        options: options.kind,
        class: 'w-full',
      },
      rules: z
        .string()
        .trim()
        .min(1, '请输入类型')
        .refine(
          (value) => options.kind?.some((item) => item.value === value),
          '请选择有效的类型',
        ),
    },
    {
      fieldName: 'parentId',
      label: '上级菜单',
      component: 'InputNumber',
      defaultValue: 0,
      componentProps: {
        class: 'w-full',
      },
      rules: z.number().finite(),
    },
    {
      fieldName: 'routeName',
      label: '路由名称',
      component: 'Input',
      defaultValue: '',
      componentProps: {
        maxlength: 64,
        showWordLimit: true,
        class: 'w-full',
      },
      rules: z.preprocess(
        (value) => (value === null || value === '' ? undefined : value),
        z.string().max(64).optional(),
      ),
    },
    {
      fieldName: 'path',
      label: '路由路径',
      component: 'Input',
      defaultValue: '',
      componentProps: {
        maxlength: 128,
        showWordLimit: true,
        class: 'w-full',
      },
      rules: z.preprocess(
        (value) => (value === null || value === '' ? undefined : value),
        z.string().max(128).optional(),
      ),
    },
    {
      fieldName: 'component',
      label: '页面组件',
      component: 'Input',
      defaultValue: '',
      componentProps: {
        maxlength: 128,
        showWordLimit: true,
        class: 'w-full',
      },
      rules: z.preprocess(
        (value) => (value === null || value === '' ? undefined : value),
        z.string().max(128).optional(),
      ),
    },
    {
      fieldName: 'permission',
      label: '权限码',
      component: 'Input',
      defaultValue: '',
      componentProps: {
        maxlength: 96,
        showWordLimit: true,
        class: 'w-full',
      },
      rules: z.preprocess(
        (value) => (value === null || value === '' ? undefined : value),
        z.string().max(96).optional(),
      ),
    },
    {
      fieldName: 'icon',
      label: '图标',
      component: 'Input',
      defaultValue: '',
      componentProps: {
        maxlength: 64,
        showWordLimit: true,
        class: 'w-full',
      },
      rules: z.preprocess(
        (value) => (value === null || value === '' ? undefined : value),
        z.string().max(64).optional(),
      ),
    },
    {
      fieldName: 'sort',
      label: '排序',
      component: 'InputNumber',
      defaultValue: 0,
      componentProps: {
        class: 'w-full',
      },
      rules: z.number().finite(),
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
  ];
}

export function getSearchSchema(): VbenFormSchema[] {
  return [
    {
      fieldName: 'name',
      label: '菜单名称',
      component: 'Input',
      componentProps: { clearable: true },
    },
    {
      fieldName: 'kind',
      label: '类型',
      component: 'Select',
      componentProps: { clearable: true, options: options.kind },
    },
    {
      fieldName: 'enabled',
      label: '状态',
      component: 'Select',
      componentProps: { clearable: true, options: options.enabled },
    },
  ];
}

export const columns: VxeTableGridOptions<SystemMenuApi.Row>['columns'] = [
  { field: 'name', title: '菜单名称', minWidth: 200, treeNode: true },
  {
    field: 'kind',
    title: '类型',
    width: 80,
    formatter: ({ cellValue }) => fieldText('kind', cellValue),
  },
  { field: 'path', title: '路由路径', minWidth: 180 },
  { field: 'permission', title: '权限码', minWidth: 200 },
  { field: 'sort', title: '排序', width: 70 },
  {
    field: 'enabled',
    title: '状态',
    width: 80,
    formatter: ({ cellValue }) => fieldText('enabled', cellValue),
  },
  { field: 'id', title: 'ID', visible: false },
  { field: 'parentId', title: '上级菜单', visible: false },
  { field: 'routeName', title: '路由名称', visible: false },
  { field: 'component', title: '页面组件', visible: false },
  { field: 'icon', title: '图标', visible: false },
  { title: '操作', width: 180, fixed: 'right', slots: { default: 'actions' } },
];
