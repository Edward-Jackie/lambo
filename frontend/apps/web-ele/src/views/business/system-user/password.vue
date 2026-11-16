<script setup lang="ts">
import { ref } from 'vue';
import { useVbenModal } from '@vben/common-ui';
import { ElAlert, ElMessage } from 'element-plus';
import { useVbenForm, z } from '#/adapter/form';
import { resetUserPassword } from '#/api/business/management';
const username = ref('');
const [Form, formApi] = useVbenForm<{ password: string }>({
  showDefaultActions: false,
  schema: [
    {
      fieldName: 'password',
      label: '新密码',
      component: 'VbenInputPassword',
      componentProps: { autocomplete: 'new-password' },
      rules: z
        .string()
        .min(1, '请输入新密码')
        .refine(
          (value) =>
            new TextEncoder().encode(value).length >= 12 &&
            new TextEncoder().encode(value).length <= 72,
          '密码必须为 12—72 字节',
        ),
    },
  ],
});
const [Modal, modalApi] = useVbenModal<{ id: number; username: string }>({
  onOpenChange: async (open) => {
    if (open) {
      username.value = modalApi.getData()?.username ?? '';
      await formApi.reset();
    }
  },
  onConfirm: async () => {
    modalApi.lock();
    try {
      const { valid } = await formApi.validate();
      if (!valid) return;
      const values = await formApi.getValues();
      await resetUserPassword(modalApi.getData()!.id, values.password);
      ElMessage.success('密码已重置，该用户需重新登录');
      modalApi.close();
    } catch {
      // 统一请求客户端显示错误，保留输入以便修改或重试。
    } finally {
      modalApi.unlock();
    }
  },
});
</script>
<template>
  <Modal :title="`重置密码 · ${username}`">
    <ElAlert
      class="mb-4"
      title="新密码保存后，该用户的所有登录会话立即失效。"
      type="warning"
      :closable="false"
    />
    <Form />
  </Modal>
</template>
