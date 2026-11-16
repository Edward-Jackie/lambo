// 由 Lambo 生成的起始路由；后台动态菜单接入后可引用同一组件路径。
import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    name: 'Business{{.StructName}}',
    path: '/business/{{.Name}}',
    component: () => import('#/views/business/{{.Name}}/index.vue'),
    meta: { title: {{json .Title}}, icon: 'lucide:folder', keepAlive: true },
  },
];

export default routes;
