// 由 Lambo 生成，可按业务修改；再次生成遇到差异会停止。
package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/Edward-Jackie/lambo/backend/internal/model"
	"github.com/Edward-Jackie/lambo/backend/internal/service"
	"github.com/Edward-Jackie/lambo/backend/pkg/helper/resp"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"io"
	"net/http"
	"strconv"
)

const (
	SystemMenuAPIPath          = "/api/system/menus"
	SystemMenuPermissionQuery  = "system:menu:query"
	SystemMenuPermissionCreate = "system:menu:create"
	SystemMenuPermissionUpdate = "system:menu:update"
	SystemMenuPermissionDelete = "system:menu:delete"
	systemMenuMaxBodyBytes     = 1 << 20
)

type SystemMenuHandler struct {
	*Handler
	systemMenuService service.SystemMenuService
}

func NewSystemMenuHandler(handler *Handler, svc service.SystemMenuService) *SystemMenuHandler {
	return &SystemMenuHandler{Handler: handler, systemMenuService: svc}
}

// RegisterRoutes 显式接入应用权限判断；未配置时拒绝注册，避免产生公开写接口。
func (h *SystemMenuHandler) RegisterRoutes(router gin.IRouter, authorize func(*gin.Context, string) bool) error {
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
	router.GET(SystemMenuAPIPath, require(SystemMenuPermissionQuery), h.ListSystemMenu)
	router.GET(SystemMenuAPIPath+"/:id", require(SystemMenuPermissionQuery), h.GetSystemMenu)
	router.POST(SystemMenuAPIPath, require(SystemMenuPermissionCreate), h.CreateSystemMenu)
	router.PUT(SystemMenuAPIPath+"/:id", require(SystemMenuPermissionUpdate), h.UpdateSystemMenu)
	router.DELETE(SystemMenuAPIPath+"/:id", require(SystemMenuPermissionDelete), h.DeleteSystemMenu)
	return nil
}

func (h *SystemMenuHandler) ListSystemMenu(ctx *gin.Context) {
	query := model.SystemMenuQuery{Page: 1, PageSize: model.SystemMenuDefaultPageSize}
	if err := ctx.ShouldBindQuery(&query); err != nil {
		h.handleSystemMenuError(ctx, fmt.Errorf("%w: 查询参数格式错误", model.ErrSystemMenuInvalid))
		return
	}
	result, err := h.systemMenuService.ListSystemMenu(ctx.Request.Context(), query)
	if err != nil {
		h.handleSystemMenuError(ctx, err)
		return
	}
	resp.HandleSuccess(ctx, result)
}

func (h *SystemMenuHandler) GetSystemMenu(ctx *gin.Context) {
	id, err := h.systemMenuID(ctx)
	if err != nil {
		h.handleSystemMenuError(ctx, err)
		return
	}
	item, err := h.systemMenuService.GetSystemMenu(ctx.Request.Context(), id)
	if err != nil {
		h.handleSystemMenuError(ctx, err)
		return
	}
	resp.HandleSuccess(ctx, item)
}

func (h *SystemMenuHandler) CreateSystemMenu(ctx *gin.Context) {
	payload, err := h.decodeSystemMenuPayload(ctx)
	if err != nil {
		h.handleSystemMenuError(ctx, err)
		return
	}
	item, err := h.systemMenuService.CreateSystemMenu(ctx.Request.Context(), payload)
	if err != nil {
		h.handleSystemMenuError(ctx, err)
		return
	}
	h.logger.WithContext(ctx).Info("新增业务记录", zap.String("resource", "system-menu"), zap.Uint64("actorID", model.ActorID(ctx.Request.Context())), zap.Any("id", item.ID))
	resp.HandleSuccess(ctx, item)
}

func (h *SystemMenuHandler) UpdateSystemMenu(ctx *gin.Context) {
	id, err := h.systemMenuID(ctx)
	if err != nil {
		h.handleSystemMenuError(ctx, err)
		return
	}
	payload, err := h.decodeSystemMenuPayload(ctx)
	if err != nil {
		h.handleSystemMenuError(ctx, err)
		return
	}
	item, err := h.systemMenuService.UpdateSystemMenu(ctx.Request.Context(), id, payload)
	if err != nil {
		h.handleSystemMenuError(ctx, err)
		return
	}
	h.logger.WithContext(ctx).Info("编辑业务记录", zap.String("resource", "system-menu"), zap.Uint64("actorID", model.ActorID(ctx.Request.Context())), zap.Any("id", id))
	resp.HandleSuccess(ctx, item)
}

func (h *SystemMenuHandler) DeleteSystemMenu(ctx *gin.Context) {
	id, err := h.systemMenuID(ctx)
	if err != nil {
		h.handleSystemMenuError(ctx, err)
		return
	}
	if err := h.systemMenuService.DeleteSystemMenu(ctx.Request.Context(), id); err != nil {
		h.handleSystemMenuError(ctx, err)
		return
	}
	h.logger.WithContext(ctx).Info("删除业务记录", zap.String("resource", "system-menu"), zap.Uint64("actorID", model.ActorID(ctx.Request.Context())), zap.Any("id", id))
	resp.HandleSuccess(ctx, nil)
}

