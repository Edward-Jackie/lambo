// 由 Lambo 生成，可按业务修改；再次生成遇到差异会停止。
package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Edward-Jackie/lambo/backend/internal/system/model"
	"github.com/Edward-Jackie/lambo/backend/internal/system/service"
	"github.com/Edward-Jackie/lambo/backend/pkg/helper/resp"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
)

const (
	SystemRoleAPIPath          = "/api/system/roles"
	SystemRolePermissionQuery  = "system:role:query"
	SystemRolePermissionCreate = "system:role:create"
	SystemRolePermissionUpdate = "system:role:update"
	SystemRolePermissionDelete = "system:role:delete"
	systemRoleMaxBodyBytes     = 1 << 20
)

type SystemRoleHandler struct {
	*Handler
	systemRoleService service.SystemRoleService
}

func NewSystemRoleHandler(handler *Handler, svc service.SystemRoleService) *SystemRoleHandler {
	return &SystemRoleHandler{Handler: handler, systemRoleService: svc}
}

// RegisterRoutes 显式接入应用权限判断；未配置时拒绝注册，避免产生公开写接口。
func (h *SystemRoleHandler) RegisterRoutes(router gin.IRouter, authorize func(*gin.Context, string) bool) error {
	if router == nil || authorize == nil {
		return errors.New("注册业务路由需要路由器和权限判断函数")
	}
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
	router.GET(SystemRoleAPIPath, require(SystemRolePermissionQuery), h.ListSystemRole)
	router.GET(SystemRoleAPIPath+"/:id", require(SystemRolePermissionQuery), h.GetSystemRole)
	router.POST(SystemRoleAPIPath, require(SystemRolePermissionCreate), h.CreateSystemRole)
	router.PUT(SystemRoleAPIPath+"/:id", require(SystemRolePermissionUpdate), h.UpdateSystemRole)
	router.DELETE(SystemRoleAPIPath+"/:id", require(SystemRolePermissionDelete), h.DeleteSystemRole)
	return nil
}

func (h *SystemRoleHandler) ListSystemRole(ctx *gin.Context) {
	query := model.SystemRoleQuery{Page: 1, PageSize: model.SystemRoleDefaultPageSize}
	if err := ctx.ShouldBindQuery(&query); err != nil {
		h.handleSystemRoleError(ctx, fmt.Errorf("%w: 查询参数格式错误", model.ErrSystemRoleInvalid))
		return
	}
	result, err := h.systemRoleService.ListSystemRole(ctx.Request.Context(), query)
	if err != nil {
		h.handleSystemRoleError(ctx, err)
		return
	}
	resp.HandleSuccess(ctx, result)
}

func (h *SystemRoleHandler) GetSystemRole(ctx *gin.Context) {
	id, err := h.systemRoleID(ctx)
	if err != nil {
		h.handleSystemRoleError(ctx, err)
		return
	}
	item, err := h.systemRoleService.GetSystemRole(ctx.Request.Context(), id)
	if err != nil {
		h.handleSystemRoleError(ctx, err)
		return
	}
	resp.HandleSuccess(ctx, item)
}

func (h *SystemRoleHandler) CreateSystemRole(ctx *gin.Context) {
	payload, err := h.decodeSystemRolePayload(ctx)
	if err != nil {
		h.handleSystemRoleError(ctx, err)
		return
	}
	item, err := h.systemRoleService.CreateSystemRole(ctx.Request.Context(), payload)
	if err != nil {
		h.handleSystemRoleError(ctx, err)
		return
	}
	h.logger.WithContext(ctx).Info("新增业务记录", zap.String("resource", "system-role"), zap.Uint64("actorID", model.ActorID(ctx.Request.Context())), zap.Any("id", item.ID))
	resp.HandleSuccess(ctx, item)
}

func (h *SystemRoleHandler) UpdateSystemRole(ctx *gin.Context) {
	id, err := h.systemRoleID(ctx)
	if err != nil {
		h.handleSystemRoleError(ctx, err)
		return
	}
	payload, err := h.decodeSystemRolePayload(ctx)
	if err != nil {
		h.handleSystemRoleError(ctx, err)
		return
	}
	item, err := h.systemRoleService.UpdateSystemRole(ctx.Request.Context(), id, payload)
	if err != nil {
		h.handleSystemRoleError(ctx, err)
		return
	}
	h.logger.WithContext(ctx).Info("编辑业务记录", zap.String("resource", "system-role"), zap.Uint64("actorID", model.ActorID(ctx.Request.Context())), zap.Any("id", id))
	resp.HandleSuccess(ctx, item)
}

