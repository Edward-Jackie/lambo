// 由 Lambo 生成，可按业务修改；再次生成遇到差异会停止。
package handler

import (
 "encoding/json"
 "errors"
 "fmt"
 "io"
 "net/http"
 {{if eq .IDType "number"}}"strconv"{{end}}
 "github.com/gin-gonic/gin"
 "go.uber.org/zap"
 "{{.Module}}/internal/model"
 "{{ .Module }}/internal/service"
 "{{ .Module }}/pkg/helper/resp"
)

const (
 {{.StructName}}APIPath = {{json .API}}
 {{.StructName}}PermissionQuery = {{json (printf "%s:query" .Permission)}}
 {{.StructName}}PermissionCreate = {{json (printf "%s:create" .Permission)}}
 {{.StructName}}PermissionUpdate = {{json (printf "%s:update" .Permission)}}
 {{.StructName}}PermissionDelete = {{json (printf "%s:delete" .Permission)}}
 {{.LowerName}}MaxBodyBytes = 1 << 20
)

type {{ .StructName }}Handler struct {
 *Handler
 {{ .LowerName }}Service service.{{ .StructName }}Service
}

func New{{ .StructName }}Handler(handler *Handler, svc service.{{ .StructName }}Service) *{{ .StructName }}Handler {
 return &{{ .StructName }}Handler{Handler: handler, {{ .LowerName }}Service: svc}
}

// RegisterRoutes 显式接入应用权限判断；未配置时拒绝注册，避免产生公开写接口。
func (h *{{.StructName}}Handler) RegisterRoutes(router gin.IRouter, authorize func(*gin.Context, string) bool) error {
 if router == nil || authorize == nil { return errors.New("注册业务路由需要路由器和权限判断函数") }
 require := func(permission string) gin.HandlerFunc {
  return func(ctx *gin.Context) {
   if !authorize(ctx, permission) {
    ctx.Abort()
    resp.HandleError(ctx, http.StatusForbidden, 1, "无操作权限", nil)
    return
   }
   ctx.Next()
  }
 }
 router.GET({{.StructName}}APIPath, require({{.StructName}}PermissionQuery), h.List{{.StructName}})
 router.GET({{.StructName}}APIPath+"/:id", require({{.StructName}}PermissionQuery), h.Get{{.StructName}})
 router.POST({{.StructName}}APIPath, require({{.StructName}}PermissionCreate), h.Create{{.StructName}})
 router.PUT({{.StructName}}APIPath+"/:id", require({{.StructName}}PermissionUpdate), h.Update{{.StructName}})
 router.DELETE({{.StructName}}APIPath+"/:id", require({{.StructName}}PermissionDelete), h.Delete{{.StructName}})
 return nil
}

func (h *{{.StructName}}Handler) List{{.StructName}}(ctx *gin.Context) {
 query := model.{{.StructName}}Query{Page: 1, PageSize: model.{{.StructName}}DefaultPageSize}
 if err := ctx.ShouldBindQuery(&query); err != nil {
  h.handle{{.StructName}}Error(ctx, fmt.Errorf("%w: 查询参数格式错误", model.Err{{.StructName}}Invalid))
  return
 }
 result, err := h.{{.LowerName}}Service.List{{.StructName}}(ctx.Request.Context(), query)
 if err != nil { h.handle{{.StructName}}Error(ctx, err); return }
 resp.HandleSuccess(ctx, result)
}

func (h *{{.StructName}}Handler) Get{{.StructName}}(ctx *gin.Context) {
 id, err := h.{{.LowerName}}ID(ctx)
 if err != nil { h.handle{{.StructName}}Error(ctx, err); return }
 item, err := h.{{.LowerName}}Service.Get{{.StructName}}(ctx.Request.Context(), id)
 if err != nil { h.handle{{.StructName}}Error(ctx, err); return }
 resp.HandleSuccess(ctx, item)
}

func (h *{{.StructName}}Handler) Create{{.StructName}}(ctx *gin.Context) {
 payload, err := h.decode{{.StructName}}Payload(ctx)
 if err != nil { h.handle{{.StructName}}Error(ctx, err); return }
 item, err := h.{{.LowerName}}Service.Create{{.StructName}}(ctx.Request.Context(), payload)
 if err != nil { h.handle{{.StructName}}Error(ctx, err); return }
 h.logger.WithContext(ctx).Info("新增业务记录", zap.String("resource", {{json .Name}}), zap.Any("id", item.ID))
 resp.HandleSuccess(ctx, item)
}

