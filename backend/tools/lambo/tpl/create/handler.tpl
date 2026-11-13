// 由 Lambo 生成的起始代码，可按业务修改；再次生成遇到差异会停止。
package handler

import (
 "net/http"
 "github.com/gin-gonic/gin"
 "{{ .Module }}/internal/service"
 "{{ .Module }}/pkg/helper/resp"
)

type {{ .StructName }}Handler struct {
 *Handler
 {{ .LowerName }}Service service.{{ .StructName }}Service
}

func New{{ .StructName }}Handler(handler *Handler, svc service.{{ .StructName }}Service) *{{ .StructName }}Handler {
 return &{{ .StructName }}Handler{Handler: handler, {{ .LowerName }}Service: svc}
}

func (h *{{ .StructName }}Handler) Get{{ .StructName }}(ctx *gin.Context) {
 // 先补齐参数校验、鉴权和业务调用，再接入路由。
 resp.HandleError(ctx, http.StatusNotImplemented, 1, "{{ .StructName }} 接口尚未实现", nil)
}
