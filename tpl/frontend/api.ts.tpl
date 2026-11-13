// 由 Lambo 生成的起始代码，可按业务修改；再次生成遇到差异会停止。
import { requestClient } from '#/api/request';

export namespace {{.StructName}}Api {
  export type ID = {{.IDType}};

  export interface Row {
    id: ID;
{{range .Fields}}    // {{.Comment}}
    {{.Name}}{{if not .Required}}?{{end}}: {{.Type}};
{{end}}  }

  export interface Payload {
{{range .EditableFields}}    {{.Name}}{{if not .Required}}?{{end}}: {{.Type}};
{{end}}  }

  export interface Query {
    page: number;
    pageSize: number;
{{range .SearchFields}}    {{.Name}}?: {{.Type}};
{{end}}  }

  export interface PageResult {
    items: Row[];
    total: number;
  }
}

const baseURL = {{json .API}};

export function get{{.StructName}}List(params: {{.StructName}}Api.Query) {
  return requestClient.get<{{.StructName}}Api.PageResult>(baseURL, { params });
}

export function get{{.StructName}}(id: {{.StructName}}Api.ID) {
  return requestClient.get<{{.StructName}}Api.Row>(`${baseURL}/${encodeURIComponent(String(id))}`);
}

export function create{{.StructName}}(data: {{.StructName}}Api.Payload) {
  return requestClient.post<{{.StructName}}Api.Row>(baseURL, data);
}

export function update{{.StructName}}(id: {{.StructName}}Api.ID, data: {{.StructName}}Api.Payload) {
  return requestClient.put<{{.StructName}}Api.Row>(`${baseURL}/${encodeURIComponent(String(id))}`, data);
}

export function delete{{.StructName}}(id: {{.StructName}}Api.ID) {
  return requestClient.delete(`${baseURL}/${encodeURIComponent(String(id))}`);
}
