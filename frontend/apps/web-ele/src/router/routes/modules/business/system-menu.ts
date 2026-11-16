// 由 Lambo 生成的起始路由；后台动态菜单接入后可引用同一组件路径。
import type { RouteRecordRaw } from 'vue-router';

const routes: RouteRecordRaw[] = [
  {
    name: 'BusinessSystemMenu',
    path: '/business/system-menu',
    component: () => import('#/views/business/system-menu/index.vue'),
    meta: { title: '菜单管理', icon: 'lucide:folder', keepAlive: true },
  },
];

export default routes;
