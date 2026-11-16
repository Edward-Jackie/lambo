<script lang="ts" setup>
import type { SystemMenuApi } from '#/api/business/system-menu';

import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import {
  ElAlert,
  ElDescriptions,
  ElDescriptionsItem,
  ElMessage,
} from 'element-plus';

import { getManagementOptions, menuTree } from '#/api/business/management';
import { useVbenForm } from '#/adapter/form';
import {
  createSystemMenu,
  getSystemMenu,
  updateSystemMenu,
} from '#/api/business/system-menu';

import { fields, fieldText, getFormSchema, title } from './data';

// 此组件与页面共用一份字段定义；只读字段不进入提交体。
interface ModalData {
  id?: SystemMenuApi.ID;
  mode: 'create' | 'detail' | 'edit';
}

const emit = defineEmits<{ success: [] }>();
const record = ref<SystemMenuApi.Row>();
const mode = ref<ModalData['mode']>('create');
const loadError = ref(false);
const recordID = ref<SystemMenuApi.ID>();
const modalTitle = computed(
  () =>
    `${mode.value === 'create' ? '新增' : mode.value === 'edit' ? '编辑' : '详情'} · ${title}`,
);
let requestSequence = 0;
let saving = false;

const [Form, formApi] = useVbenForm<SystemMenuApi.Payload>({
  schema: getFormSchema(),
  showDefaultActions: false,
  commonConfig: { componentProps: { class: 'w-full' } },
});

const [Modal, modalApi] = useVbenModal<ModalData>({
  onOpenChange: async (open) => {
    const sequence = ++requestSequence;
    if (!open) return;
    const data = modalApi.getData();
    mode.value = data?.mode ?? 'create';
    recordID.value = data?.id;
    record.value = undefined;
    loadError.value = false;
    modalApi.setState({ loading: true, confirmDisabled: true });
    try {
      // 详情模式不渲染编辑表单，不能等待它挂载后再加载记录。
      if (mode.value !== 'detail') {
        const options = await getManagementOptions();
        if (sequence !== requestSequence) return;
        const excluded = new Set<number>(
          recordID.value === undefined ? [] : [recordID.value],
        );
        let changed = true;
        while (changed) {
          changed = false;
          for (const menu of options.menus) {
            if (excluded.has(menu.parentId) && !excluded.has(menu.id)) {
              excluded.add(menu.id);
              changed = true;
            }
          }
        }
        const parentOptions = [
          {
            id: 0,
            name: '顶级菜单',
            children: menuTree(
              options.menus.filter(
                (menu) => menu.kind !== 'button' && !excluded.has(menu.id),
              ),
            ),
          },
        ];
        const schema = getFormSchema().map((field) => {
          if (field.fieldName === 'parentId')
            return {
              ...field,
              component: 'TreeSelect' as const,
              componentProps: {
                data: parentOptions,
                defaultExpandedKeys: [0],
                nodeKey: 'id',
                valueKey: 'id',
                checkStrictly: true,
                props: { label: 'name', value: 'id', children: 'children' },
                class: 'w-full',
              },
            };
          if (
            field.fieldName === 'component' ||
            field.fieldName === 'permission'
          ) {
            const source =
              field.fieldName === 'component'
                ? options.components
                : options.permissions;
            return {
              ...field,
              component: 'Select' as const,
              componentProps: {
                clearable: true,
                filterable: true,
                options: Object.entries(source).map(([value, label]) => ({
                  value,
                  label: label + ' · ' + value,
                })),
                class: 'w-full',
              },
            };
          }
          if (
            ['routeName', 'path', 'component', 'icon'].includes(
              field.fieldName ?? '',
            )
          )
            return {
              ...field,
              dependencies: {
                show: (values: Record<string, unknown>) =>
                  values.kind !== 'button',
                triggerFields: ['kind'],
              },
            };
          return field;
        });
        formApi.setState({ schema });
        await formApi.reset();
      }
      if (sequence !== requestSequence) return;
      if (mode.value !== 'create') {
        if (recordID.value === undefined) throw new Error('缺少记录 ID');
        const result = await getSystemMenu(recordID.value);
        if (sequence !== requestSequence) return;
        record.value = result;
        if (mode.value === 'edit') {
          await formApi.setValues({
            name: result.name,
            kind: result.kind,
            parentId: result.parentId,
            routeName: result.routeName,
            path: result.path,
            component: result.component,
            permission: result.permission,
            icon: result.icon,
            sort: result.sort,
            enabled: result.enabled,
          });
        }
      }
      if (sequence === requestSequence)
        modalApi.setState({ confirmDisabled: false });
    } catch {
      if (sequence === requestSequence) loadError.value = true;
    } finally {
      if (sequence === requestSequence) modalApi.setState({ loading: false });
    }
  },
  onConfirm: async () => {
    if (loadError.value || mode.value === 'detail' || saving) return;
    saving = true;
    modalApi.lock();
    try {
      const { valid } = await formApi.validate();
      if (!valid) return;
      const values = await formApi.getValues();
      const payload: SystemMenuApi.Payload = {
        name: values.name,
        kind: values.kind,
        parentId: values.parentId,
        routeName:
          values.kind === 'button'
            ? undefined
            : values.routeName === ''
              ? undefined
              : (values.routeName ?? undefined),
        path:
          values.kind === 'button'
            ? undefined
            : values.path === ''
              ? undefined
              : (values.path ?? undefined),
        component:
          values.kind !== 'page'
            ? undefined
            : values.component === ''
              ? undefined
              : (values.component ?? undefined),
        permission:
          values.kind === 'directory'
            ? undefined
            : values.permission === ''
              ? undefined
              : (values.permission ?? undefined),
        icon:
          values.kind === 'button'
            ? undefined
            : values.icon === ''
              ? undefined
              : (values.icon ?? undefined),
        sort: values.sort,
        enabled: values.enabled,
      };
      if (mode.value === 'edit') {
        if (recordID.value === undefined) throw new Error('缺少记录 ID');
        await updateSystemMenu(recordID.value, payload);
      } else {
        await createSystemMenu(payload);
      }
      ElMessage.success('保存成功');
      modalApi.close();
      emit('success');
    } catch {
      // 请求客户端已显示错误，保留表单供用户修改或重试。
    } finally {
      saving = false;
      modalApi.unlock();
    }
  },
});
</script>

<template>
  <Modal
    :title="modalTitle"
    :show-confirm-button="mode !== 'detail'"
    class="w-full max-w-2xl"
  >
    <ElAlert
      v-if="loadError"
      title="加载失败，请关闭后重试"
      type="error"
      :closable="false"
    />
    <ElDescriptions v-else-if="mode === 'detail'" :column="1" border>
      <ElDescriptionsItem
        v-for="field in fields"
        :key="field.name"
        :label="field.label"
      >
        {{ fieldText(field.name, record?.[field.name]) }}
      </ElDescriptionsItem>
    </ElDescriptions>
    <template v-else
      ><ElAlert
        class="mb-4"
        title="编辑菜单后，已登录用户需重新登录以刷新导航与按钮。新增菜单需在角色中授权。"
        type="info"
        :closable="false" /><Form
    /></template>
  </Modal>
</template>
