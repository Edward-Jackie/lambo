// 由 Lambo 生成的起始路由；后台动态菜单接入后可引用同一组件路径。
import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    name: 'BusinessCategory',
    path: '/business/category',
    component: () => import('#/views/business/category/index.vue'),
    meta: { title: "分类管理", icon: 'lucide:folder', keepAlive: true },
  },
];

export default routes;
