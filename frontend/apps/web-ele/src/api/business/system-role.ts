// 由 Lambo 生成的起始代码，可按业务修改；再次生成遇到差异会停止。
import { requestClient } from '#/api/request';

export namespace SystemRoleApi {
  export type ID = number;

  export interface Row {
    menuIds: number[];
    id: ID;
    // 角色编码
    code: string;
    // 角色名称
    name: string;
    // 状态
    enabled: boolean;
    // 备注
    remark?: string;
  }

  export interface Payload {
    menuIds: number[];
    code: string;
    name: string;
    enabled: boolean;
    remark?: string;
  }

  export interface Query {
    page: number;
    pageSize: number;
    code?: string;
    name?: string;
    enabled?: boolean;
  }

  export interface PageResult {
    items: Row[];
    total: number;
  }
}

const baseURL = '/api/system/roles';

export function getSystemRoleList(params: SystemRoleApi.Query) {
  return requestClient.get<SystemRoleApi.PageResult>(baseURL, { params });
}

export function getSystemRole(id: SystemRoleApi.ID) {
  return requestClient.get<SystemRoleApi.Row>(
    `${baseURL}/${encodeURIComponent(String(id))}`,
  );
}

export function createSystemRole(data: SystemRoleApi.Payload) {
  return requestClient.post<SystemRoleApi.Row>(baseURL, data);
}

export function updateSystemRole(
  id: SystemRoleApi.ID,
  data: SystemRoleApi.Payload,
) {
  return requestClient.put<SystemRoleApi.Row>(
    `${baseURL}/${encodeURIComponent(String(id))}`,
    data,
  );
}

export function deleteSystemRole(id: SystemRoleApi.ID) {
  return requestClient.delete(`${baseURL}/${encodeURIComponent(String(id))}`);
}
