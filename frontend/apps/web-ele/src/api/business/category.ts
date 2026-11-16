// 由 Lambo 生成的起始代码，可按业务修改；再次生成遇到差异会停止。
import { requestClient } from '#/api/request';

export namespace CategoryApi {
  export type ID = number;

  export interface Row {
    id: ID;
    // 分类名称
    name: string;
    // 状态
    status: number;
    // 排序
    sort: number;
    // 备注
    remark?: string;
  }

  export interface Payload {
    name: string;
    status: number;
    sort: number;
    remark?: string;
  }

  export interface Query {
    page: number;
    pageSize: number;
    name?: string;
    status?: number;
  }

  export interface PageResult {
    items: Row[];
    total: number;
  }
}

const baseURL = "/api/categories";

export function getCategoryList(params: CategoryApi.Query) {
  return requestClient.get<CategoryApi.PageResult>(baseURL, { params });
}

export function getCategory(id: CategoryApi.ID) {
  return requestClient.get<CategoryApi.Row>(`${baseURL}/${encodeURIComponent(String(id))}`);
}

export function createCategory(data: CategoryApi.Payload) {
  return requestClient.post<CategoryApi.Row>(baseURL, data);
}

export function updateCategory(id: CategoryApi.ID, data: CategoryApi.Payload) {
  return requestClient.put<CategoryApi.Row>(`${baseURL}/${encodeURIComponent(String(id))}`, data);
}

export function deleteCategory(id: CategoryApi.ID) {
  return requestClient.delete(`${baseURL}/${encodeURIComponent(String(id))}`);
}
