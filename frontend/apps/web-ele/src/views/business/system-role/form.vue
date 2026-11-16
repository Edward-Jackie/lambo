<script lang="ts" setup>
import type { SystemRoleApi } from '#/api/business/system-role';

import { computed, nextTick, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';

import {
  ElAlert,
  ElDescriptions,
  ElDescriptionsItem,
  ElMessage,
  ElTree,
} from 'element-plus';

import { getManagementOptions, menuTree } from '#/api/business/management';
import type { MenuNode } from '#/api/business/management';
import { useVbenForm } from '#/adapter/form';
import {
  createSystemRole,
  getSystemRole,
  updateSystemRole,
} from '#/api/business/system-role';

import { fields, fieldText, getFormSchema, title } from './data';

// 此组件与页面共用一份字段定义；只读字段不进入提交体。
interface ModalData {
  id?: SystemRoleApi.ID;
  mode: 'create' | 'detail' | 'edit';
}

const emit = defineEmits<{ success: [] }>();
const record = ref<SystemRoleApi.Row>();
const mode = ref<ModalData['mode']>('create');
const loadError = ref(false);
const recordID = ref<SystemRoleApi.ID>();
const modalTitle = computed(
  () =>
    `${mode.value === 'create' ? '新增' : mode.value === 'edit' ? '编辑' : '详情'} · ${title}`,
);
let requestSequence = 0;
let saving = false;
const tree = ref<InstanceType<typeof ElTree>>();
const treeData = ref<MenuNode[]>([]);
const selectedMenus = ref<number[]>([]);
function flattenMenus(items: MenuNode[]): MenuNode[] {
  return items.flatMap((item) => [item, ...flattenMenus(item.children ?? [])]);
}
const detailTree = computed(() =>
  menuTree(
    flattenMenus(treeData.value).filter((item) =>
      selectedMenus.value.includes(item.id),
    ),
  ),
);

const [Form, formApi] = useVbenForm<SystemRoleApi.Payload>({
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
      const options = await getManagementOptions();
      if (sequence !== requestSequence) return;
      treeData.value = menuTree(
        options.menus.map((menu) => ({
          ...menu,
          disabled:
            !!menu.permission &&
            !options.grantedCodes.includes(menu.permission),
        })),
      );
      selectedMenus.value = [];
      // 详情模式不渲染编辑表单，不能等待它挂载后再加载记录。
      if (mode.value !== 'detail') await formApi.reset();
      if (sequence !== requestSequence) return;
      if (mode.value !== 'create') {
        if (recordID.value === undefined) throw new Error('缺少记录 ID');
        const result = await getSystemRole(recordID.value);
        if (sequence !== requestSequence) return;
        record.value = result;
        selectedMenus.value = result.menuIds;
        if (mode.value === 'edit') {
          await formApi.setValues({
            code: result.code,
            name: result.name,
            enabled: result.enabled,
            remark: result.remark,
          });
        }
      }
      await nextTick();
      tree.value?.setCheckedKeys(selectedMenus.value);
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
      const payload: SystemRoleApi.Payload = {
        menuIds: (tree.value?.getCheckedKeys(false) ?? []).map(Number),
        code: values.code,
        name: values.name,
        enabled: values.enabled,
        remark: values.remark === '' ? undefined : (values.remark ?? undefined),
      };
      if (mode.value === 'edit') {
        if (recordID.value === undefined) throw new Error('缺少记录 ID');
        await updateSystemRole(recordID.value, payload);
      } else {
        await createSystemRole(payload);
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
      <ElDescriptionsItem label="已授权菜单"
        >{{ treeData.length ? selectedMenus.length : 0 }} 项</ElDescriptionsItem
      >
    </ElDescriptions>
    <template v-else>
      <ElAlert
        class="mb-4"
        title="授权或启停变更后，关联用户需重新登录。勾选操作时会自动补齐上级目录和页面。"
        type="info"
        :closable="false"
      />
      <Form />
      <div class="mt-4 rounded border p-4">
        <div class="mb-3 font-medium">菜单与按钮权限</div>
        <ElTree
          ref="tree"
          :data="treeData"
          node-key="id"
          show-checkbox
          check-strictly
          default-expand-all
          :props="{ label: 'name', children: 'children' }"
          :default-checked-keys="selectedMenus"
          empty-text="暂无可授权菜单"
        />
      </div>
    </template>
    <ElTree
      v-if="mode === 'detail' && !loadError"
      class="mt-4"
      :data="detailTree"
      node-key="id"
      default-expand-all
      :props="{ label: 'name', children: 'children' }"
    />
  </Modal>
</template>