func (h *SystemRoleHandler) DeleteSystemRole(ctx *gin.Context) {
	id, err := h.systemRoleID(ctx)
	if err != nil {
		h.handleSystemRoleError(ctx, err)
		return
	}
	if err := h.systemRoleService.DeleteSystemRole(ctx.Request.Context(), id); err != nil {
		h.handleSystemRoleError(ctx, err)
		return
	}
	h.logger.WithContext(ctx).Info("删除业务记录", zap.String("resource", "system-role"), zap.Uint64("actorID", model.ActorID(ctx.Request.Context())), zap.Any("id", id))
	resp.HandleSuccess(ctx, nil)
}

func (h *SystemRoleHandler) systemRoleID(ctx *gin.Context) (uint64, error) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: ID 无效", model.ErrSystemRoleInvalid)
	}
	return id, model.ValidateSystemRoleID(id)
}

func (h *SystemRoleHandler) decodeSystemRolePayload(ctx *gin.Context) (model.SystemRolePayload, error) {
	var fields map[string]json.RawMessage
	invalid := fmt.Errorf("%w: 请求体必须是合法的 JSON 对象，且只能包含可编辑字段", model.ErrSystemRoleInvalid)
	if ctx.ContentType() != "application/json" {
		return model.SystemRolePayload{}, invalid
	}
	decoder := json.NewDecoder(http.MaxBytesReader(ctx.Writer, ctx.Request.Body, systemRoleMaxBodyBytes))
	if err := decoder.Decode(&fields); err != nil {
		var limit *http.MaxBytesError
		if errors.As(err, &limit) {
			return model.SystemRolePayload{}, err
		}
		return model.SystemRolePayload{}, invalid
	}
	if fields == nil {
		return model.SystemRolePayload{}, invalid
	}
	var extra interface{}
	if err := decoder.Decode(&extra); err != io.EOF {
		var limit *http.MaxBytesError
		if errors.As(err, &limit) {
			return model.SystemRolePayload{}, err
		}
		return model.SystemRolePayload{}, invalid
	}
	var payload model.SystemRolePayload
	for name, value := range fields {
		switch name {
		case "code":
			if err := json.Unmarshal(value, &payload.Code); err != nil {
				return model.SystemRolePayload{}, invalid
			}
			if payload.Code == nil {
				return model.SystemRolePayload{}, invalid
			}
		case "name":
			if err := json.Unmarshal(value, &payload.Name); err != nil {
				return model.SystemRolePayload{}, invalid
			}
			if payload.Name == nil {
				return model.SystemRolePayload{}, invalid
			}
		case "enabled":
			if err := json.Unmarshal(value, &payload.Enabled); err != nil {
				return model.SystemRolePayload{}, invalid
			}
			if payload.Enabled == nil {
				return model.SystemRolePayload{}, invalid
			}
		case "remark":
			if err := json.Unmarshal(value, &payload.Remark); err != nil {
				return model.SystemRolePayload{}, invalid
			}

		case "menuIds":
			if err := json.Unmarshal(value, &payload.MenuIDs); err != nil || payload.MenuIDs == nil {
				return model.SystemRolePayload{}, invalid
			}
		default:
			return model.SystemRolePayload{}, invalid
		}
	}
	return payload, nil
}

func (h *SystemRoleHandler) handleSystemRoleError(ctx *gin.Context, err error) {
	var limit *http.MaxBytesError
	if h.handleIAMError(ctx, err) {
		return
	}
	switch {
	case errors.As(err, &limit):
		resp.HandleError(ctx, http.StatusRequestEntityTooLarge, 1, "请求体过大", nil)
	case errors.Is(err, model.ErrSystemRoleInvalid):
		resp.HandleError(ctx, http.StatusBadRequest, 1, err.Error(), nil)
	case errors.Is(err, model.ErrSystemRoleNotFound):
		resp.HandleError(ctx, http.StatusNotFound, 1, "记录不存在", nil)
	default:
		h.logger.WithContext(ctx).Error("业务接口失败", zap.String("resource", "system-role"), zap.String("method", ctx.Request.Method), zap.Error(err))
		resp.HandleError(ctx, http.StatusInternalServerError, 1, "操作失败，请稍后重试", nil)
	}
}
