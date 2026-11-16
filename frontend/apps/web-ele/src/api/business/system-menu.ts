// 由 Lambo 生成的起始代码，可按业务修改；再次生成遇到差异会停止。
import { requestClient } from '#/api/request';

export namespace SystemMenuApi {
  export type ID = number;

  export interface Row {
    id: ID;
    // 菜单名称
    name: string;
    // 类型
    kind: string;
    // 上级菜单
    parentId: number;
    // 路由名称
    routeName?: string;
    // 路由路径
    path?: string;
    // 页面组件
    component?: string;
    // 权限码
    permission?: string;
    // 图标
    icon?: string;
    // 排序
    sort: number;
    // 状态
    enabled: boolean;
  }

  export interface Payload {
    name: string;
    kind: string;
    parentId: number;
    routeName?: string;
    path?: string;
    component?: string;
    permission?: string;
    icon?: string;
    sort: number;
    enabled: boolean;
  }

  export interface Query {
    page: number;
    pageSize: number;
    name?: string;
    kind?: string;
    enabled?: boolean;
  }

  export interface PageResult {
    items: Row[];
    total: number;
  }
}

const baseURL = '/api/system/menus';

export function getSystemMenuList(params: SystemMenuApi.Query) {
  return requestClient.get<SystemMenuApi.PageResult>(baseURL, { params });
}

export function getSystemMenu(id: SystemMenuApi.ID) {
  return requestClient.get<SystemMenuApi.Row>(
    `${baseURL}/${encodeURIComponent(String(id))}`,
  );
}

export function createSystemMenu(data: SystemMenuApi.Payload) {
  return requestClient.post<SystemMenuApi.Row>(baseURL, data);
}

export function updateSystemMenu(
  id: SystemMenuApi.ID,
  data: SystemMenuApi.Payload,
) {
  return requestClient.put<SystemMenuApi.Row>(
    `${baseURL}/${encodeURIComponent(String(id))}`,
    data,
  );
}

export function deleteSystemMenu(id: SystemMenuApi.ID) {
  return requestClient.delete(`${baseURL}/${encodeURIComponent(String(id))}`);
}