func (h *{{.StructName}}Handler) Update{{.StructName}}(ctx *gin.Context) {
 id, err := h.{{.LowerName}}ID(ctx)
 if err != nil { h.handle{{.StructName}}Error(ctx, err); return }
 payload, err := h.decode{{.StructName}}Payload(ctx)
 if err != nil { h.handle{{.StructName}}Error(ctx, err); return }
 item, err := h.{{.LowerName}}Service.Update{{.StructName}}(ctx.Request.Context(), id, payload)
 if err != nil { h.handle{{.StructName}}Error(ctx, err); return }
 h.logger.WithContext(ctx).Info("编辑业务记录", zap.String("resource", {{json .Name}}), zap.Any("id", id))
 resp.HandleSuccess(ctx, item)
}

func (h *{{.StructName}}Handler) Delete{{.StructName}}(ctx *gin.Context) {
 id, err := h.{{.LowerName}}ID(ctx)
 if err != nil { h.handle{{.StructName}}Error(ctx, err); return }
 if err := h.{{.LowerName}}Service.Delete{{.StructName}}(ctx.Request.Context(), id); err != nil { h.handle{{.StructName}}Error(ctx, err); return }
 h.logger.WithContext(ctx).Info("删除业务记录", zap.String("resource", {{json .Name}}), zap.Any("id", id))
 resp.HandleSuccess(ctx, nil)
}

func (h *{{.StructName}}Handler) {{.LowerName}}ID(ctx *gin.Context) ({{.IDGoType}}, error) {
 {{if eq .IDType "string"}}id := ctx.Param("id"){{else}}id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
 if err != nil { return 0, fmt.Errorf("%w: ID 无效", model.Err{{.StructName}}Invalid) }{{end}}
 return id, model.Validate{{.StructName}}ID(id)
}

func (h *{{.StructName}}Handler) decode{{.StructName}}Payload(ctx *gin.Context) (model.{{.StructName}}Payload, error) {
 var fields map[string]json.RawMessage
 invalid := fmt.Errorf("%w: 请求体必须是合法的 JSON 对象，且只能包含可编辑字段", model.Err{{.StructName}}Invalid)
 if ctx.ContentType() != "application/json" { return model.{{.StructName}}Payload{}, invalid }
 decoder := json.NewDecoder(http.MaxBytesReader(ctx.Writer, ctx.Request.Body, {{.LowerName}}MaxBodyBytes))
 if err := decoder.Decode(&fields); err != nil {
  var limit *http.MaxBytesError
  if errors.As(err, &limit) { return model.{{.StructName}}Payload{}, err }
  return model.{{.StructName}}Payload{}, invalid
 }
 if fields == nil { return model.{{.StructName}}Payload{}, invalid }
 var extra interface{}
 if err := decoder.Decode(&extra); err != io.EOF {
  var limit *http.MaxBytesError
  if errors.As(err, &limit) { return model.{{.StructName}}Payload{}, err }
  return model.{{.StructName}}Payload{}, invalid
 }
 var payload model.{{.StructName}}Payload
 for name, value := range fields {
  switch name {
{{range .EditableFields}}  case {{json .Name}}:
   if err := json.Unmarshal(value, &payload.{{.GoName}}); err != nil { return model.{{$.StructName}}Payload{}, invalid }
   {{if .Required}}if payload.{{.GoName}} == nil { return model.{{$.StructName}}Payload{}, invalid }{{end}}
{{end}}  default: return model.{{.StructName}}Payload{}, invalid
  }
 }
 return payload, nil
}

func (h *{{.StructName}}Handler) handle{{.StructName}}Error(ctx *gin.Context, err error) {
 var limit *http.MaxBytesError
 switch {
 case errors.As(err, &limit): resp.HandleError(ctx, http.StatusRequestEntityTooLarge, 1, "请求体过大", nil)
 case errors.Is(err, model.Err{{.StructName}}Invalid): resp.HandleError(ctx, http.StatusBadRequest, 1, err.Error(), nil)
 case errors.Is(err, model.Err{{.StructName}}NotFound): resp.HandleError(ctx, http.StatusNotFound, 1, "记录不存在", nil)
 default:
  h.logger.WithContext(ctx).Error("业务接口失败", zap.String("resource", {{json .Name}}), zap.String("method", ctx.Request.Method), zap.Error(err))
  resp.HandleError(ctx, http.StatusInternalServerError, 1, "操作失败，请稍后重试", nil)
 }
}