func (h *SystemMenuHandler) systemMenuID(ctx *gin.Context) (uint64, error) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		return 0, fmt.Errorf("%w: ID 无效", model.ErrSystemMenuInvalid)
	}
	return id, model.ValidateSystemMenuID(id)
}

func (h *SystemMenuHandler) decodeSystemMenuPayload(ctx *gin.Context) (model.SystemMenuPayload, error) {
	var fields map[string]json.RawMessage
	invalid := fmt.Errorf("%w: 请求体必须是合法的 JSON 对象，且只能包含可编辑字段", model.ErrSystemMenuInvalid)
	if ctx.ContentType() != "application/json" {
		return model.SystemMenuPayload{}, invalid
	}
	decoder := json.NewDecoder(http.MaxBytesReader(ctx.Writer, ctx.Request.Body, systemMenuMaxBodyBytes))
	if err := decoder.Decode(&fields); err != nil {
		var limit *http.MaxBytesError
		if errors.As(err, &limit) {
			return model.SystemMenuPayload{}, err
		}
		return model.SystemMenuPayload{}, invalid
	}
	if fields == nil {
		return model.SystemMenuPayload{}, invalid
	}
	var extra interface{}
	if err := decoder.Decode(&extra); err != io.EOF {
		var limit *http.MaxBytesError
		if errors.As(err, &limit) {
			return model.SystemMenuPayload{}, err
		}
		return model.SystemMenuPayload{}, invalid
	}
	var payload model.SystemMenuPayload
	for name, value := range fields {
		switch name {
		case "name":
			if err := json.Unmarshal(value, &payload.Name); err != nil {
				return model.SystemMenuPayload{}, invalid
			}
			if payload.Name == nil {
				return model.SystemMenuPayload{}, invalid
			}
		case "kind":
			if err := json.Unmarshal(value, &payload.Kind); err != nil {
				return model.SystemMenuPayload{}, invalid
			}
			if payload.Kind == nil {
				return model.SystemMenuPayload{}, invalid
			}
		case "parentId":
			if err := json.Unmarshal(value, &payload.ParentId); err != nil {
				return model.SystemMenuPayload{}, invalid
			}
			if payload.ParentId == nil {
				return model.SystemMenuPayload{}, invalid
			}
		case "routeName":
			if err := json.Unmarshal(value, &payload.RouteName); err != nil {
				return model.SystemMenuPayload{}, invalid
			}

		case "path":
			if err := json.Unmarshal(value, &payload.Path); err != nil {
				return model.SystemMenuPayload{}, invalid
			}

		case "component":
			if err := json.Unmarshal(value, &payload.Component); err != nil {
				return model.SystemMenuPayload{}, invalid
			}

		case "permission":
			if err := json.Unmarshal(value, &payload.Permission); err != nil {
				return model.SystemMenuPayload{}, invalid
			}

		case "icon":
			if err := json.Unmarshal(value, &payload.Icon); err != nil {
				return model.SystemMenuPayload{}, invalid
			}

		case "sort":
			if err := json.Unmarshal(value, &payload.Sort); err != nil {
				return model.SystemMenuPayload{}, invalid
			}
			if payload.Sort == nil {
				return model.SystemMenuPayload{}, invalid
			}
		case "enabled":
			if err := json.Unmarshal(value, &payload.Enabled); err != nil {
				return model.SystemMenuPayload{}, invalid
			}
			if payload.Enabled == nil {
				return model.SystemMenuPayload{}, invalid
			}
		default:
			return model.SystemMenuPayload{}, invalid
		}
	}
	return payload, nil
}

func (h *SystemMenuHandler) handleSystemMenuError(ctx *gin.Context, err error) {
	var limit *http.MaxBytesError
	if h.handleIAMError(ctx, err) {
		return
	}
	switch {
	case errors.As(err, &limit):
		resp.HandleError(ctx, http.StatusRequestEntityTooLarge, 1, "请求体过大", nil)
	case errors.Is(err, model.ErrSystemMenuInvalid):
		resp.HandleError(ctx, http.StatusBadRequest, 1, err.Error(), nil)
	case errors.Is(err, model.ErrSystemMenuNotFound):
		resp.HandleError(ctx, http.StatusNotFound, 1, "记录不存在", nil)
	default:
		h.logger.WithContext(ctx).Error("业务接口失败", zap.String("resource", "system-menu"), zap.String("method", ctx.Request.Method), zap.Error(err))
		resp.HandleError(ctx, http.StatusInternalServerError, 1, "操作失败，请稍后重试", nil)
	}
}
