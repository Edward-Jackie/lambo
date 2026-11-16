<script lang="ts" setup>
import type { CategoryApi } from '#/api/business/category';

import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import { ElAlert, ElDescriptions, ElDescriptionsItem, ElMessage } from 'element-plus';

import { useVbenForm } from '#/adapter/form';
import { createCategory, getCategory, updateCategory } from '#/api/business/category';

import { fields, fieldText, getFormSchema, title } from './data';

// 此组件与页面共用一份字段定义；只读字段不进入提交体。
interface ModalData {
  id?: CategoryApi.ID;
  mode: 'create' | 'detail' | 'edit';
}

const emit = defineEmits<{ success: [] }>();
const record = ref<CategoryApi.Row>();
const mode = ref<ModalData['mode']>('create');
const loadError = ref(false);
const recordID = ref<CategoryApi.ID>();
const modalTitle = computed(() => `${mode.value === 'create' ? '新增' : mode.value === 'edit' ? '编辑' : '详情'} · ${title}`);
let requestSequence = 0;
let saving = false;

const [Form, formApi] = useVbenForm<CategoryApi.Payload>({
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
      if (mode.value !== 'detail') await formApi.reset();
      if (sequence !== requestSequence) return;
      if (mode.value !== 'create') {
        if (recordID.value === undefined) throw new Error('缺少记录 ID');
        const result = await getCategory(recordID.value);
        if (sequence !== requestSequence) return;
        record.value = result;
        if (mode.value === 'edit') {
          await formApi.setValues({
            name: result.name,
            status: result.status,
            sort: result.sort,
            remark: result.remark,
          });
        }
      }
      if (sequence === requestSequence) modalApi.setState({ confirmDisabled: false });
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
      const payload: CategoryApi.Payload = {
        name: values.name,
        status: values.status,
        sort: values.sort,
        remark: values.remark === '' ? undefined : values.remark ?? undefined,
      };
      if (mode.value === 'edit') {
        if (recordID.value === undefined) throw new Error('缺少记录 ID');
        await updateCategory(recordID.value, payload);
      } else {
        await createCategory(payload);
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
  <Modal :title="modalTitle" :show-confirm-button="mode !== 'detail'" class="w-full max-w-2xl">
    <ElAlert v-if="loadError" title="加载失败，请关闭后重试" type="error" :closable="false" />
    <ElDescriptions v-else-if="mode === 'detail'" :column="1" border>
      <ElDescriptionsItem v-for="field in fields" :key="field.name" :label="field.label">
        {{ fieldText(field.name, record?.[field.name]) }}
      </ElDescriptionsItem>
    </ElDescriptions>
    <Form v-else />
  </Modal>
</template>
