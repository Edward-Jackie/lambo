// 由 Lambo 生成的起始代码，可按业务修改；再次生成遇到差异会停止。
import { requestClient } from '#/api/request';

export namespace SystemUserApi {
  export type ID = number;

  export interface Row {
    roleIds: number[];
    roleNames: string;
    id: ID;
    // 用户名
    username: string;
    // 姓名
    realName: string;
    // 状态
    enabled: boolean;
    // 备注
    remark?: string;
  }

  export interface Payload {
    roleIds: number[];
    password?: string;
    username: string;
    realName: string;
    enabled: boolean;
    remark?: string;
  }

  export interface Query {
    page: number;
    pageSize: number;
    username?: string;
    realName?: string;
    enabled?: boolean;
  }

  export interface PageResult {
    items: Row[];
    total: number;
  }
}

const baseURL = '/api/system/users';

export function getSystemUserList(params: SystemUserApi.Query) {
  return requestClient.get<SystemUserApi.PageResult>(baseURL, { params });
}

export function getSystemUser(id: SystemUserApi.ID) {
  return requestClient.get<SystemUserApi.Row>(
    `${baseURL}/${encodeURIComponent(String(id))}`,
  );
}

export function createSystemUser(data: SystemUserApi.Payload) {
  return requestClient.post<SystemUserApi.Row>(baseURL, data);
}

export function updateSystemUser(
  id: SystemUserApi.ID,
  data: SystemUserApi.Payload,
) {
  return requestClient.put<SystemUserApi.Row>(
    `${baseURL}/${encodeURIComponent(String(id))}`,
    data,
  );
}

export function deleteSystemUser(id: SystemUserApi.ID) {
  return requestClient.delete(`${baseURL}/${encodeURIComponent(String(id))}`);
}
