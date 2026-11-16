<script lang="ts" setup>
import type { SystemUserApi } from '#/api/business/system-user';

import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import {
  ElAlert,
  ElDescriptions,
  ElDescriptionsItem,
  ElMessage,
} from 'element-plus';

import { getManagementOptions } from '#/api/business/management';
import { z } from '#/adapter/form';
import { useVbenForm } from '#/adapter/form';
import {
  createSystemUser,
  getSystemUser,
  updateSystemUser,
} from '#/api/business/system-user';

import { fields, fieldText, getFormSchema, title } from './data';

// 此组件与页面共用一份字段定义；只读字段不进入提交体。
interface ModalData {
  id?: SystemUserApi.ID;
  mode: 'create' | 'detail' | 'edit';
}

const emit = defineEmits<{ success: [] }>();
const record = ref<SystemUserApi.Row>();
const mode = ref<ModalData['mode']>('create');
const loadError = ref(false);
const recordID = ref<SystemUserApi.ID>();
const modalTitle = computed(
  () =>
    `${mode.value === 'create' ? '新增' : mode.value === 'edit' ? '编辑' : '详情'} · ${title}`,
);
let requestSequence = 0;
let saving = false;

const [Form, formApi] = useVbenForm<SystemUserApi.Payload>({
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
        const schema = getFormSchema();
        schema.push({
          fieldName: 'roleIds',
          label: '所属角色',
          component: 'Select',
          defaultValue: [],
          componentProps: {
            multiple: true,
            filterable: true,
            options: options.roles.map((role) => ({
              label: role.name + (role.enabled ? '' : '（停用）'),
              value: role.id,
              disabled: (role.menuIds ?? []).some((id) => {
                const menu = options.menus.find((item) => item.id === id);
                return (
                  !!menu?.permission &&
                  !options.grantedCodes.includes(menu.permission)
                );
              }),
            })),
          },
          rules: z.array(z.number()),
        });
        if (mode.value === 'create')
          schema.push({
            fieldName: 'password',
            label: '初始密码',
            component: 'VbenInputPassword',
            componentProps: { autocomplete: 'new-password' },
            rules: z
              .string()
              .min(1, '请输入初始密码')
              .refine(
                (value) =>
                  new TextEncoder().encode(value).length >= 12 &&
                  new TextEncoder().encode(value).length <= 72,
                '密码必须为 12—72 字节',
              ),
          });
        formApi.setState({ schema });
        await formApi.reset();
      }
      if (sequence !== requestSequence) return;
      if (mode.value !== 'create') {
        if (recordID.value === undefined) throw new Error('缺少记录 ID');
        const result = await getSystemUser(recordID.value);
        if (sequence !== requestSequence) return;
        record.value = result;
        if (mode.value === 'edit') {
          await formApi.setValues({
            roleIds: result.roleIds,
            username: result.username,
            realName: result.realName,
            enabled: result.enabled,
            remark: result.remark,
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
      const payload: SystemUserApi.Payload = {
        roleIds: values.roleIds ?? [],
        ...(mode.value === 'create' ? { password: values.password } : {}),
        username: values.username,
        realName: values.realName,
        enabled: values.enabled,
        remark: values.remark === '' ? undefined : (values.remark ?? undefined),
      };
      if (mode.value === 'edit') {
        if (recordID.value === undefined) throw new Error('缺少记录 ID');
        await updateSystemUser(recordID.value, payload);
      } else {
        await createSystemUser(payload);
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
      <ElDescriptionsItem label="所属角色">{{
        record?.roleNames || '—'
      }}</ElDescriptionsItem>
    </ElDescriptions>
    <template v-else
      ><ElAlert
        class="mb-4"
        title="修改用户信息或角色后，该用户需重新登录；停用将立即阻止访问。"
        type="info"
        :closable="false" /><Form
    /></template>
  </Modal>
</template>
