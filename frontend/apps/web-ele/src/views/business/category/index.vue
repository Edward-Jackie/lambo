<script lang="ts" setup>
import type { CategoryApi } from '#/api/business/category';

import { ref } from 'vue';

import { Page, useVbenModal } from '@vben/common-ui';
import { Plus } from '@vben/icons';

import { ElButton, ElMessage, ElMessageBox } from 'element-plus';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { deleteCategory, getCategoryList } from '#/api/business/category';

import { columns, getSearchSchema, permissions, title } from './data';
import RecordForm from './form.vue';

defineOptions({ name: 'BusinessCategory' });

const deleting = ref(new Set<CategoryApi.ID>());
const [Modal, modalApi] = useVbenModal({ connectedComponent: RecordForm });
const [Grid, gridApi] = useVbenVxeGrid<CategoryApi.Row>({
  formOptions: { schema: getSearchSchema(), submitOnChange: false, submitOnEnter: true },

  gridOptions: {
    columns,
    height: 'auto',
    pagerConfig: { pageSize: 20 },
    rowConfig: { keyField: 'id' },
    toolbarConfig: { refresh: true, custom: true, zoom: true },
    proxyConfig: {
      ajax: {
        query: async ({ page }: { page: { currentPage: number; pageSize: number } }, values: Partial<CategoryApi.Query> = {}) => getCategoryList({
          page: page.currentPage,
          pageSize: page.pageSize,
          name: values.name === '' ? undefined : values.name ?? undefined,
          status: values.status ?? undefined,
        }),
      },
    },
  },
});

function openForm(mode: 'create' | 'detail' | 'edit', row?: CategoryApi.Row) {
  modalApi.setData({ mode, id: row?.id }).open();
}

async function onDelete(row: CategoryApi.Row) {
  if (deleting.value.has(row.id)) return;
  deleting.value.add(row.id);
  try {
    await ElMessageBox.confirm('删除后无法恢复，确定删除这条记录吗？', '确认删除', {
      confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning',
    });
    await deleteCategory(row.id);
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
        <ElButton v-access:code="[permissions.create]" type="primary" @click="openForm('create')">
          <Plus class="mr-1 size-4" />新增
        </ElButton>
      </template>
      <template #actions="{ row }">
        <ElButton v-access:code="[permissions.query]" link type="primary" @click="openForm('detail', row)">详情</ElButton>
        <ElButton v-access:code="[permissions.update]" link type="primary" @click="openForm('edit', row)">编辑</ElButton>
        <ElButton v-access:code="[permissions.delete]" link type="danger" :loading="deleting.has(row.id)" @click="onDelete(row)">删除</ElButton>
      </template>
    </Grid>
  </Page>
</template>
