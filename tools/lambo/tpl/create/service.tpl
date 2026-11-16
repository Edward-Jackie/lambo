// 由 Lambo 生成的起始代码，可按业务修改；再次生成遇到差异会停止。
package service

import (
 "context"
 "{{ .Module }}/internal/model"
 "{{ .Module }}/internal/repository"
)

type {{ .StructName }}Service interface {
 Get{{ .StructName }}(ctx context.Context, id {{if eq .IDType "string"}}string{{else}}uint64{{end}}) (*model.{{ .StructName }}, error)
}

type {{ .LowerName }}Service struct {
 *Service
 {{ .LowerName }}Repository repository.{{ .StructName }}Repository
}

func New{{ .StructName }}Service(service *Service, repo repository.{{ .StructName }}Repository) {{ .StructName }}Service {
 return &{{ .LowerName }}Service{Service: service, {{ .LowerName }}Repository: repo}
}

func (s *{{ .LowerName }}Service) Get{{ .StructName }}(ctx context.Context, id {{if eq .IDType "string"}}string{{else}}uint64{{end}}) (*model.{{ .StructName }}, error) {
 return s.{{ .LowerName }}Repository.Get{{ .StructName }}(ctx, id)
}
