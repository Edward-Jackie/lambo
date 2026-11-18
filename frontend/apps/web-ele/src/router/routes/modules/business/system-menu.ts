// 由 Lambo 生成，仅登记组件路径与路由名称，不参与实际路由。
// Vben 以 accessMode: 'backend' 运行，侧边栏与路由完全来自 /api/menu/all 返回的数据库菜单；
// 调整菜单请改「系统管理 → 菜单管理」或 backend/internal/<domain>/seed/<名称>.sql。
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
