<script lang="ts" setup>
import type { {{.StructName}}Api } from '#/api/business/{{.Name}}';

import { ref } from 'vue';

import { Page, useVbenModal } from '@vben/common-ui';
import { Plus } from '@vben/icons';

import { ElButton, ElMessage, ElMessageBox } from 'element-plus';

import { useVbenVxeGrid } from '#/adapter/vxe-table';
import { delete{{.StructName}}, get{{.StructName}}List } from '#/api/business/{{.Name}}';

import { columns, {{if .SearchFields}}getSearchSchema, {{end}}permissions, title } from './data';
import RecordForm from './form.vue';

defineOptions({ name: 'Business{{.StructName}}' });

const deleting = ref(new Set<{{.StructName}}Api.ID>());
const [Modal, modalApi] = useVbenModal({ connectedComponent: RecordForm });
const [Grid, gridApi] = useVbenVxeGrid<{{.StructName}}Api.Row>({
{{if .SearchFields}}  formOptions: { schema: getSearchSchema(), submitOnChange: false, submitOnEnter: true },
{{end}}
  gridOptions: {
    columns,
    height: 'auto',
    pagerConfig: { pageSize: 20 },
    rowConfig: { keyField: 'id' },
    toolbarConfig: { refresh: true, custom: true, zoom: true },
    proxyConfig: {
      ajax: {
        query: async ({ page }: { page: { currentPage: number; pageSize: number } }{{if .SearchFields}}, values: Partial<{{.StructName}}Api.Query> = {}{{end}}) => get{{.StructName}}List({
          page: page.currentPage,
          pageSize: page.pageSize,
{{range .SearchFields}}          {{.Name}}: {{if eq .Type "string"}}values.{{.Name}} === '' ? undefined : {{end}}values.{{.Name}} ?? undefined,
{{end}}        }),
      },
    },
  },
});

function openForm(mode: 'create' | 'detail' | 'edit', row?: {{.StructName}}Api.Row) {
  modalApi.setData({ mode, id: row?.id }).open();
}

async function onDelete(row: {{.StructName}}Api.Row) {
  if (deleting.value.has(row.id)) return;
  deleting.value.add(row.id);
  try {
    await ElMessageBox.confirm('删除后无法恢复，确定删除这条记录吗？', '确认删除', {
      confirmButtonText: '删除', cancelButtonText: '取消', type: 'warning',
    });
    await delete{{.StructName}}(row.id);
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
