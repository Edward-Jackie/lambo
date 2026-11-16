<script lang="ts" setup>
import type { SystemMenuApi } from '#/api/business/system-menu';

import { ref } from 'vue';

import { Page, useVbenModal } from '@vben/common-ui';
import { Plus } from '@vben/icons';

import { ElButton, ElMessage, ElMessageBox } from 'element-plus';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { deleteSystemMenu } from '#/api/business/system-menu';

import { getManagementOptions } from '#/api/business/management';
import { columns, getSearchSchema, permissions, title } from './data';
import RecordForm from './form.vue';

defineOptions({ name: 'BusinessSystemMenu' });

const deleting = ref(new Set<SystemMenuApi.ID>());
const [Modal, modalApi] = useVbenModal({ connectedComponent: RecordForm });
const [Grid, gridApi] = useVbenVxeGrid<SystemMenuApi.Row>({
  formOptions: {
    schema: getSearchSchema(),
    submitOnChange: false,
    submitOnEnter: true,
  },

  gridOptions: {
    columns,
    height: 'auto',
    pagerConfig: { enabled: false },
    treeConfig: {
      transform: true,
      rowField: 'id',
      parentField: 'parentId',
      expandAll: true,
    },
    rowConfig: { keyField: 'id' },
    toolbarConfig: { refresh: true, custom: true, zoom: true },
    proxyConfig: {
      ajax: {
        query: async (
          _params: unknown,
          values: Partial<SystemMenuApi.Query> = {},
        ) => {
          const { menus } = await getManagementOptions();
          const matches = menus.filter(
            (item) =>
              (!values.name || item.name.includes(values.name)) &&
              (!values.kind || item.kind === values.kind) &&
              (values.enabled === undefined || item.enabled === values.enabled),
          );
          const selected = new Set(matches.map((item) => item.id));
          for (const item of matches) {
            let parent = item.parentId;
            const seen = new Set<number>();
            while (parent && !seen.has(parent)) {
              seen.add(parent);
              selected.add(parent);
              parent = menus.find((menu) => menu.id === parent)?.parentId ?? 0;
            }
          }
          const items = menus.filter((item) => selected.has(item.id));
          return { items, total: items.length };
        },
      },
    },
  },
});

function openForm(mode: 'create' | 'detail' | 'edit', row?: SystemMenuApi.Row) {
  modalApi.setData({ mode, id: row?.id }).open();
}

async function onDelete(row: SystemMenuApi.Row) {
  if (deleting.value.has(row.id)) return;
  deleting.value.add(row.id);
  try {
    await ElMessageBox.confirm(
      '删除后无法恢复，确定删除这条记录吗？',
      '确认删除',
      {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning',
      },
    );
    await deleteSystemMenu(row.id);
    ElMessage.success('删除成功');
    await gridApi.reload();
  } catch {
    // 取消无需提示；接口错误由统一请求客户端显示。
  } finally {
    deleting.value.delete(row.id);
  }
}
</script>

<template>
  <Page auto-content-height>
    <Modal @success="gridApi.query()" />
    <Grid :table-title="title">
      <template #toolbar-tools>
        <ElButton
          v-access:code="[permissions.create]"
          type="primary"
          @click="openForm('create')"
        >
          <Plus class="mr-1 size-4" />新增
        </ElButton>
      </template>
      <template #actions="{ row }">
        <ElButton
          v-access:code="[permissions.query]"
          link
          type="primary"
          @click="openForm('detail', row)"
          >详情</ElButton
        >
        <ElButton
          v-access:code="[permissions.update]"
          link
          type="primary"
          @click="openForm('edit', row)"
          >编辑</ElButton
        >
        <ElButton
          v-access:code="[permissions.delete]"
          link
          type="danger"
          :loading="deleting.has(row.id)"
          @click="onDelete(row)"
          >删除</ElButton
        >
      </template>
    </Grid>
  </Page>
</template>
