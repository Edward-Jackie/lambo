// 由 Lambo 生成，可按业务修改；再次生成遇到差异会停止。
package service

import (
 "context"
 {{if eq .IDType "string"}}"github.com/google/uuid"{{end}}
 "{{ .Module }}/internal/{{.Domain}}/model"
 "{{ .Module }}/internal/{{.Domain}}/repository"
)

type {{ .StructName }}Service interface {
 List{{.StructName}}(ctx context.Context, query model.{{.StructName}}Query) (*model.{{.StructName}}Page, error)
 Get{{.StructName}}(ctx context.Context, id {{.IDGoType}}) (*model.{{.StructName}}, error)
 Create{{.StructName}}(ctx context.Context, payload model.{{.StructName}}Payload) (*model.{{.StructName}}, error)
 Update{{.StructName}}(ctx context.Context, id {{.IDGoType}}, payload model.{{.StructName}}Payload) (*model.{{.StructName}}, error)
 Delete{{.StructName}}(ctx context.Context, id {{.IDGoType}}) error
}

type {{ .LowerName }}Service struct {
 *Service
 {{ .LowerName }}Repository repository.{{ .StructName }}Repository
}

func New{{ .StructName }}Service(service *Service, repo repository.{{ .StructName }}Repository) {{ .StructName }}Service {
 return &{{ .LowerName }}Service{Service: service, {{ .LowerName }}Repository: repo}
}

func (s *{{.LowerName}}Service) List{{.StructName}}(ctx context.Context, query model.{{.StructName}}Query) (*model.{{.StructName}}Page, error) {
 if err := query.Validate(); err != nil { return nil, err }
 return s.{{.LowerName}}Repository.List{{.StructName}}(ctx, query)
}

func (s *{{ .LowerName }}Service) Get{{ .StructName }}(ctx context.Context, id {{.IDGoType}}) (*model.{{ .StructName }}, error) {
 if err := model.Validate{{.StructName}}ID(id); err != nil { return nil, err }
 return s.{{ .LowerName }}Repository.Get{{ .StructName }}(ctx, id)
}

func (s *{{.LowerName}}Service) Create{{.StructName}}(ctx context.Context, payload model.{{.StructName}}Payload) (*model.{{.StructName}}, error) {
 payload.ApplyDefaults()
 if err := payload.Validate(); err != nil { return nil, err }
 item := &model.{{.StructName}}{
 {{if eq .IDType "string"}}ID: uuid.NewString(),{{end}}
{{range .EditableFields}} {{.GoName}}: {{if .Required}}*{{end}}payload.{{.GoName}},
{{end}} }
{{range .Fields}}{{if and .ReadOnly .HasDefault}} {
  value := {{if eq .Type "number"}}float64({{json .Default}}){{else}}{{json .Default}}{{end}}
  item.{{.GoName}} = {{if not .Required}}&{{end}}value
 }
{{end}}{{end}} if err := s.{{.LowerName}}Repository.Create{{.StructName}}(ctx, item); err != nil { return nil, err }
 return item, nil
}

func (s *{{.LowerName}}Service) Update{{.StructName}}(ctx context.Context, id {{.IDGoType}}, payload model.{{.StructName}}Payload) (*model.{{.StructName}}, error) {
 if err := model.Validate{{.StructName}}ID(id); err != nil { return nil, err }
 if err := payload.Validate(); err != nil { return nil, err }
 return s.{{.LowerName}}Repository.Update{{.StructName}}(ctx, id, payload)
}

func (s *{{.LowerName}}Service) Delete{{.StructName}}(ctx context.Context, id {{.IDGoType}}) error {
 if err := model.Validate{{.StructName}}ID(id); err != nil { return err }
 return s.{{.LowerName}}Repository.Delete{{.StructName}}(ctx, id)
}
