import type { SystemMenuApi } from './system-menu';
import type { SystemRoleApi } from './system-role';
import { requestClient } from '#/api/request';

export interface ManagementOptions {
  roles: SystemRoleApi.Row[];
  grantedCodes: string[];
  menus: SystemMenuApi.Row[];
  permissions: Record<string, string>;
  components: Record<string, string>;
}
export function getManagementOptions() {
  return requestClient.get<ManagementOptions>('/api/system/options');
}
export interface MenuNode extends SystemMenuApi.Row {
  children?: MenuNode[];
}
export function menuTree(items: SystemMenuApi.Row[], parentId = 0): MenuNode[] {
  return items
    .filter((item) => item.parentId === parentId)
    .sort((a, b) => a.sort - b.sort || a.id - b.id)
    .map((item) => ({ ...item, children: menuTree(items, item.id) }));
}
export function resetUserPassword(id: number, password: string) {
  return requestClient.post(`/api/system/users/${id}/password`, { password });
}
